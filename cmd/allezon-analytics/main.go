package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func UserTagsPostHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/user_tags", UserTagsPostHandler).Methods(http.MethodPost)

	log.Fatal(http.ListenAndServe(":8080", r))
}
