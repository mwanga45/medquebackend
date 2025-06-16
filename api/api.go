package api

import (
	"database/sql"
	"encoding/json"
	"log"
	sidefunc_test "medquemod/booking/verification"
	handlerconn "medquemod/db_conn"
	"net/http"
	"time"
)

type (
	Response struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Data    interface{} `json:"data,omitempty"`
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
		Fullname    string `json:"fullname"`
		Specialist   string `json:"specialty"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
		Dayofweek   string `json:"dayofweek"`
		IsAvailable bool   `json:"isAvailable"`
	}
	Service struct {
		ID                 int     `json:"id"`
		Servicename           string  `json:"servicename"`
		ConsultationFee    float64 `json:"consultationFee"`
		CreatedAt  string `json:"created_at"`
		DurationMin int `json:"duration_minutes"`
	}
)
func DoctorsAvailability(w http.ResponseWriter, r *http.Request) {

    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        json.NewEncoder(w).Encode(Response{
            Message: "Invalid method; only GET allowed",
            Success: false,
        })
        return
    }

    w.Header().Set("Content-Type", "application/json")

    query := `
        SELECT d.doctorname, d.specialist_name, ds.start_time, ds.end_time, ds.day_of_week
        FROM doctors AS d
        JOIN doctorschedule AS ds
          ON ds.doctor_id = d.doctor_id
    `

    tx, err := handlerconn.Db.Begin()
    if err != nil {
        log.Printf("failed to start transaction: %v", err)
        json.NewEncoder(w).Encode(Response{"failed to start transaction", false, nil})
        return
    }
    defer tx.Rollback()

    rows, err := tx.Query(query)
    if err != nil {
        log.Printf("query failed: %v", err)
        json.NewEncoder(w).Encode(Response{"query execution failed", false, nil})
        return
    }
    defer rows.Close()


    now := time.Now()

    var doctors []doctorInfo
    for rows.Next() {
        var (
            name, spec, startStr, endStr string
            dow                            int
        )
        if err := rows.Scan(&name, &spec, &startStr, &endStr, &dow); err != nil {
            log.Printf("row scan error: %v", err)
            continue
        }

    
        weekday, err := sidefunc_test.Dayofweek(dow)
        if err != nil {
            log.Printf("weekday conversion error: %v", err)
            weekday = "unknown"
        }

        
        layout := "03:04 PM"
        startTime, err1 := time.ParseInLocation(layout, startStr, now.Location())
        endTime, err2 := time.ParseInLocation(layout, endStr, now.Location())
        if err1 != nil || err2 != nil {
            log.Printf("time parse error: %v / %v", err1, err2)
            continue
        }

        
        isAvail := !now.Before(startTime) && !now.After(endTime)
        doctors = append(doctors, doctorInfo{
            Fullname:    name,
            Specialist:  spec,
            StartTime:   startStr,
            EndTime:     endStr,
            Dayofweek:   weekday,
            IsAvailable: isAvail,
        })
    }
    if err := rows.Err(); err != nil {
        log.Printf("row iteration error: %v", err)
    }

    if err := tx.Commit(); err != nil {
        log.Printf("transaction commit failed: %v", err)
    }

    json.NewEncoder(w).Encode(Response{
        Message: "Successfully fetched availability",
        Success: true,
        Data:    doctors,
    })
}

func Verifyuser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Method not allowed",
			Success: false,
		})
		return
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

// RETURN ALL  THE SERVICE AVALAIBLE
func GetService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Method not allowed",
			Success: false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	rows, err := handlerconn.Db.Query("SELECT serv_id, servicename, duration_minutes, fee, created_at FROM  serviceAvailable")
	if err != nil {
		log.Printf("Database query error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Server Error: Failed to fetch data",
			Success: false,
		})
		return
	}
	defer rows.Close()

	var services []Service

	for rows.Next() {
		var service Service
		err := rows.Scan(
			&service.ID,
			&service.Servicename,
            &service.DurationMin,
			&service.ConsultationFee,
			&service.CreatedAt,
		)
		if err != nil {
			log.Printf("Row scanning error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{
				Message: "Server Error: Failed to process data",
				Success: false,
			})
			return
		}
		services = append(services, service)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Server Error: Data processing failed",
			Success: false,
		})
		return
	}

	json.NewEncoder(w).Encode(Response{
		Message: "Successfully returned data",
		Success: true,
		Data:    services,
	})
}
