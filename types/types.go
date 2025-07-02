package types

type (
	SmsRespond struct {
		Message string `json:"message"`
		Success bool   `json:"success"`
	}

	SmsPayload struct {
		SenderID   int        `json:"sender_id"`
		Schedule   string     `json:"schedule"`
		Sms        string     `json:"sms"`
		Recipients []SmsReceiver `json:"recipients"`
	}

	SmsReceiver struct {
		Number string `json:"number"`
	}
)