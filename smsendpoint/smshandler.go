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
func SmsBookingCancellationInform(username string, servicename string, start_time string, end_time string, phoneNumber string) error {
	message := fmt.Sprintf(
		"Hi %s!  Your booking for %s is confirmed from %s to %s. "+
			"If you’d like to change your time slot, please do it now while spots are still available. "+
			"Thank you!",
		username, servicename, start_time, end_time,
	)

	Payload := types.SmsPayload{
		SenderID: 55,
		Schedule: "none",
		Sms:      message,
		Recipients: []types.SmsReceiver{{
			Number: phoneNumber,
		}},
	}
	err := utils.SendSms(Payload)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	return nil
}
func SmsSlotAvailableInform(username, servicename, phoneNumber string) error {
	message := fmt.Sprintf(
		"Hi %s! A time slot has become available for %s today. "+
			"If you'd like to reschedule to an earlier appointment time, "+
			"please check our app for available slots.",
		username, servicename,
	)

	payload := types.SmsPayload{
		SenderID: 55,
		Schedule: "none",
		Sms:      message,
		Recipients: []types.SmsReceiver{{
			Number: phoneNumber,
		}},
	}
	return utils.SendSms(payload)
}