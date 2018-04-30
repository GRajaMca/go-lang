package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

//var environmentInfo *environment
var environmentInfo environmentInformation

func main() {

	environmentInfo.getConf()
	router := mux.NewRouter()
	router.HandleFunc("/getManifest", getManifest).Methods("POST")
	router.HandleFunc("/processUpdate", processUpdate).Methods("POST")
	router.HandleFunc("/updateDockerService", updateDockerService).Methods("POST")
	router.HandleFunc("/getNotification/{RepositoryURL}/{ImageURL}", getNotification).Methods("GET")
	router.HandleFunc("/checkforUpdate", checkforUpdate).Methods("GET")
	log.Fatal(http.ListenAndServe(":8002", router))
}
