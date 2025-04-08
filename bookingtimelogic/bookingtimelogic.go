package bookingtimelogic

import (
	"encoding/json"
	"strconv"
	"time"

	// handlerconn "medquemod/db_conn"
	"net/http"
)

type (
	DayofToday struct {
		Day string `json:"day" validate:"required"`
		Date int
	}
	Respond struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
	}
)

func Timelogic(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Respond{
			Message: "Bad request",
			Success: false,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	// tx,errTx  := handlerconn.Db.Begin()

	// if errTx != nil{
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	json.NewEncoder(w).Encode(Respond{
	// 		Message: "something went wrong here ",
	// 		Success: false,
	// 	})
	// 	return
	// }
}

func DayInterval(){
	// create an object that will carry current time
	currentDate  := time.Now()
	IntervalDate := 1
	looptime := 4
	// create an  map slice to hold the date interval 
	var Timeslot []map[string]interface {}
	// create loop iteration for the time available to make booking

	for i:= 0; i<=looptime; i++{
     NextTime := currentDate.AddDate(0,0,IntervalDate)
	//  create an slice to hold time slot per each iteration 
	timeslot := map[string]interface{}{
		"From":currentDate.Format("2/1/2006"),
		"To":NextTime.Format("2/1/2006"),
	}
	

	}
}

