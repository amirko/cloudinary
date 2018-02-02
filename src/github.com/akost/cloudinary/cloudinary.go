package main

import (
	"fmt"
	"bytes"
	"image"
	"github.com/nfnt/resize"
	"image/jpeg"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	log "log"
	"strconv"
	"image/color"
	"image/draw"
	"os"
)
type ThumbnailError struct{
	StatusOK bool
	Reason string
}

func main() {
	fmt.Printf("================ Thumbnail service is up and running!\n")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/thumbnail", thumbnail)
	log.Fatal(http.ListenAndServe(getPort(), router))
}
func getPort() string {
	var port = os.Getenv("PORT")
	// Set a default port if there is nothing in the environment
	if port == "" {
		port = "8080"
		log.Print("INFO: No PORT environment variable detected, defaulting to " + port)
	}
	return ":" + port
}



func thumbnail(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	url := params.Get("url")
	x, err1 := strconv.ParseUint(params.Get("x"), 10, 32)
	if err1 != nil {
		log.Print("Couldn't parse x param ", err1)
		writeError(w, "Couldn't parse x param", http.StatusBadRequest)
		return
	}
	y, err2 := strconv.ParseUint(params.Get("y"), 10, 32)
	if err2 != nil {
		log.Print("Couldn't parse y param ", err2)
		writeError(w, "Couldn't parse y param", http.StatusBadRequest)
		return
	}
	response, e := http.Get(url)
	if e != nil {
		log.Print("Could not read image ", e)
		writeError(w, "Not Found", http.StatusNotFound)
		return
	}
	im, _, e := image.Decode(response.Body)
	if e != nil {
		log.Print("Could not decode image ", e)
		writeError(w, "Error decoding image", http.StatusInternalServerError)
		return
	}
	im = resizeImage(int(x), int(y), im)
	writeImage(w, &im)
	defer response.Body.Close()
}

func writeError(w http.ResponseWriter, reason string, httpCode int) {
	w.WriteHeader(httpCode)
	w.Header().Set("Content-Type", "application/json")
	tmbError := ThumbnailError{false, reason}
	js, err := json.Marshal(tmbError)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(js)
}



// writeImage encodes an image 'img' in jpeg format and writes it into ResponseWriter.
func writeImage(w http.ResponseWriter, img *image.Image) {
	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, *img, nil); err != nil {
		log.Println("Unable to encode image.")
		writeError(w, "Unable to encode image.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Println("Unable to write image.")
		writeError(w, "Unable to write image.", http.StatusInternalServerError)
		return
	}
}

func resizeImage(x int, y int, img image.Image) (image.Image) {
	requiredAspectRatio := float64(y)/float64(x)
	imgX := img.Bounds().Dx()
	imgY := img.Bounds().Dy()
	imageAspectRatio := float64(imgY)/float64(imgX)
	if requiredAspectRatio == imageAspectRatio && x == imgX {
		return img
	} else if requiredAspectRatio == imageAspectRatio && x < imgX{
		return resize.Resize(uint(x), uint(y), img, resize.Lanczos3)
	} else  {
		if x >= imgX && y >= imgY {
			return composeImage(x, y, img)
		}
		ratioX := float64(x)/float64(imgX)
		ratioY := float64(y)/float64(imgY)
		if(ratioX < ratioY) { // scale by x
			img = resize.Resize(uint(x), uint(float64(x) * imageAspectRatio), img, resize.Lanczos3)
			return composeImage(x, y, img)
		}
		// scale by y
		newX := uint(float64(y) / imageAspectRatio)
		img = resize.Resize(newX, uint(y), img, resize.Lanczos3)
		return composeImage(x, y, img)
	}
	return img
}

// compose the image on a larger black rectangle
func composeImage(x int, y int, img image.Image) image.Image {
	rectangle := image.NewRGBA(image.Rect(0, 0, x, y))
	black := color.RGBA{0, 0, 0, 0}
	draw.Draw(rectangle, rectangle.Bounds(), &image.Uniform{black}, image.ZP, draw.Src)
	xPad := (x - img.Bounds().Dx()) / 2
	yPad := (y - img.Bounds().Dy()) / 2
	pt := image.Point{xPad, yPad}
	imgRect := image.Rectangle{pt, pt.Add(img.Bounds().Size())}
	img.Bounds().At(xPad, yPad)
	draw.Draw(rectangle, imgRect, img, image.ZP, draw.Src)
	return rectangle
}


