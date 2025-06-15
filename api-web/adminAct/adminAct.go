package adminact

import (
	"database/sql"
	"encoding/json"
	"fmt"
	handlerconn "medquemod/db_conn"
	"net/http"
)

type (
	ServAssignpayload struct {
		Servname    string  `json:"servname"`
		DurationMin int     `json:"duration_time"`
		Fee         float64 `json:"fee"`
	}

	Response struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Data    interface{} `json:"data,omitempty"`
	}
	Asdocshedulepayload struct{
		DoctorID int `json:"docId"`
		Dayofweek int `json:"dow"`
		StartTime  string `json:"start_time"`
		EndTime string `json:"end_time"`
	}
)

func AssignService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(Response{
			Message: "Invalidpayload",
			Success: false,
		})
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var Assreq ServAssignpayload
	json.NewDecoder(r.Body).Decode(&Assreq)
	var serv_id string
	checkifexisterr := handlerconn.Db.QueryRow(`SELECT serv_id,servicename FROM serviceAvailable  WHERE servicename = $1`, Assreq.Servname).Scan(&serv_id)
	if checkifexisterr != nil {
		if checkifexisterr == sql.ErrNoRows {
			_, queryerr := handlerconn.Db.Exec(`INSERT INTO serviceAvailable (servicename,duration_minutes, fee) VALUES($1,$2,$3)`, Assreq.Servname, Assreq.DurationMin, Assreq.Fee)
			if queryerr != nil {
				json.NewEncoder(w).Encode(Response{
					Message: "Internal serverError",
					Success: false,
				})
				fmt.Println("failed to insertdata", queryerr)
				return
			}
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "failed to assign new service",
			Success: false,
		})
		return
	}
	json.NewEncoder(w).Encode(Response{
		Message: "Servecename is already exist",
		Success: false,
		Data: serv_id,
	})


}

func Asdocshedule(w http.ResponseWriter, r *http.Request)  {
	if r.Method != http.MethodPost{
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Badrequest",
			Success: false,
		})
		return
	}


}
