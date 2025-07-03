package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"medquemod/types"
	"net/http"
	"os"
)

func SendSms(payload types.SmsPayload) error {
	apiKey := os.Getenv("SMS_APIKEY")
	baseURL := os.Getenv("BASEURL")+ "/send-sms"
	if apiKey == "" {
		return fmt.Errorf("SMS_APIKEY environment variable is not set")
	}

	if baseURL == "" {
		return fmt.Errorf("BASEURL environment variable is not set")
	}

	
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(bodyBytes))
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
