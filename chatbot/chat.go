package handler_chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// create struct that will hold text field
type (
	Part struct {
		Text string `json:"text"`
	}
	// create that will hold two field   that role will be either user or model and another field will be slice of content
	Content struct {
		Role  string `json:"role"`
		Parts []Part `json:"parts"`
	}
	// create struct that will hold field to configure the userInput  in property
	GenerateConfig struct {
		Temperature    float32 `json:"temperature"`
		TopK           int     `json:"topk"`
		TopP           float32 `json:"topP"`
		MaxOutputToken int     `json:"maxoutputtoken"`
	}
	// create the struct  will check the safety of message
	SafetySetting struct {
		Category  string `json:"category"`
		Threshold string `json:"threshold"`
	}
	// create struct that will able to generate request
	GenerateContentRequest struct {
		Contents       []Content       `json:"contents"`
		SafetySetting  []SafetySetting `json:"safetysetting"`
		GenerateConfig GenerateConfig  `json:"generateconfig"`
	}
	// create struct that will able to return response to user
	GenerateContentResponse struct {
		Candidates  []struct {
			Content struct {
				Parts []Part `json:"parts"`
			} `json:"contents"`
		} `json:"candidate"`
	}
	ChatResponse struct {
		Response            string `json:"response"`
		MessageResonseError string `json:"messageResponseError,omitempty"`
	}
	ChatRequest struct {
		UserInput string `json:"userInput"`
	}
)
const geminiEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent"

func Chatbot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendErr(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := ProcessChatRequest(req.UserInput)
	if err != nil {
		SendErr(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}
func ProcessChatRequest(userInput string) (*ChatResponse, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing GEMINI_API_KEY environment variable")
	}

	geminiReq := CreateGeminiRequest(userInput)
	responseText, err := CallGeminiAPI(geminiReq, apiKey)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %v", err)
	}

	return &ChatResponse{Response: responseText}, nil
}
func CreateGeminiRequest(userInput string) *GenerateContentRequest {
	return &GenerateContentRequest{
		Contents: []Content{
			{
				Role: "user",
				Parts: []Part{
					{
						Text: "Modele promit here ",
					},
				},
			},
			{
				Role: "user",
				Parts: []Part{
					{
						Text: userInput,
					},
				},
			},
		},
		GenerateConfig: GenerateConfig{
			Temperature: 0.9,
			TopK:1 ,
			TopP: 1,
			MaxOutputToken: 1000,
			
		},
		SafetySetting: []SafetySetting{
			{
				Category: "HARM_CATEGORY_HARASSMENT",
				Threshold: "BLOCK_MEDIUM_AND_ABOVE",
			},
		},
	}
}
func CallGeminiAPI(req *GenerateContentRequest, apiKey string) (string, error) {
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	url := fmt.Sprintf("%s?key=%s", geminiEndpoint, apiKey)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned %d: %s", resp.StatusCode, body)
	}

	var geminiResp GenerateContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}
func SendErr(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{
		MessageResonseError: message,
	})
}
