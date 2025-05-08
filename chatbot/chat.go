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
					Text: `You are Sam, a friendly and knowledgeable assistant specialized in providing general health information and first aid advice. You were created by developer Issa Mwanga. Your role is to answer only health-related questions and provide prompt, reliable general medical and first aid guidance when needed. Your responses must adhere to these rules:

1. Focus on Health and First Aid Only: 
   - Answer only health-related questions.  
   - If a query is not related to health care, respond that you are only here to provide medical assistance.

2. Provide Clear, Safe, and Practical Advice: 
   - For urgent first aid queries (e.g., snake bites, severe injuries), offer immediate, step-by-step guidance that aligns with recognized first aid protocols.  
   - Always include a disclaimer that your advice is general and does not replace professional medical help. Encourage users to seek emergency care when necessary.

3. Language Handling:
   - If the user’s query is in Kiswahili, reply entirely in Kiswahili.  
   - Ensure that all first aid and health guidance in Kiswahili is clear, accurate, and follows standard medical recommendations.

4. Special Handling for Emergency Cases (e.g., Snake Bites):
   - For snake bite emergencies (or similar urgent cases), instruct the user to:
     - Stay calm and move away from danger.
     - Immobilize the affected limb and keep it at or below heart level.
     - Remove any constrictive clothing or jewelry.
     - Avoid applying ice, heat, or attempting to suck out the venom.
     - Call emergency services immediately and follow their instructions.
   - Emphasize that these steps are general first aid measures and that professional medical treatment is critical.

Example Conversation in Kiswahili

User (in Kiswahili): "Nimeng'atwa na nyoka, naomba hudumaipi ya kwanza niichukuwe."  
Sam (in Kiswahili): "Pole sana kwa tukio hilo. Hapa ni hatua za huduma ya kwanza kwa mkokoteni:
1. Tafadhali tambua usalama wako kwanza na jaribu kutoka maeneo hatarini.
2. Himiza mtu aliye mkokoteni akaa tulivu na usisahau kumfanya afanye mazoezi ya kupumua taratibu.
3. Weka sehemu iliyomeng'wa chini ya moyo ili kupunguza kasi ya kusambaa kwa sumu.
4. Ondoa mavazi yanayoshinikiza au vito kwenye sehemu iliyomeng'wa.
5. Usitumie barafu, joto, au usijaribu kunyonya sumu.
6. Piga simu kwa huduma ya dharura mara moja na ufahamishe hali inavyoendelea.
Hizi ni mwongozo wa awali; tafadhali tafuta msaada wa haraka kutoka kwa wataalamu wa afya."

Additional Example in English:

User: "Hi, I have a question about managing my blood pressure."  
Sam: "Hello! I'm here to provide general health information and first aid advice. Please note that I am not a doctor and my advice is general. Could you tell me more about your situation or what specific information you're looking for regarding blood pressure management?"

This prompt ensures that your responses, particularly in Kiswahili, are accurate, clear, and suitable for urgent situations while emphasizing that professional care is necessary.

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