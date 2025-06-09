package booking

import (
	"encoding/json"
	"fmt"
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
	 if err :=  json.NewDecoder(r.Body).Decode(&bsr)
	row := tx.QueryRow(` SELECT dk.doctor_id, dk.doctorname,ds.day_of_week,ds.start_time, ds.end_time,sa.servicename,sa.duration_minutes,sa.fee FROM doctors AS dk JOIN doctor_services AS ds ON dk.doctor_id = ds.doctor_id JOIN serviceAvailable AS sa ON ds.service_id = sa.serv_id WHERE sa.servicename = $1`, bsr.ServiceName) 

err := row.Scan(
    &bsr.DoctorID,
    &bsr.DoctorName,
    &bsr.DayOfWeek,
    &bsr.StartTime,
    &bsr.EndTime,
    &bsr.ServiceName,
    &bsr.DurationMinutes,
    &bsr.Fee,
)
if err != nil {
	w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(Response{
		Message: "InternaserverError ",
		Success: false,
	})
    return fmt.Errorf("Soemthing went wrong", err)
}


}
