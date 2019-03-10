package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/suyashkumar/bin/handlers"
	"golang.org/x/crypto/acme/autocert"
)

var domain = flag.String("domain", "",
	"domain used for this currently running instance (enables SSL, and mints certs through LetsEncrypt")

func main() {
	router := httprouter.New()
	router.GET("/:username/:repo", handlers.Download)
	if *domain != "" {
		log.Printf("Listening at https://%s", *domain)
		log.Fatal(http.Serve(autocert.NewListener(*domain), router))
	} else {
		log.Printf("Listening at localhost:8000")
		log.Fatal(http.ListenAndServe(":8000", router))
	}
}
