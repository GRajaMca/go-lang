package main

import (
	"log"
	"net/http"
	"encoding/json"

	"github.com/gorilla/mux"
)

//Config
type Version struct {
	Version string
}


func checkforImage(res http.ResponseWriter, req *http.Request) {
	log.Println("Get Check for Image method started ")

	body := "Image Version 0.0.2" 
	response, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	res.Header().Set("Content-Type", "application/json")
	res.Write(response)
	log.Println("Get Check for Image method completed")
}

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/checkforImage", checkforImage).Methods("GET")
	log.Fatal(http.ListenAndServe(":8880", router))
}
