package routes

import (
	"medquemod/api"
	adminact "medquemod/api-web/adminAct"
	docact "medquemod/api-web/docAct"
	"medquemod/booking"
	handler_chat "medquemod/chatbot"
	authentic "medquemod/handleauthentic"
	"medquemod/middleware"
	"medquemod/profile"

	"github.com/gorilla/mux"
)

func HandleRoutes(r *mux.Router) {

	auth := r.PathPrefix("/auth").Subrouter()
	// Authentication alongSide   the  chatbot routes
	auth.HandleFunc("/login", authentic.HandleLogin).Methods("POST")
	auth.HandleFunc("/register", authentic.Handler).Methods("POST")
	auth.HandleFunc("/chatbot", handler_chat.Chatbot).Methods("POST")
	auth.HandleFunc("/dklogin", docact.DoctLogin).Methods("POST")
	//  SHEDULE  ROUTER FOR DOCTOR INFORMATION
	shedule := r.PathPrefix("/info").Subrouter()
	shedule.HandleFunc("/docAv", api.DoctorsAvailability).Methods("GET")

	// booking routers - protected with authentication middleware
	bookingRoutes := r.PathPrefix("/booking").Subrouter()
	bookingRoutes.Use(middleware.VerifyTokenMiddleware)

	bookingRoutes.HandleFunc("/getservice", api.GetService).Methods("GET")
	bookingRoutes.HandleFunc("/serviceslot", booking.Bookinglogic).Methods("POST")
	bookingRoutes.HandleFunc("/bookingreq", booking.Bookingpayload).Methods("POST")
	bookingRoutes.HandleFunc("/cancelbooking", booking.CancelBooking).Methods("POST")

	// ROUTES FOR THE ADMIN
	Adm := r.PathPrefix("/admin").Subrouter()
	Adm.HandleFunc("/registerserv", adminact.AssignService).Methods("POST")
	Adm.HandleFunc("/regiNonIntervalserv", adminact.AssignNonTimeserv).Methods("POST")
	Adm.HandleFunc("/docschedule", adminact.Asdocschedule).Methods("POST")
	Adm.HandleFunc("/regspecialist", adminact.RegSpecialist).Methods("POST")
	Adm.HandleFunc("/getspecInfo", adminact.ReturnSpec).Methods("GET")
	Adm.HandleFunc("/docAsgnServ", adminact.DocServAssign).Methods("POST")
	Adm.HandleFunc("/DocVsServ", adminact.DocVsServ).Methods("GET")
	Adm.HandleFunc("/login", adminact.AdminLogin).Methods("POST")
	Adm.HandleFunc("/getDocInfo", adminact.GetDoctorInfo).Methods("GET")
	Adm.HandleFunc("/getregserv", adminact.GetsevAvailable).Methods("GET")

	// ROUTES FOR THE DOCTOR
	dkt := r.PathPrefix("/dkt").Subrouter()
	dkt.HandleFunc("/register", docact.Registration).Methods("POST")
	// UserActivit
	userAct := r.PathPrefix("/user").Subrouter()
	userAct.Use(middleware.VerifyTokenMiddleware)
	userAct.HandleFunc("/assignspec", profile.UserAct).Methods("POST")
	userAct.HandleFunc("/bookinghistory", profile.BookingHistory).Methods("GET")

}
