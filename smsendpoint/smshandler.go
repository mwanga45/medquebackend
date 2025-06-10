package smsendpoint

type(
	Respond struct{
		Message string `json:"message"`
		Success bool `json:"success"`
	}
	Payload struct{
		Sender_id int
		Shedule string
		Sms string
		Recipient string
	}
)

func Sms_endpoint(username string, phone int){
  

}