package booking

import (
	"encoding/json"
	handlerconn "medquemod/db_conn"
	"net/http"
)

type (
	Respond struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
	}
	BookingRequest struct {
		Time       int64  `json:"time" validate:"required"`
		Department string `json:"department" validate:"required"`
		Day        string `json:"day" validate:"required"`
		Diseases   string `json:"desease" validate:"required"`
		Doctor     string `json:"doctor" validate:"required"`
	}
)

func Booking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost{
       w.WriteHeader(http.StatusMethodNotAllowed)
	   json.NewEncoder(w).Encode(Respond{
		Message: "Method isnt Allowed",
		Success: false,
	   })
	   return
	}
	w.Header().Set("Content-Type", "application/json")
	tx, errTx := handlerconn.Db.Begin()

	if errTx != nil{
     json.NewEncoder(w).Encode(Respond{
		Message:"Something went wrong Transaction Failed",
		Success:false,
	 })
	 
	}

	defer tx.Rollback()

}
