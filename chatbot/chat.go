package handler_chat

import (
	"encoding/json"
	"net/http"
)
type chatbotResponse struct{
	MessageResonseError string `json:"messageResponseError"`
}

func Chatbot() {

}

func SendErr(w http.ResponseWriter, message string, statusCode int){
	w.Header().Set("Content-Type","application/json")
	json.NewEncoder(w).Encode(chatbotResponse{
		MessageResonseError: message,
	})
}