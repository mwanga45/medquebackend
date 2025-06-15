package adminact

import (
	"encoding/json"
	"net/http"
)
type (
	ServAssignpayload struct{
		Servname string `json:"servname"`
		DurationMin  int `json:"duration_time"`
		Fee float64 `json:"fee"`
	}
	Response struct{
		Message  string `json:"message"`
		Success bool `json:"success"`
	}
)

func AssignService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost{
		json.NewEncoder(w).Encode(Response{
			Message: "Invalidpayload",
			Success: false,
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}