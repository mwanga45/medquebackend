package utils

import (
	"log"
	"net/http"
	"net/url"
	"os"
)

func SendSms(to string, message string) error {
	apiKey := os.Getenv("SMS_APIKEY")
	baseURL := os.Getenv("BASEURL")

	// Construct the API endpoint
	endpoint := baseURL + "/send"

	// Create the request payload
	data := url.Values{}
	data.Set("to", to)
	data.Set("message", message)

	// Make the HTTP POST request
	resp, err := http.PostForm(endpoint, data)
	if err != nil {
		log.Printf("Error sending SMS: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to send SMS: %s", resp.Status)
		return err
	}

	log.Println("SMS sent successfully")
	return nil
}
