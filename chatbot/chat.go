package handler_chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
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
	GenerationConfig struct {
		Temperature     float32 `json:"temperature"`
		TopK            int     `json:"topK"`
		TopP            float32 `json:"topP"`
		MaxOutputTokens int     `json:"maxOutputTokens"`
	}
	// create the struct  will check the safety of message
	SafetySetting struct {
		Category  string `json:"category"`
		Threshold string `json:"threshold"`
	}
	// create struct that will able to generate request
	GenerateContentRequest struct {
		Contents         []Content        `json:"contents"`
		SafetySettings   []SafetySetting  `json:"safetySettings"`   // Fixed case
		GenerationConfig GenerationConfig `json:"generationConfig"` // Fixed name
	}
	// create struct that will able to return response to user
	GenerateContentResponse struct {
		Candidates []struct {
			Content struct {
				Parts []Part `json:"parts"`
			} `json:"content"` // Correct key
		} `json:"candidates"` // Correct plural
	}
	ChatResponse struct {
		Response string `json:"response"`
		Error    string `json:"error,omitempty"` // Standardized error field
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
		SendErr(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}
func ProcessChatRequest(userInput string) (*ChatResponse, error) {
	godotenv.Load(".env")
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("Missing GEMINI_API_KEY environment variable")
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
				Parts: []Part{{
					Text: `You are Sam, a friendly assistant...younhave been create by Issa Mwanga [User: Hi, I have a question about managing my blood pressure.
AI Chatbot: Hello! I’m here to provide you with general health information. However, please note that I am not a doctor, and any information I provide should not replace professional medical advice. Could you tell me a bit more about your situation or what specifically you’d like to know about managing blood pressure?
User: I'm looking for lifestyle changes that might help lower my blood pressure.
AI Chatbot: Great, I can share some common lifestyle recommendations that many people find helpful. Often, these include:

Dietary Changes: Incorporating a diet rich in fruits, vegetables, and whole grains while reducing salt intake.

Physical Activity: Engaging in regular exercise like brisk walking or swimming.

Stress Management: Techniques such as meditation, yoga, or deep breathing exercises.

Monitoring: Keeping track of your blood pressure regularly and noting any changes.

Please keep in mind that these suggestions are general, and it's important to consult with a healthcare provider for personalized advice. Do you have any more questions or need further details on any of these points?]`,
				}},
			},
			{
				Role:  "model",
				Parts: []Part{{Text: "Hello! Welcome to Coding Money..."}},
			},
			{
				Role:  "user",
				Parts: []Part{{Text: "Hi"}},
			},
			{
				Role:  "model",
				Parts: []Part{{Text: "Hi there! Thanks for reaching out..."}},
			},
			{
				Role:  "user",
				Parts: []Part{{Text: userInput}},
			},
		},
		GenerationConfig: GenerationConfig{
			Temperature:     0.9,
			TopK:            1,
			TopP:            1,
			MaxOutputTokens: 1000,
		},
		SafetySettings: []SafetySetting{
			{
				Category:  "HARM_CATEGORY_HARASSMENT",
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
	w.WriteHeader(statusCode) // Add this line
	json.NewEncoder(w).Encode(ChatResponse{
		Error: message, // Use corrected error field
	})
}
