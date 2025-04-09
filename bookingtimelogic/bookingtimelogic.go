package bookingtimelogic

import (
	"encoding/json"
	"fmt"
	"time"

	// handlerconn "medquemod/db_conn"
	"net/http"
)

type (
	Respond struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Tslot   interface{} `json:"tslot"`
		Dslot   interface{} `json:dslot`
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
	Dayslote, err := DayInterval()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Respond{
			Message: "Something went wrong",
			Success: false,
		})
	}
	Timeslots, err := Timeslot()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Respond{
			Message: "Something went wrong here",
			Success: false,
		})
	}

	json.NewEncoder(w).Encode(Respond{
		Message: "successfuly",
		Success: true,
		Tslot:   Timeslots,
		Dslot:   Dayslote,
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
func Timeslot() (interface{}, error) {
	// Create an slice to hold time slote
	var Timeslote []map[string]interface{}
	timelayout := "15:04"
	startTime, errst := time.Parse(timelayout, "06:30")

	if errst != nil {
		return "", fmt.Errorf("something went wrong failed to format time")
	}

	endTime, erred := time.Parse(timelayout, "20:30")

	if erred != nil {
		return "", fmt.Errorf("something went wrong failed to format time")
	}

	now := time.Now()
	startTime = time.Date(now.Year(), now.Month(), now.Day(), startTime.Hour(), startTime.Minute(), 0, 0, now.Location())
	endTime = time.Date(now.Year(), now.Month(), now.Day(), endTime.Hour(), endTime.Minute(), 0, 0, now.Location())
	intervalMinutes := 10 * time.Minute

	for t := startTime; t.Before(endTime); t = t.Add(intervalMinutes) {
		// slote to carry  time slote per each iteretion
		timeslote := map[string]interface{}{
			"time": t.Format("03:04 PM"),
		}
		Timeslote = append(Timeslote, timeslote)
	}
	return Timeslote, nil
}
