package registerservice

import (
	"encoding/json"
	"fmt"
	handlerconn "medquemod/db_conn"
	"net/http"
)

type (
	Payload struct {
		Servname       string  `json:"Servname"`
		DurationMinute int     `json:"durationminutes"`
		Fee            float32 `json:"fee"`
	}
	Respond struct{
		Message string
		Success bool
	}

)

func registerservice(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost{
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Respond{
			Message: "Invalid payload",
			Success: false,
		})
      return
	}
	w.Header().Set("Content-Type", "application/json")

	var payload Payload

	json.NewDecoder(r.Body).Decode(&payload)

	_,err :=  handlerconn.Db.Exec(`INSERT INTO serviceAvailable (servicename, duration_minutes, fee) VALUES($1,$2,$3) `, payload.Servname, payload.DurationMinute, payload.Fee)
	if err != nil{
		json.NewEncoder(w).Encode(Respond{
			Message: "Internal Server Error",
			Success: false,
		})
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Something went wrong", err)
		return
	}
	json.NewEncoder(w).Encode(Respond{
		Message: "Successfuly register new service",
		Success: true,
	})
}