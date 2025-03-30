package api

import (
	"database/sql"
	"encoding/json"
	"log"
	handlerconn "medquemod/db_conn"
	"net/http"
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
)

func Doctors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid method used ",
			Success: false,
		})
		return
	}

	rows, err := handlerconn.Db.Query("SELECT * FROM doctors")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "database failed to fetch data",
			Success: false,
		})
		return
	}
	defer rows.Close()
	// create are slice which will hold the row as value and column name as key at end
	var doctors []map[string]interface{}
	// store columns name variable
	columns, _ := rows.Columns()
	count := len(columns)
	// create slice for store temporarly  column as key and data as value with corresponding data type of column
	value := make([]interface{}, count)
	// create  slice pointer neccessary during scan value data to corresponding column name
	valueptrs := make([]interface{}, count)
	// automatic move cursor to next rows
	for rows.Next() {
		for i := range columns {
			valueptrs[i] = &value[i]
		}
		rows.Scan(valueptrs...)
		// create another slice that will hold  single row data as value and also column name  as key after each loop
		doctor := make(map[string]interface{})
		for i, col := range columns {
			val := value[i]
			doctor[col] = val
		}
		doctors = append(doctors, doctor)
	}
	// handling rows error
	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "data failed to be proccessed",
			Success: false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(Response{
		Message: "successfuly fetch data",
		Success: true,
		Data:    doctors,
	})
	// handling encoding process if error occur
	if err != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "failed to encode data ",
			Success: false,
		})
		return
	}
}
func Verifyuser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Println("Invalid method used")
		return
	}
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
	err = handlerconn.Db.QueryRow(query,deviceId.DeviceId).Scan(&check_deviceid)
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
	err = handlerconn.Db.QueryRow(query, dvId.DeviceId).Scan(&details.Name, &details.Email, &details.Home_address, &details.Phone_num, &details.Age, &details.DeviceId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed to fetch value")
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
func BookingList(w http.ResponseWriter, r *http.Request) {

}
func BandlebookingTime(w http.ResponseWriter, r *http.Request) {

}
