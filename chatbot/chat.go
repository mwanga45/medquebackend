package handler_chat

import (
	"encoding/json"
	"net/http"
)
//create struct that will hold text field
type Part struct{
	Text string `json:"text"`
}
// create that will hold two field   that role will be either user or model and another field will be slice of content 
type Content struct{
	Role string `json:"role"`
	Parts []Part `json:"parts"`
}
// create struct that will hold field to configure the userInput  in property 
type GenerateConfig struct{
	Temperature float32 `json:"temperature"`
	TopK int`json:"topk"`
	TopP  float32 `json:"topP"`
	MaxOutputToken int `json:"maxoutputtoken"`
}
// 
type chatbotResponse struct{
	Response string `json:"response"`
	MessageResonseError string `json:"messageResponseError,omitempty"`
}

func Chatbot() {

}

func SendErr(w http.ResponseWriter, message string, statusCode int){
	w.Header().Set("Content-Type","application/json")
	json.NewEncoder(w).Encode(chatbotResponse{
		MessageResonseError: message,
	})
}