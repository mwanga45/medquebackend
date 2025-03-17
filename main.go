package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"medquemod/api"
	"medquemod/db_conn"
	"medquemod/handleauthentic"
	"net/http"
)

func main() {

	r := mux.NewRouter()
	// r.HandleFunc("/register",authentic.Reg_authentic)
	r.HandleFunc("/register", authentic.Handler).Methods("POST")
	r.HandleFunc("/doctorinfo", api.Doctors).Methods("GET")
	r.HandleFunc("/userinfo", api.Userdetails).Methods("GET")

	// call function connectionpool
	const conn_string = "user=postgres dbname=medque password=lynx host=localhost sslmode=disable"
	if err := handlerconn.Connectionpool(conn_string); err != nil {
		log.Fatalf("something went wrong failed to connect to database %v", err)
	}
	// defer func(){
	// 	if err := handlerconn.Db.Close();err !=nil{
	// 		log.Fatalf("there is no connection to database")
	// 	}
	// }()
	// configure Cors middleware
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"POST", "PUT", "GET", "DELETE"}),
		handlers.AllowedHeaders([]string{"content-type", "Authorization"}),
	)

	fmt.Println("server listen and serve port 8800...")
	err := http.ListenAndServe(":8800", cors(r))
	if err != nil {
		log.Fatalln("Failed to create server ")
	}
}
