package booking

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	sidefunc_test "medquemod/booking/verification"
	handlerconn "medquemod/db_conn"
	"medquemod/middleware"
	"medquemod/smsendpoint"

	"golang.org/x/crypto/bcrypt"
)

type (
	Bkservrequest struct {
		ServId       int    `json:"servid"      validate:"required"`
		IntervalTime int    `json:"timeInter"   validate:"required"`
		Servicename  string `json:"servicename" validate:"required"`
	}

	Response struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Data    interface{} `json:"data,omitempty"`
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
	BKpayload struct {
		DoctorID  string `json:"doctorId"`
		ServiceID string `json:"serviceId"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Date      string `json:"date"`
		Dayofweek string `jsom:"dayofweek"`
		ForMe     bool   `json:"forme"`
		Specname  string `json:"specname"`
		Speckey   string `json:"speckey"`
	}
)

func Bookingpayload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid Method",
			Success: false,
		})
		return
	}
	claims, ok := r.Context().Value("user").(*middleware.CustomClaims)
	if !ok {
		json.NewEncoder(w).Encode(Response{
			Message: "Unathorized try to access data",
			Success: false,
		})
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println()
		return
	}
	var (
		Username, UserId, UserRole string
	)
	UserRole = claims.Role
	UserId = claims.ID
	Username = claims.Username
	client, errTx := handlerconn.Db.Begin()
	if errTx != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Internal serverError",
			Success: false,
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			client.Rollback()
		}
	}()
	defer client.Rollback()
	var phone string
	checkuserexist := client.QueryRow(`SELECT dial FROM users WHERE fullname = $1 AND user_id = $2 AND user_type = $3`, Username, UserId, UserRole).Scan(&phone)
	if checkuserexist != nil {
		if checkuserexist == sql.ErrNoRows {
			json.NewEncoder(w).Encode(Response{
				Message: "User not exists",
				Success: false,
			})
			return
		}
		// fmt.Println("Error checking user existence:", checkuserexist)
		json.NewEncoder(w).Encode(Response{
			Message: "Internal serverError",
			Success: false,
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var bkreq BKpayload

	err := json.NewDecoder(r.Body).Decode(&bkreq)
	if err != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "InterserverError",
			Success: false,
		})
		log.Printf("Something went wrong: %v", err)
		return
	}
	fmt.Println(bkreq.Dayofweek)
	fmt.Println(UserId)
	fmt.Println(Username)
	daynumber, err := sidefunc_test.DayOfWeekReverse(bkreq.Dayofweek)
	if err != nil {
		fmt.Println("Error converting dayofweek:", err)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalidpayload",
			Success: false,
		})
		return
	}

	if bkreq.ForMe {
		fmt.Println("Processing ForMe booking...")
		isExist, errorfunc := sidefunc_test.CheckalreadybookedToday(UserId, bkreq.Date, client)
		if errorfunc != nil {
			fmt.Println("Error checking if already booked:", errorfunc)
			json.NewEncoder(w).Encode(Response{
				Message: "Internal serverError",
				Success: false,
			})
			return
		}
		fmt.Println("Is already booked:", isExist)
		if !isExist {
			fmt.Println("Inserting booking for user...")
			fmt.Println("DoctorID being inserted:", bkreq.DoctorID)
			fmt.Println("ServiceID being inserted:", bkreq.ServiceID)
			fmt.Println("UserId being inserted:", UserId)

			// Convert string to int for doctor_id
			doctorIDInt, err := strconv.Atoi(bkreq.DoctorID)
			if err != nil {
				fmt.Println("Error converting doctor_id to int:", err)
				json.NewEncoder(w).Encode(Response{
					Message: "Invalid doctor ID",
					Success: false,
				})
				return
			}

			// Convert string to int for service_id
			serviceIDInt, err := strconv.Atoi(bkreq.ServiceID)
			if err != nil {
				fmt.Println("Error converting service_id to int:", err)
				json.NewEncoder(w).Encode(Response{
					Message: "Invalid service ID",
					Success: false,
				})
				return
			}

			_, errbk := client.Exec(`INSERT INTO bookingtrack_tb (user_id, doctor_id, service_id, booking_date, dayofweek, start_time, end_time, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, UserId, doctorIDInt, serviceIDInt, bkreq.Date, daynumber, bkreq.StartTime, bkreq.EndTime, "completed")
			if errbk != nil {
				fmt.Println("Error inserting booking:", errbk)
				json.NewEncoder(w).Encode(Response{
					Message: "Internal serverError",
					Success: false,
				})
				return
			}
			fmt.Println("Booking inserted successfully")
			errsms := smsendpoint.SmsEndpoint(Username, phone, bkreq.StartTime, bkreq.EndTime)
			if errsms != nil {
				fmt.Println("Error sending SMS:", errsms)
				json.NewEncoder(w).Encode(Response{
					Message: "Internal ServerError",
					Success: false,
				})
				return
			}
		} else {
			json.NewEncoder(w).Encode(Response{
				Message: "You already have a booking for today",
				Success: false,
			})
			return
		}

	} else if !bkreq.ForMe {
		fmt.Println("Processing ForSomeoneElse booking...")
		var hashedsecretkey string
		var spec_id string
		errRow := client.QueryRow(`SELECT secretkey, spec_id FROM specialgroup WHERE Username = $1 AND managedby_id = $2`, bkreq.Specname, UserId).Scan(&hashedsecretkey, &spec_id)
		if errRow != nil {
			if errRow == sql.ErrNoRows {
				fmt.Println("no such special-group entry for user", errRow)
				json.NewEncoder(w).Encode(Response{
					Message: "Invalid payload",
					Success: false,
				})
				return
			}
			fmt.Println("Error querying special group:", errRow)
			json.NewEncoder(w).Encode(Response{
				Message: "Internal serverError",
				Success: false,
			})
			return
		}
		errcomparedhash := bcrypt.CompareHashAndPassword([]byte(hashedsecretkey), []byte(bkreq.Speckey))
		if errcomparedhash != nil {
			fmt.Println("Error comparing hash:", errcomparedhash)
			json.NewEncoder(w).Encode(Response{
				Message: "Invalid payload or Internal serverError try again",
				Success: false,
			})
			return
		}
		fmt.Println("Inserting booking for someone else...")
		fmt.Println("DoctorID being inserted:", bkreq.DoctorID)
		fmt.Println("ServiceID being inserted:", bkreq.ServiceID)

		// Convert string to int for doctor_id
		doctorIDInt, err := strconv.Atoi(bkreq.DoctorID)
		if err != nil {
			fmt.Println("Error converting doctor_id to int:", err)
			json.NewEncoder(w).Encode(Response{
				Message: "Invalid doctor ID",
				Success: false,
			})
			return
		}

		// Convert string to int for service_id
		serviceIDInt, err := strconv.Atoi(bkreq.ServiceID)
		if err != nil {
			fmt.Println("Error converting service_id to int:", err)
			json.NewEncoder(w).Encode(Response{
				Message: "Invalid service ID",
				Success: false,
			})
			return
		}

		_, errexec := client.Exec(`INSERT INTO bookingtrack_tb (user_id, spec_id, doctor_id, service_id, booking_date, dayofweek, start_time, end_time, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`, UserId, spec_id, doctorIDInt, serviceIDInt, bkreq.Date, daynumber, bkreq.StartTime, bkreq.EndTime, "completed")
		if errexec != nil {
			fmt.Println("Error inserting booking for someone else:", errexec)
			json.NewEncoder(w).Encode(Response{
				Message: "Internal ServerError",
				Success: false,
			})
			return
		}
		errsms := smsendpoint.SmsEndpoint(bkreq.Specname, phone, bkreq.StartTime, bkreq.EndTime)
		if errsms != nil {
			fmt.Println("Error sending SMS for someone else:", errsms)
			json.NewEncoder(w).Encode(Response{
				Message: "InternetError",
				Success: false,
			})
			return
		}

	}
	fmt.Println("Committing transaction...")
	if err := client.Commit(); err != nil {
		fmt.Println("Error committing transaction:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "Failed to commit", Success: false})
		return
	}
	fmt.Println("Transaction committed successfully")

	var bookingID int
	err = handlerconn.Db.QueryRow(`SELECT id FROM bookingtrack_tb WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, UserId).Scan(&bookingID)
	if err == nil {
		_, errNotif := handlerconn.Db.Exec(`INSERT INTO scheduled_notifications (booking_id, status) VALUES ($1, 'pending')`, bookingID)
		if errNotif != nil {
			log.Printf("Failed to insert scheduled notification: %v", errNotif)
		} else {
			log.Printf("Scheduled notification created for booking ID: %d", bookingID)
		}
	} else {
		log.Printf("Failed to get booking id for notification: %v", err)
	}

	json.NewEncoder(w).Encode(Response{
		Message: "Successfully make booking",
		Success: true,
	})
}
func Bookinglogic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{Message: "Invalid method", Success: false})
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req Bkservrequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Message: "Invalid payload", Success: false})
		return
	}

	tx, err := handlerconn.Db.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "Internal server error", Success: false})
		return
	}
	defer tx.Rollback()

	// Fetch all schedule rows first
	rows, err := tx.Query(`
		SELECT 
		  dk.doctor_id,
		  dk.doctorname,
		  ds.day_of_week,
		  ds.start_time,
		  ds.end_time,
		  sa.duration_minutes,
		  sa.fee
		FROM serviceAvailable AS sa
		JOIN doctor_services AS dsv ON sa.serv_id = dsv.service_id
		JOIN doctors AS dk ON dk.doctor_id = dsv.doctor_id 
		JOIN doctorshedule AS ds ON ds.doctor_id = dk.doctor_id
		WHERE sa.serv_id = $1
	`, req.ServId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "Database query error", Success: false})
		return
	}

	// Store schedules in memory and close rows immediately
	type schedule struct {
		docID   int
		docName string
		dayInt  int
		start   string
		end     string
		durMins int
		fee     float64
	}
	var schedules []schedule
	for rows.Next() {
		var s schedule
		if err := rows.Scan(&s.docID, &s.docName, &s.dayInt, &s.start, &s.end, &s.durMins, &s.fee); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{Message: "Scan error", Success: false})
			rows.Close()
			return
		}
		schedules = append(schedules, s)
	}
	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "Rows error", Success: false})
		rows.Close()
		return
	}
	rows.Close() // Free transaction for reuse

	now := time.Now()
	allowedDates := make(map[int]string, 4)
	for i := 0; i < 4; i++ {
		d := now.AddDate(0, 0, i)
		allowedDates[int(d.Weekday())] = d.Format("2006-01-02")
	}
	todayInt := int(now.Weekday())
	currentTime := now.Format("15:04")

	var results []bkservrespond

	// Process stored schedules
	for _, s := range schedules {
		dateStr, ok := allowedDates[s.dayInt]
		if !ok {
			continue
		}

		dayName, err := sidefunc_test.Dayofweek(s.dayInt)
		if err != nil {
			continue
		}

		layout := "15:04"
		sT := s.start[11:16]
		eT := s.end[11:16]
		S, _ := time.Parse(layout, sT)
		E, _ := time.Parse(layout, eT)
		startTime := S.Format("15:04")
		endTime := E.Format("15:04")

		slots, err := sidefunc_test.GenerateTimeSlote(s.durMins, startTime, endTime)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{Message: "Slot generation error", Success: false})
			return
		}

		for _, ts := range slots {
			if s.dayInt == todayInt && ts.EndTime <= currentTime {
				continue
			}

			// Check slot availability in separate query
			var status string
			err := tx.QueryRow(`
				SELECT status
				FROM bookingTrack_tb
				WHERE doctor_id   = $1
				  AND service_id    = $2
				  AND dayofweek = $3
				  AND start_time  = $4
			`, s.docID, req.ServId, s.dayInt, ts.StartTime).Scan(&status)

			if err != nil && err != sql.ErrNoRows {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(Response{Message: "Booking lookup error", Success: false})
				log.Println("booking query failed:", err)
				return
			}

			if status != "" && status != "cancellation" {
				continue
			}

			results = append(results, bkservrespond{
				DoctorID:       s.docID,
				DoctorName:     s.docName,
				Servicename:    req.Servicename,
				StartTime:      ts.StartTime,
				EndTime:        ts.EndTime,
				DayOfWeek:      dayName,
				Date:           dateStr,
				DurationMinute: s.durMins,
				Fee:            s.fee,
			})
		}
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "Failed to commit", Success: false})
		return
	}

	json.NewEncoder(w).Encode(Response{
		Message: "Successfully retrieved slots",
		Success: true,
		Data:    results,
	})
}

// Use built-in Expo push notification logic
// This function triggers the built-in notification worker to process and send pending notifications
func TriggerExpoPushNotifications() {
	// checkPendingNotifications is defined in notify.go and uses SendNotification
	ManualTriggerNotifications()
}
