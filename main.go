package main

import (
	"medquemod/db_conn"
	"log"
	"net/http"
	"fmt"
)

func main() {

// call function connectionpool
const conn_string = "user=postgres dbname=medque password=lynx host=localhost sslmode=disable"
if err := handlerconn.Connectionpool(conn_string);err != nil{
	log.Fatalf("something went wrong failed to connect to database %v", err)
}
defer func(){
	if err := handlerconn.Db.Close();err !=nil{
		log.Fatalf("there is no connection to database")
	}
}()

	fmt.Println("server listen and serve port 8800...")
	err:= http.ListenAndServe(":8800",nil)
	if err != nil{
		log.Fatalln("Failed to create server ")
	}
}