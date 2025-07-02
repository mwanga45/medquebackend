package utils

import (
	"medquemod/types"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestSendSms(t *testing.T) {
	// Setup a test HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		// Verify headers
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header 'Bearer test-api-key', got '%s'", authHeader)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "SMS sent successfully"}`))
	}))
	defer mockServer.Close()

	// Set required environment variables
	os.Setenv("SMS_APIKEY", "test-api-key")
	os.Setenv("BASEURL", mockServer.URL)

	// Prepare payload
	payload := types.SmsPayload{
		SenderID:   123,
		Schedule:   "now",
		Sms:        "Test SMS",
		Recipients: []types.SmsReceiver{{Number: "+255700000000"}},
	}

	// Call SendSms
	err := SendSms(payload)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}
