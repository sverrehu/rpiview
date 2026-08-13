package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"slices"
	"strconv"

	"github.com/sverrehu/goutils/getopt"
	"github.com/sverrehu/rpiview/internal/imagewindow"
)

// curl -X POST --data-binary @"$HOME/Pictures/statements/BushIsrael.png" http://localhost:8086/img

type indexHandler struct{}

type imageHandler struct{}

var imgWindow *imagewindow.ImageWindow

func (h *indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, err := w.Write([]byte("POST a binary image to /img\n"))
	if err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (h *imageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	contentType := http.DetectContentType(content)
	if !supportedFormat(contentType) {
		http.Error(w, "Unsupported content: "+contentType, http.StatusBadRequest)
		return
	}
	img, _, err := image.Decode(bytes.NewReader(content))
	if err != nil {
		http.Error(w, "Unable to decode image", http.StatusBadRequest)
		return
	}
	showImage(&img)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, err = w.Write([]byte("OK\n"))
	if err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func supportedFormat(f string) bool {
	supported := []string{"image/jpeg", "image/png"}
	return slices.Contains(supported, f)
}

func showImage(img *image.Image) {
	imgWindow.SetImage(img)
}

func startWebserver(port int) {
	mux := http.NewServeMux()
	mux.Handle("/", &indexHandler{})
	mux.Handle("POST /img", &imageHandler{})
	log.Printf("Starting server, Ctrl-C to abort. POST images to http://%s:%d/img", localIP(), port)
	err := http.ListenAndServe(":"+strconv.Itoa(port), mux)
	if err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Println("Usage: rpiview [option ...]")
	fmt.Println("")
	fmt.Println("  -p, --port=PORT    web server port to listen to, default: 8086")
	fmt.Println("")
	os.Exit(0)
}

func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "???"
}

func startImageViewer() {
	imgWindow = imagewindow.NewImageWindow()
	imgWindow.Open()
}

func main() {
	help := false
	port := 8086
	opts := []getopt.Option{
		{ShortName: 'h', LongName: "help", Type: getopt.Flag, Target: &help},
		{ShortName: 'p', LongName: "port", Type: getopt.Integer, Target: &port},
	}
	getopt.Parse(&os.Args, opts, false)
	if help {
		usage()
	}
	go startWebserver(port)
	startImageViewer()
}
