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

	// SmsRequest struct for the SMS endpoint
	// Accepts username, phone, startAt, and endAt as JSON fields
	SmsRequest struct {
		Username string `json:"username"`
		Phone    string `json:"phone"`
		StartAt  string `json:"startAt"`
		EndAt    string `json:"endAt"`
	}
)

func SmsEndpoint(username, phone, startAt, endAt string) error {

	url := fmt.Sprintf("https://api.notify.africa/v2/send-sms")

	apiKey := os.Getenv("SMS_APIKEY")
	if apiKey == "" {
		return fmt.Errorf("SMS_APIKEY environment variable is not set")
	}

	message := fmt.Sprintf("%s, you're expected at the hospital 10 minutes before your appointment. Start: %s, End: %s", username, startAt, endAt)

	payload := Payload{
		SenderID: 1,
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

// SendSMSHandler is a simple HTTP handler to send an SMS using SmsEndpoint
func SendSMSHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Respond{
			Message: "Method not allowed",
			Success: false,
		})
		return
	}
	var req SmsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Respond{
			Message: "Invalid request payload",
			Success: false,
		})
		return
	}
	err := SmsEndpoint(req.Username, req.Phone, req.StartAt, req.EndAt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Respond{
			Message: err.Error(),
			Success: false,
		})
		return
	}
	json.NewEncoder(w).Encode(Respond{
		Message: "SMS sent successfully",
		Success: true,
	})
}
