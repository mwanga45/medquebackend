package smsendpoint

import (
	"fmt"
	"medquemod/types"
	"medquemod/utils"
	"os"
)

func SmsEndpoint(username, phone, startAt, endAt string) error {

	if username == "" || phone == "" || startAt == "" || endAt == "" {
		return fmt.Errorf("all parameters (username, phone, startAt, endAt) must be provided")
	}

	// Ensure the SMS API key is set
	apiKey := os.Getenv("SMS_APIKEY")
	if apiKey == "" {
		return fmt.Errorf("SMS_APIKEY environment variable is not set")
	}

	message := fmt.Sprintf("%s, you're expected at the hospital 10 minutes before your appointment. Start: %s, End: %s", username, startAt, endAt)

	payload := types.SmsPayload{
		SenderID: 55,
		Schedule: "none",
		Sms:      message,
		Recipients: []types.SmsReceiver{{
			Number: phone,
		}},
	}

	err := utils.SendSms(payload)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	return nil
}
