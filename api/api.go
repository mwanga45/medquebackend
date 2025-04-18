package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	handlerconn "medquemod/db_conn"
	"net/http"
	"strings"
	"time"
)

type (
	Response struct {
		Message string      `json:"message"`
		Success bool        `json:"success,omitempty"`
		Data    interface{} `json:"data"`
	}
	DeviceUid struct {
		DeviceId string `json:"deviceId"`
	}
	User_details struct {
		Name         string `json:"name"`
		Email        string `json:"email"`
		Phone_num    string `json:"phone_numb"`
		DeviceId     string `json:"deviceId"`
		Home_address string `json:"home_address,omitempty"`
		Age          string `json:"age"`
	}
	Verfiy_user struct {
		User_exist bool `json:"user_exist"`
	}
	doctorInfo struct {
		Fullname     string `json:"fullname"`
		Specialty    string `json:"speciality"`
		TimeInterval string `json:"timeinterval"`
		Rating       string `json:"rating"`
		IsAvailable  bool   `json:"isAvailable"`
	}
)

func Doctors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader((http.StatusBadRequest))
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid method used",
			Success: false,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")

	query := "SELECT full_name, specialty, time_available, rating FROM doctor_status"

	tx, errTx := handlerconn.Db.Begin()

	if errTx != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "failed to start Transaction",
			Success: false,
		})
		fmt.Errorf("failed to start transaction %v", errTx)
		return
	}
	if errcommit := tx.Commit(); errcommit != nil {
		fmt.Errorf("something went wrong failed to commit transaction")
	}
	rows, err := tx.Query(query)

	if err != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Something went wrong here",
			Success: false,
		})
		fmt.Errorf("Something went went failed to execute query %v", err)
		return
	}
	loc, err := time.LoadLocation("Local")
	if err !=nil{
		fmt.Errorf("something went wrong")
	}
	now := time.Now().In(loc)
	var doctors []doctorInfo

	for rows.Next() {
		var name, specialty, timeInterval, rating string

		err := rows.Scan(&name, &specialty, &timeInterval, &rating)
		if err != nil {
			fmt.Errorf("something went wrong %s", &name, err)
			continue
		}
		SplitInterval := strings.Split(timeInterval, "-")
		if len(SplitInterval) != 2 {
			fmt.Errorf("something went wrong here %v", SplitInterval)
		}

		start := strings.TrimSpace(SplitInterval[0])
		end := strings.TrimSpace(SplitInterval[1])

		currentTime := now.Format("03:04 PM")
		startTime, err1 := time.ParseInLocation("03:04 PM", start, loc)
		endTime, err2 := time.ParseInLocation("03:04 PM", end, loc)
		nowParser, _ := time.ParseInLocation("03:04 PM", currentTime, loc)

		if err1 != nil || err2 != nil {
			fmt.Errorf("failed to ParserLocation", err1, err2)
			continue
		}
		IsAvailable := nowParser.After(startTime) && nowParser.Before(endTime)
		doctors = append(doctors, doctorInfo{
			Fullname:     name,
			Specialty:    specialty,
			TimeInterval: timeInterval,
			IsAvailable:  IsAvailable,
			Rating:       rating,
		})
	}
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Message: "Succesfuly",
		Data: doctors,
	})
}

// func Doctors(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodGet {
// 		w.WriteHeader(http.StatusMethodNotAllowed)
// 		json.NewEncoder(w).Encode(Response{
// 			Message: "Invalid method used ",
// 			Success: false,
// 		})
// 		return
// 	}
// 	tx, errTx := handlerconn.Db.Begin()
// 	if errTx != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(Response{
// 			Message: "Transaction Failed",
// 		})
// 		return
// 	}
// 	defer tx.Rollback()

// 	rows, err := tx.Query("SELECT * FROM doctors")

// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(Response{
// 			Message: "database failed to fetch data",
// 			Success: false,
// 		})
// 		return
// 	}
// 	defer rows.Close()
// 	// create are slice which will hold the row as value and column name as key at end
// 	var doctors []map[string]interface{}
// 	// store columns name variable
// 	columns, _ := rows.Columns()
// 	count := len(columns)
// 	// create slice for store temporarly  column as key and data as value with corresponding data type of column
// 	value := make([]interface{}, count)
// 	// create  slice pointer neccessary during scan value data to corresponding column name
// 	valueptrs := make([]interface{}, count)
// 	// automatic move cursor to next rows
// 	for rows.Next() {
// 		for i := range columns {
// 			valueptrs[i] = &value[i]
// 		}
// 		rows.Scan(valueptrs...)
// 		// create another slice that will hold  single row data as value and also column name  as key after each loop
// 		doctor := make(map[string]interface{})
// 		for i, col := range columns {
// 			val := value[i]
// 			doctor[col] = val
// 		}
// 		doctors = append(doctors, doctor)
// 	}
// 	// handling rows error
// 	if err := rows.Err(); err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(Response{
// 			Message: "data failed to be proccessed",
// 			Success: false,
// 		})
// 		return
// 	}
//     if err := tx.Commit();err !=nil{
//        json.NewEncoder(w).Encode(Response{
// 		Message: "Failed to commit transaction",
// 		Success: false,
// 	   })
// 	   return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	err = json.NewEncoder(w).Encode(Response{
// 		Message: "successfuly fetch data",
// 		Success: true,
// 		Data:    doctors,
// 	})
// 	// handling encoding process if error occur
// 	if err != nil {
// 		json.NewEncoder(w).Encode(Response{
// 			Message: "failed to encode data ",
// 			Success: false,
// 		})
// 		return
// 	}

// }
func Verifyuser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Println("Invalid method used")
		return
	}
	tx, errTx := handlerconn.Db.Begin()
	if errTx != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Transaction failed",
			Success: true,
		})
		return
	}
	defer tx.Rollback()
	var deviceId DeviceUid
	query := "SELECT deviceId FROM Users WHERE deviceId = $1"
	err := json.NewDecoder(r.Body).Decode(&deviceId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "failed to decode",
			Success: false,
		})
		return
	}
	var verify Verfiy_user
	var check_deviceid string
	err = tx.QueryRow(query, deviceId.DeviceId).Scan(&check_deviceid)
	if err != nil {
		if err == sql.ErrNoRows {
			verify = Verfiy_user{User_exist: false}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Response{
				Message: "Not yet Registered",
				Success: true,
				Data:    verify,
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		verify = Verfiy_user{User_exist: true}
	}
	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Transaction failed to commit",
			Success: false,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Message: "success",
		Data:    verify,
		Success: true,
	})
}

func Userdetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	tx, errTx := handlerconn.Db.Begin()
	if errTx != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "transaction failed",
			Success: false,
		})
		return
	}
	defer tx.Rollback()

	// check if the deviceId is Available for this
	query := "SELECT full_name,email,home_address,phone_number,Age,deviceId FROM Users WHERE deviceId = $1"
	var dvId DeviceUid
	err := json.NewDecoder(r.Body).Decode(&dvId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	var details User_details
	err = tx.QueryRow(query, dvId.DeviceId).Scan(&details.Name, &details.Email, &details.Home_address, &details.Phone_num, &details.Age, &details.DeviceId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed to fetch value")
		return
	}
	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "something went wrong Transaction faild to commit",
			Success: false,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(Response{
		Message: "Successfully ",
		Success: true,
		Data:    details,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
