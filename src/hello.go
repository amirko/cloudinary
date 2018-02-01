package main

import (
	"fmt"
	"bytes"
//	"encoding/base64"
	"image"
	//"image/color"
	//"image/draw"
	"github.com/nfnt/resize"
	"image/jpeg"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	log "log"
	"strconv"
)
type ThumbnailError struct{
	Reason string
}

func main() {
	fmt.Printf("Welcome to thumbnail service!\n")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/thumbnail", thumbnail)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func thumbnail(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	url := params.Get("url")
	x, err1 := strconv.ParseUint(params.Get("x"), 10, 32)
	y, err2 := strconv.ParseUint(params.Get("y"), 10, 32)
	if err1 != nil || err2 != nil {
		writeError(w, "Cant parse params", http.StatusBadRequest)
		return
	}
	response, e := http.Get(url)
	if e != nil {
		log.Print("Could not read image")
		writeError(w, "Not Found", http.StatusNotFound)
		return
	}
	im, _, e := image.Decode(response.Body)
	if e != nil {
		log.Print("Could not decode image")
		writeError(w, "Error decoding image", http.StatusInternalServerError)
		return
	}
	im = resizeImage(x, y, im)
	writeImage(w, &im)
	defer response.Body.Close()

//	writeImage(w, )
}

func writeError(w http.ResponseWriter, reason string, httpCode int) {
	w.WriteHeader(httpCode)
	tmbError := ThumbnailError{reason}
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

func resizeImage(x uint64, y uint64, img image.Image) (image.Image) {
	requiredAspectRatio := float64(y)/float64(x)
	imgX := uint64(img.Bounds().Max.X)
	imgY := uint64(img.Bounds().Max.Y)
	imageAspectRatio := float64(imgY)/float64(imgX)
	if requiredAspectRatio == imageAspectRatio && imgX == x {
		return img
	} else if requiredAspectRatio == imageAspectRatio && imgX < x {
		return resize.Resize(uint(x), uint(y), img, resize.Lanczos3)
	}
	return img
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}


