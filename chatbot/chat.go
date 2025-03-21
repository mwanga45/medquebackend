package handler_chat

import "net/http"

func Chatbot() {

}

func SendErr(w http.ResponseWriter, message string, statusCode int){
	w.Header().Set("Content-Type","application/json")
}