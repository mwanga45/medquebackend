package routes

import (
	"medquemod/api"
	handler_chat "medquemod/chatbot"
	authentic "medquemod/handleauthentic"

	"github.com/gorilla/mux"
)

func HandleRoutes(r *mux.Router) {

	auth:= r.PathPrefix("/auth").Subrouter() //subrouter for the authentication
  
	// Authentication alongSide   the  chatbot routes
	auth.HandleFunc("/login",authentic.HandleLogin).Methods("POST")
	auth.HandleFunc("/register", authentic.Handler).Methods("POST")
	auth.HandleFunc("/chatbot", handler_chat.Chatbot).Methods("POST")

	// booking routers
	booking := r.PathPrefix("/booking").Subrouter() //subrouter for the booking

	booking.HandleFunc("/getservice", api.GetService).Methods("GET")

	

}