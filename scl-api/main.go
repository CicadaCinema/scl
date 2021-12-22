package main

import (
	"log"
	"net/http"
	postHandler "scl-server/api/authenticate"
)

func main() {
	http.HandleFunc("/api/authenticate", postHandler.Handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
