package Routesf_test

import (
	"medquemod/api"
	"medquemod/booking"
	handler_chat "medquemod/chatbot"
	authentic "medquemod/handleauthentic"
	"medquemod/registerservice"

	"github.com/gorilla/mux"
)

func HandleRoutes(r *mux.Router) {

	auth:= r.PathPrefix("/auth").Subrouter() //subrouter for the authentication
  
	// Authentication alongSide   the  chatbot routes
	auth.HandleFunc("/login",authentic.HandleLogin).Methods("POST")
	auth.HandleFunc("/register", authentic.Handler).Methods("POST")
	auth.HandleFunc("/chatbot", handler_chat.Chatbot).Methods("POST")

	// booking routers
	bookingRoutes := r.PathPrefix("/booking").Subrouter() //subrouter for the booking

	bookingRoutes.HandleFunc("/getservice", api.GetService).Methods("GET")
	bookingRoutes.HandleFunc("/serviceslot", booking.Bookinglogic).Methods("POST")

	// ROUTES FOR THE ADMIN
	Adm := r.PathPrefix("/adim").Subrouter()

	Adm.HandleFunc("/registerserv",Service.Registerserv ).Methods("POST")

}