package api

import (
	"database/sql"
	"encoding/json"
	"log"
	handlerconn "medquemod/db_conn"
	"net/http"
	"strings"
	"time"
)

type (
	Response struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
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
		Specialty    string `json:"specialty"`
		TimeInterval string `json:"timeinterval"`
		Rating       string `json:"rating"`
		IsAvailable  bool   `json:"isAvailable"`
	}
)

func Doctors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader((http.StatusMethodNotAllowed))
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
		log.Printf("failed to start transaction %v", errTx)
		return
	}
	defer tx.Rollback()
	rows, err := tx.Query(query)

	if err != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Something went wrong here",
			Success: false,
		})
		log.Printf("Something went went failed to execute query %v", err)
		return
	}
	defer rows.Close()
	loc, err := time.LoadLocation("Local")
	if err != nil {
		log.Printf("something went wrong: %v", err)
	}
	now := time.Now().In(loc)
	var doctors []doctorInfo

	for rows.Next() {
		var name, specialty, timeInterval, rating string

		err := rows.Scan(&name, &specialty, &timeInterval, &rating)
		if err != nil {
			log.Printf("something went wrong %v", &name)
			continue
		}
		SplitInterval := strings.SplitN(timeInterval, "-",2)
		if len(SplitInterval) != 2 {
			log.Printf("something went wrong here %v", SplitInterval)
			continue
		}

		start := strings.TrimSpace(SplitInterval[0])
		end := strings.TrimSpace(SplitInterval[1])

		currentTime := now.Format("03:04 PM")
		startTime, err1 := time.ParseInLocation("03:04 PM", start, loc)
		endTime, err2 := time.ParseInLocation("03:04 PM", end, loc)
		nowParser, _ := time.ParseInLocation("03:04 PM", currentTime, loc)

		if err1 != nil || err2 != nil {
			log.Printf("failed to ParserLocation:%v, %v ", err1, err2)
			continue
		}
		// IsAvailable := nowParser.After(startTime) && nowParser.Before(endTime)
		isAvailable := (nowParser.Equal(startTime) || nowParser.After(startTime) && nowParser.Equal(endTime)|| nowParser.Before(endTime))
		doctors = append(doctors, doctorInfo{
			Fullname:     name,
			Specialty:    specialty,
			TimeInterval: timeInterval,
			IsAvailable:  isAvailable,
			Rating:       rating,
		})
	}
	if err := rows.Err(); err !=nil{
		log.Printf("row iteration failed: %v", err)
	}
	if errcommit := tx.Commit(); errcommit != nil {
		log.Printf("something went wrong failed to commit transaction")
	}
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: "Succesfuly",
		Data:    doctors,
	})
}

func Verifyuser(w http.ResponseWriter, r *http.Request) {
    // 1) Always set JSON header up front
    w.Header().Set("Content-Type", "application/json")

    
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        json.NewEncoder(w).Encode(Response{
            Message: "Method not allowed",
            Success: false,
        })
        return                                  // ← return after Encode
    }

    
    var payload struct {
        DeviceId string `json:"deviceId"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        w.WriteHeader(http.StatusBadRequest)     
        json.NewEncoder(w).Encode(Response{
            Message: "Invalid JSON payload",
            Success: false,
        })
        return                                  
    }


    tx, err := handlerconn.Db.Begin()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(Response{
            Message: "Transaction failed to start",
            Success: false,
        })
        return
    }
    defer tx.Rollback()

    
    var checkDeviceID string
    err = tx.
        QueryRow("SELECT deviceId FROM Users WHERE deviceId = $1", payload.DeviceId).
        Scan(&checkDeviceID)


    if err == sql.ErrNoRows {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(Response{
            Message: "Not yet Registered",
            Success: true,
            Data: struct {
                UserExist bool `json:"user_exist"`
            }{false},
        })
        return                                  
    }
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(Response{
            Message: "Database query error",
            Success: false,
        })
        return
    }

    if err := tx.Commit(); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(Response{
            Message: "Transaction failed to commit",
            Success: false,
        })
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(Response{
        Message: "Success",
        Success: true,
        Data: struct {
            UserExist bool `json:"user_exist"`
        }{true},
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
 func getService(w http.ResponseWriter, r*http.Request){
       if r.Method != http.MethodGet{
		w .WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Bad request Method isn`t allowed",
			Success: false,
		})
		return
	   }
	   w.Header().Set("Content-type", "application/json")
 }