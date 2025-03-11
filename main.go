package main

import (
	"log"
	"net/http"
	"fmt"
)

func main() {
	fmt.Println("server listen and serve port 8800...")
	err:= http.ListenAndServe(":8800",nil)
	if err != nil{
		log.Fatalln("Failed to create server ")
	}
}