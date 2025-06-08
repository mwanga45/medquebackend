package booking

import (
	"encoding/json"
	handlerconn "medquemod/db_conn"
	"net/http"
)

type (
	Bkservrequest struct {
		ServId       string `json:"servid" validate:"required"`
		IntervalTime string `json:"timeInter" validate:"required"`
		Servicename  string `json:"servicename" validate:"required"`
	
	}
	Response struct{
		Message string
		Success  bool
	}
	bkservrespond struct{
		Doctor string
		doctorId string
		Servicename string
		Start_time string
		End_time string
		Dayofweek string
		fee string
	}

)
func ReturnSlot (w http.ResponseWriter, r* http.Request){

}
func bookinglogic(w http.ResponseWriter, r *http.Request) {
     if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid payload",
			Success: false,
		})
		return
	 }
	 w.Header().Set("Content-type", "application/json")

	 var bsr Bkservrequest
	 tx,errTx := handlerconn.Db.Begin() 
	 if errTx != nil{
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "InternaServerError",
			Success: false,
		})
		return
	 }
	 
	 

}
