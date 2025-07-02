package smsendpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type (
	Respond struct {
		Message string `json:"message"`
		Success bool   `json:"success"`
	}

	Payload struct {
		SenderID   int        `json:"sender_id"`
		Schedule   string     `json:"schedule"`
		Sms        string     `json:"sms"`
		Recipients []Receiver `json:"recipients"`
	}

	Receiver struct {
		Number string `json:"number"`
	}
)

func SmsEndpoint(username, phone, startAt, endAt string) error {

	url := "https://api.notify.africa/v2/send-sms"

	apiKey := os.Getenv("SMS_APIKEY")
	if apiKey == "" {
		return fmt.Errorf("SMS_APIKEY environment variable is not set")
	}

	message := fmt.Sprintf("%s, you're expected at the hospital 10 minutes before your appointment. Start: %s, End: %s", username, startAt, endAt)

	payload := Payload{
		SenderID: 55,
		Schedule: "none",
		Sms:      message,
		Recipients: []Receiver{{
			Number: phone,
		}},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("Status: %s\nResponse: %s\n", resp.Status, string(respBody))

	return nil
}
