package booking

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	handlerconn "medquemod/db_conn"
	sidefunc "medquemod/booking/verification"
)

type (
	
	Bkservrequest struct {
		ServId       string `json:"servid"      validate:"required"`
		IntervalTime string `json:"timeInter"   validate:"required"`
		Servicename  string `json:"servicename" validate:"required"`
	}

	
	Response struct {
		Message string `json:"message"`
		Success bool   `json:"success"`
	}

	
	bkservrespond struct {
		DoctorName     string  `json:"doctorName"`
		DoctorID       int     `json:"doctorId"`
		Servicename    string  `json:"servicename"`
		StartTime      string  `json:"start_time"`
		EndTime        string  `json:"end_time"`
		DayOfWeek      string  `json:"day_of_week"`
		Date           string  `json:"date"`
		DurationMinute int     `json:"duration_minutes"`
		Fee            float64 `json:"fee"`
	}
)

func Bookinglogic(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{"Invalid method", false})
		return
	}
	w.Header().Set("Content-Type", "application/json")


	var req Bkservrequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{"Invalid payload", false})
		return
	}

	tx, err := handlerconn.Db.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{"Internal server error", false})
		return
	}
	defer tx.Rollback()


	rows, err := tx.Query(`
		SELECT 
		  dk.doctor_id,
		  dk.doctorname,
		  ds.day_of_week,
		  ds.start_time,
		  ds.end_time,
		  sa.duration_minutes,
		  sa.fee
		FROM doctors AS dk
		JOIN doctor_services AS ds ON dk.doctor_id = ds.doctor_id
		JOIN serviceAvailable AS sa ON ds.service_id = sa.serv_id
		WHERE sa.servicename = $1
	`, req.Servicename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{"Database query error", false})
		return
	}
	defer rows.Close()

	
	now := time.Now()
	allowedDates := make(map[int]string, 4)
	for i := 0; i < 4; i++ {
		d := now.Add(time.Duration(i) * 24 * time.Hour)
		allowedDates[int(d.Weekday())] = d.Format("2006-01-02")
	}
	todayInt := int(now.Weekday())
	currentTime := now.Format("15:04")

	var results []bkservrespond


	for rows.Next() {
		var (
			docID       int
			docName     string
			dayInt      int
			start, end  string
			durMins     int
			fee         float64
		)
		if err := rows.Scan(&docID, &docName, &dayInt, &start, &end, &durMins, &fee); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{Message:"Scan error", Success: false})
			return
		}

	
		dateStr, ok := allowedDates[dayInt]
		if !ok {
			continue
		}

		
		dayName, err := sidefunc.Dayofweek(dayInt)
		if err != nil {
			continue
		}

		
		slots, err := sidefunc.GenerateTimeSlote(durMins, start, end)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{"Slot generation error", false})
			return
		}

	
		for _, ts := range slots {
			if dayInt == todayInt && ts.EndTime <= currentTime {
				continue
			}

		
			var status string
			err := tx.QueryRow(`
				SELECT status 
				  FROM bookings 
				 WHERE doctor_id = $1
				   AND serv_id   = $2
				   AND day_of_week = $3
				   AND start_time  = $4
			`, docID, req.ServId, strconv.Itoa(dayInt), ts.StartTime).Scan(&status)

			if err != nil && err != sql.ErrNoRows {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(Response{"Booking lookup error", false})
				return
			}
			if status != "" && status != "cancellation" {
				continue
			}
			results = append(results, bkservrespond{
				DoctorID:       docID,
				DoctorName:     docName,
				Servicename:    req.Servicename,
				StartTime:      ts.StartTime,
				EndTime:        ts.EndTime,
				DayOfWeek:      dayName,
				Date:           dateStr,
				DurationMinute: durMins,
				Fee:            fee,
			})
		}
	}

	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{"Rows iteration error", false})
		return
	}


	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{"Failed to commit", false})
		return
	}

	
	json.NewEncoder(w).Encode(results)
}
