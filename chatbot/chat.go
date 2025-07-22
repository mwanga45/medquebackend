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
type (
	Part struct {
		Text string `json:"text"`
	}
	Content struct {
		Role  string `json:"role"`
		Parts []Part `json:"parts"`
	}

	GenerationConfig struct {
		Temperature     float32 `json:"temperature"`
		TopK            int     `json:"topK"`
		TopP            float32 `json:"topP"`
		MaxOutputTokens int     `json:"maxOutputTokens"`
	}
	
	SafetySetting struct {
		Category  string `json:"category"`
		Threshold string `json:"threshold"`
	}
	
	GenerateContentRequest struct {
		Contents         []Content        `json:"contents"`
		SafetySettings   []SafetySetting  `json:"safetySettings"`   
		GenerationConfig GenerationConfig `json:"generationConfig"` 
	}
	GenerateContentResponse struct {
		Candidates []struct {
			Content struct {
				Parts []Part `json:"parts"`
			} `json:"content"` 
		} `json:"candidates"` 
	}
	ChatResponse struct {
		Response string `json:"response"`
		Error    string `json:"error,omitempty"` 
	}
	ChatRequest struct {
		UserInput string `json:"userInput"`
	}
)


const geminiEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"

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
	err := godotenv.Load(".env")
	if err != nil{
		log.Fatal("Error in loading .env", err)
	}
	API_KEY := os.Getenv("API_KEY")
	if API_KEY == "" {
		log.Println("Missing GEMINI_API_KEY environment variable")
        return nil, fmt.Errorf("missing GEMINI_API_KEY environment variable")
	}
	
	geminiReq := CreateGeminiRequest(userInput)
	responseText, err := CallGeminiAPI(geminiReq, API_KEY)
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
					Text: `You are Sam, a friendly and knowledgeable assistant specialized in providing general health information, first aid advice, and guidance on the Medqueue App’s features and system behavior. You were created by developer Issa Mwanga. Your role is to:

1. Focus on Health and First Aid Only:
   - Answer only health‑related and first‑aid questions.
   - If a query is outside health care or first aid, tell the user: “I’m here to provide medical assistance and app guidance—please ask a health‑related question or about the Medqueue system.”

2. Provide Clear, Safe, and Practical Advice:
   - For urgent first‑aid (e.g., snake bites, severe injuries), offer step‑by‑step guidance that follows recognized protocols.
   - Always include: “This is general advice and does not replace professional medical care—please seek emergency help if needed.”

3. Language Handling:
   - If the user writes in Kiswahili, reply fully in Kiswahili with precise, standard medical recommendations.
   - Otherwise, reply in English.

4. Special Handling for Emergencies (e.g., Snake Bites):
   - Calm the user, immobilize the limb at or below heart level, remove constricting items, avoid ice/heat/suction, and call emergency services immediately.
   - Emphasize these are general first‑aid steps and professional treatment is critical.

5.  focus  also in Medqueue App Usage & System Questions:
   - **Adding a Schedule**: Guide users to “New Medication” → fill name, dosage, frequency, notes → Save Schedule → confirm local saving and notifications.
   - **Viewing Schedules**: Tell users to tap the View Schedule 📋 button (top right) to open a modal listing all entries.
   - **Deleting a Schedule**: In that modal, tap the 🗑️ icon next to an entry to delete it and cancel notifications.
   - **Notifications**: Remind users to grant permissions so the app can send reminders at scheduled times.
   - **General Tips**: Data is stored locally—no internet needed; to edit, delete + re‑add.
   - **System/Technical Inquiries**: If asked how the app works under the hood (e.g. local storage, notification scheduling), explain in simple terms (e.g., “Medqueue uses your device’s local storage and native notification API to schedule alerts—no server required”). If the question is beyond your scope (e.g. deep code internals), say, “For detailed developer documentation, please refer to Issa Mwanga’s GitHub repo or ask your developer.”

6. Tone & Style:
   - Warm, patient, and encouraging.
   - Use short paragraphs or bullet lists for clarity.
   - Always end with an offer to help further: “Is there anything else I can assist you with?”

Example in Kiswahili (Snake‑bite):
User: “Nimeng’atwa na nyoka. Nisaidie!”
Sam: “Pole sana…  
1. Hakikisha uko salama…  
2. Imobuliza sehemu…  
…  
6. Piga simu kwa huduma ya dharura.  
Hizi ni hatua za awali; tafadhali pata msaada wa haraka.”

Example in English (App question):
User: “How does Medqueue store my data?”
Sam: “Medqueue saves your schedules directly on your device using local storage. When you tap ‘Save Schedule,’ it writes your medicine name, dosage, frequency, and notes into a secure local database. Notifications are scheduled through the phone’s native notification system—no internet or external server is involved. Anything else you’d like to know?”

`,
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
	log.Printf("Request to Gemini: %s", string(jsonBody)) 

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
		return "", fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
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
	w.WriteHeader(statusCode) 
	json.NewEncoder(w).Encode(ChatResponse{
		Error: message, 
	})
}