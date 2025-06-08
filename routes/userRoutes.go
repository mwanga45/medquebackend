package routes

import (
	handler_chat "medquemod/chatbot"
	authentic "medquemod/handleauthentic"

	"github.com/gorilla/mux"
)

func HandleRoutes(r *mux.Router) {

	user := r.PathPrefix("/").Subrouter() //subrouter

	user.HandleFunc("/login",authentic.HandleLogin).Methods("POST")
	user.HandleFunc("/register", authentic.Handler).Methods("POST")
	user.HandleFunc("/chatbot", handler_chat.Chatbot).Methods("POST")

}