package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/cenkalti/remux"
)

const maxContentLength = 256 * 1024

var transport = &http.Transport{
	DisableCompression: true,
}

var client = &http.Client{
	Transport: transport,
}

var secret string

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}
	secret = os.Getenv("SECRET")
	if secret == "" {
		log.Fatal("$SECRET must be set")
	}
	var r remux.Remux
	r.HandleFunc("/(?P<secret>.+)/md5/(?P<target>.+)", handleMD5).Get()
	r.HandleFunc("/", handleIndex).Get()
	http.ListenAndServe(":"+port, r)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("https://github.com/cenkalti/hashgen"))
}

func handleMD5(w http.ResponseWriter, r *http.Request) {
	if r.FormValue(":secret") != secret {
		http.Error(w, "invalid secret", http.StatusUnauthorized)
		return
	}
	target := r.FormValue(":target")
	log.Println(target)
	resp, err := client.Get(target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	if resp.StatusCode >= 400 {
		http.Error(w, "bad status code: "+strconv.Itoa(resp.StatusCode), http.StatusBadGateway)
		return
	}
	if resp.ContentLength == -1 {
		http.Error(w, "no content length", http.StatusBadGateway)
		return
	}
	h := md5.New()
	_, err = io.Copy(h, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	fmt.Fprintf(w, "%x", h.Sum(nil))
}
