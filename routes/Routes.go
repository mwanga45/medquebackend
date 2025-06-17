package Routesf_test

import (
	"medquemod/api"
	adminact "medquemod/api-web/adminAct"
	docact "medquemod/api-web/docAct"
	"medquemod/booking"
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
	//  SHEDULE  ROUTER FOR DOCTOR INFORMATION
	shedule := r.PathPrefix("/info").Subrouter()
	shedule.HandleFunc("/docAv", api.DoctorsAvailability).Methods("GET")

	// booking routers
	bookingRoutes := r.PathPrefix("/booking").Subrouter() //subrouter for the booking

	bookingRoutes.HandleFunc("/getservice", api.GetService).Methods("GET")
	bookingRoutes.HandleFunc("/serviceslot", booking.Bookinglogic).Methods("POST")

	// ROUTES FOR THE ADMIN
	Adm := r.PathPrefix("/adim").Subrouter()
	Adm.HandleFunc("/registerserv",adminact.AssignService ).Methods("POST")
	Adm.HandleFunc("/docshedule", adminact.Asdocshedule).Methods("POST")
	Adm.HandleFunc("/regspecilist",adminact.RegSpecialist).Methods("POST")
	Adm.HandleFunc("/getspecInfo",adminact.ReturnSpec).Methods("GET")
	Adm.HandleFunc("/docAsgnServ",adminact.DocServAssign).Methods("POST")

	// ROUTES FOR THE DOCTOR
	dkt := r.PathPrefix("/dkt").Subrouter()
	dkt.HandleFunc("/register",docact.Registration).Methods("POST")

}