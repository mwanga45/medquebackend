package api

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Messsage string      `json:"message"`
	Success  bool        `json:"success,omitempty"`
	Data     interface{} `json:"Data" `
}

func Doctors(w http.ResponseWriter, r *http.Request) {
   if r.Method != http.MethodGet{
	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(Response{
		Messsage: "Invalid method used ",
		Success: false,
	})
	return
   }

}