package smsendpoint

import (
	"fmt"
	"os"
)

type (
	Respond struct {
		Message string `json:"message"`
		Success bool   `json:"success"`
	}
	Payload struct {
		Sender_id string `json:"sender_id"`
		Shedule   string `json:"shedule"`
		Sms       string `json: "sms:`
		Recipient string `json:"recepient"`
	}
	Receiver struct {
		Number int `json:"number"`
	}
)

func Sms_endpoint(username string, phone string, start_at string, end_at string) {
    Url :=  fmt.Sprintf("https://api.notify.africa/v2")
	apikey  :=  os.Getenv("SMS_APIKEY")
	messagestring := fmt.Sprintf( username+" " + "your suppose to be at hospital  10 minitues less than  before the appointment reach  start at_%s and end at %s",start_at,end_at)
	payload := Payload{
		Sender_id: "TAARIFA",
		Shedule: "none",
		Sms:messagestring,
		Recipient: phone,
	}
	bodypayload , err := 
}