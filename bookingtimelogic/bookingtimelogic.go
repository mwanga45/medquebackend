package bookingtimelogic

import (
	"encoding/json"
	"time"

	// handlerconn "medquemod/db_conn"
	"net/http"
)
type (
	Respond struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
	}
)
func Timelogic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Respond{
			Message: "Bad request",
			Success: false,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")

	// return days interval
	Timeslote, err := DayInterval()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Respond{
			Message: "Something went wrong",
			Success: false,
		})
	}
	json.NewEncoder(w).Encode(Respond{
		Message: "successfuly",
		Success: true,
		Data:    Timeslote,
	})
}
func DayInterval() (interface{}, error) {
	// create an object that will carry current time
	currentDate := time.Now()
	IntervalDate := 1
	looptime := 4
	// create an  map slice to hold the date interval
	var Timeslot []map[string]interface{}

	// create an layout for date
    layoutDate := "2/1/2006"

	// create loop iteration for the time available to make booking
	for i := 0; i <= looptime; i++ {
		NextTime := currentDate.AddDate(0, 0, IntervalDate)

		//  create an slice to hold time slot per each iteration
		timeslot := map[string]interface{}{
			"From": currentDate.Format(layoutDate),
			"To":   NextTime.Format(layoutDate),
		}
		currentDate = NextTime
		Timeslot = append(Timeslot, timeslot)
	}
	return Timeslot, nil
}
func Timeslot()(interface{},error)  {
	timelayout := "15:04"
   startTime,_ := time.Parse(timelayout, "06:30")
   endTime,_ := time.Parse(timelayout, "20:30")

   now := time.Now()
   
	
}