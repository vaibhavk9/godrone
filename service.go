package main

import (
	"github.com/gorilla/mux"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

func HttpServe() error {

	log.Infoln("Starting HTTP Service on - ", httpAddr)

	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(NoHandleFound)

	router.HandleFunc("/Health", HealthCehck).Methods("GET")

	ra := router.PathPrefix("/{APIVersion}/Accounts/{AccountSid:AC[0-9a-fA-F]{32}}/Applications/{ApplicationSid:AP[0-9a-fA-F]{32}}").Subrouter()
	ra.HandleFunc("/Clients{format:(?:\\.xml|\\.csv|\\.json)?}", ListApplicationClients).Methods("GET")
	ra.HandleFunc("/Clients/{ClientSid:GT[0-9a-fA-F]{32}}{format:(?:\\.xml|\\.csv|\\.json)?}", GetApplicationClient).Methods("GET")
	ra.HandleFunc("/Clients/{ClientSid:GT[0-9a-fA-F]{32}}{format:(?:\\.xml|\\.csv|\\.json)?}", DeleteApplicationClient).Methods("DELETE")
	ra.HandleFunc("/Clients/{ClientSid:GT[0-9a-fA-F]{32}}{format:(?:\\.xml|\\.csv|\\.json)?}", CreateApplicationClient).Methods("POST")

	ServeWithContext := ReqContextWithAuth(router)

	return http.ListenAndServe(httpAddr, ServeWithContext)

}
