package booking

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	sidefunc_test "medquemod/booking/verification"
	handlerconn "medquemod/db_conn"
	"medquemod/middleware"
	"medquemod/smsendpoint"
	"medquemod/types"
	"medquemod/utils"

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
	Payload struct {
		SenderID   int        `json:"sender_id"`
		Schedule   string     `json:"schedule"`
		Sms        string     `json:"sms"`
		Recipients []Receiver `json:"recipients"`
	}

	Receiver struct {
		Number string `json:"number"`
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

			doctorIDInt, err := strconv.Atoi(bkreq.DoctorID)
			if err != nil {
				fmt.Println("Error converting doctor_id to int:", err)
				json.NewEncoder(w).Encode(Response{
					Message: "Invalid doctor ID",
					Success: false,
				})
				return
			}

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
			phone_number := strings.Replace(phone, "+", "", -1)
			errsms := smsendpoint.SmsEndpoint(Username, phone_number, bkreq.StartTime, bkreq.EndTime)
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

		doctorIDInt, err := strconv.Atoi(bkreq.DoctorID)
		if err != nil {
			fmt.Println("Error converting doctor_id to int:", err)
			json.NewEncoder(w).Encode(Response{
				Message: "Invalid doctor ID",
				Success: false,
			})
			return
		}

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
	rows.Close()

	now := time.Now()
	allowedDates := make(map[int]string, 4)
	for i := 0; i < 4; i++ {
		d := now.AddDate(0, 0, i)
		allowedDates[int(d.Weekday())] = d.Format("2006-01-02")
	}
	todayInt := int(now.Weekday())
	currentTime := now.Format("15:04")

	var results []bkservrespond

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
	fmt.Print(results)

	json.NewEncoder(w).Encode(Response{
		Message: "Successfully retrieved slots",
		Success: true,
		Data:    results,
	})
}

func TriggerExpoPushNotifications() {

	ManualTriggerNotifications()
}
func CancelBooking(w http.ResponseWriter, r *http.Request) {
	const StatusCancelled = "cancelled"

	fmt.Println("CancelBooking: received request")

	if r.Method != http.MethodPost {
		fmt.Println("CancelBooking: invalid method")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{Message: "Invalid method", Success: false})
		return
	}

	claims, ok := r.Context().Value("user").(*middleware.CustomClaims)
	if !ok {
		fmt.Println("CancelBooking: unauthorized - no user claims")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Response{Message: "Unauthorized", Success: false})
		return
	}
	fmt.Printf("CancelBooking: user ID from claims: %s\n", claims.ID)

	var req struct {
		BookingID int `json:"booking_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println("CancelBooking: invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Message: "Invalid payload", Success: false})
		return
	}
	fmt.Printf("CancelBooking: booking ID to cancel: %d\n", req.BookingID)

	var userID, status, bookingDate, startTime, endTime string
	var serviceID, doctorID int

	err := handlerconn.Db.QueryRow(`
		SELECT user_id, status, booking_date, start_time, end_time, service_id, doctor_id 
		FROM bookingtrack_tb WHERE id = $1`, req.BookingID).Scan(&userID, &status, &bookingDate, &startTime, &endTime, &serviceID, &doctorID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("CancelBooking: booking not found")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(Response{Message: "Booking not found", Success: false})
			return
		}
		fmt.Printf("CancelBooking: DB error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "Database error", Success: false})
		return
	}

	fmt.Printf("CancelBooking: booking found, status: %s, userID: %s\n", status, userID)

	if userID != claims.ID {
		fmt.Println("CancelBooking: forbidden - user mismatch")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(Response{Message: "You are not allowed to cancel this booking", Success: false})
		return
	}

	if status == StatusCancelled {
		fmt.Println("CancelBooking: already cancelled")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Message: "Booking is already cancelled", Success: false})
		return
	}

	_, err = handlerconn.Db.Exec(`UPDATE bookingtrack_tb SET status = $1 WHERE id = $2`, StatusCancelled, req.BookingID)
	if err != nil {
		fmt.Printf("CancelBooking: failed to update booking status: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "Failed to cancel booking", Success: false})
		return
	}
	fmt.Println("CancelBooking: booking marked as cancelled")

	extractTime := func(t string) string {
		parsed, err := time.Parse(time.RFC3339, t)
		if err == nil {
			return parsed.Format("15:04")
		}
		if len(t) >= 5 && t[2] == ':' {
			return t[:5]
		}
		return t
	}

	// Get service name
	var serviceName string
	err = handlerconn.Db.QueryRow(`
		SELECT servicename FROM serviceavailable WHERE serv_id = $1`, serviceID).Scan(&serviceName)
	if err != nil {
		fmt.Printf("CancelBooking: failed to get service name: %v\n", err)
		serviceName = "Medical Service"
	}
	fmt.Printf("CancelBooking: service name: %s\n", serviceName)

	// Notify cancelling user
	var username, phone string
	err = handlerconn.Db.QueryRow(`
		SELECT fullname, dial FROM users WHERE user_id = $1`, userID).Scan(&username, &phone)
	if err == nil {
		fmt.Printf("CancelBooking: found user %s with phone %s\n", username, phone)
		startTimeOnly := extractTime(startTime)
		endTimeOnly := extractTime(endTime)
		cleanPhone := strings.Replace(phone, "+", "", -1)

		err = smsendpoint.SmsBookingCancellationInform(username, serviceName, startTimeOnly, endTimeOnly, cleanPhone)
		if err != nil {
			fmt.Printf("CancelBooking: failed to send cancellation SMS: %v\n", err)
		} else {
			fmt.Println("CancelBooking: cancellation SMS sent successfully")
		}
	} else {
		fmt.Printf("CancelBooking: failed to fetch user details for SMS: %v\n", err)
	}

	// Notify other users
	now := time.Now()
	fmt.Printf("CancelBooking: current time for comparison: %s\n", now.Format("15:04:05"))
	currentTime := now.Format("15:04:05")


	rows, err := handlerconn.Db.Query(`
		SELECT DISTINCT u.fullname, u.dial 
		FROM bookingtrack_tb b
		JOIN users u ON b.user_id = u.user_id
		WHERE b.booking_date = $1 
		AND b.service_id = $2 
		AND b.status != $3
		AND b.user_id != $4
		AND (b.booking_date::date > $1::date OR (b.booking_date = $1 AND b.start_time::time >= $5::time))`,
		bookingDate, serviceID, StatusCancelled, userID, currentTime)

	if err != nil {
		fmt.Printf("CancelBooking: error querying other users: %v\n", err)
	} else {
		defer rows.Close()
		fmt.Print("CancelBooking: notifying other users about available slot for %s\n", rows)
		for rows.Next() {
			var otherUsername, otherPhone string
			if err := rows.Scan(&otherUsername, &otherPhone); err != nil {
				fmt.Printf("CancelBooking: error scanning other user: %v\n", err)
				continue
			}
			cleanPhone := strings.Replace(otherPhone, "+", "", -1)
			fmt.Printf("CancelBooking: notifying %s at %s\n", otherUsername, cleanPhone)
			err = smsendpoint.SmsSlotAvailableInform(otherUsername, serviceName, cleanPhone)
			if err != nil {
				fmt.Printf("CancelBooking: failed to send available slot SMS to %s: %v\n", cleanPhone, err)
			} else {
				fmt.Printf("CancelBooking: notified %s of availability\n", cleanPhone)
			}
		}
		if err := rows.Err(); err != nil {
			fmt.Printf("CancelBooking: error after iterating rows: %v\n", err)
		}
	}

	fmt.Println("CancelBooking: completed successfully")
	json.NewEncoder(w).Encode(Response{
		Message: "Booking cancelled successfully",
		Success: true,
	})
}


func SmsBookingCancellationInform(username, servicename, start_time, end_time, phoneNumber string) error {
	message := fmt.Sprintf(
		"Hi %s! Your booking for %s (slot: %s-%s) has been CANCELLED. "+
			"Contact us for assistance.",
		username, servicename, start_time, end_time,
	)

	payload := types.SmsPayload{
		SenderID: 55,
		Schedule: "none",
		Sms:      message,
		Recipients: []types.SmsReceiver{{
			Number: phoneNumber,
		}},
	}
	return utils.SendSms(payload)
}

func SmsSlotAvailableInform(username, servicename, phoneNumber string) error {
	message := fmt.Sprintf(
		"Hi %s! A time slot has become available for %s today. "+
			"If you'd like to change your appointment to an earlier time, "+
			"please visit our app now.",
		username, servicename,
	)

	payload := types.SmsPayload{
		SenderID: 55,
		Schedule: "none",
		Sms:      message,
		Recipients: []types.SmsReceiver{{
			Number: phoneNumber,
		}},
	}
	return utils.SendSms(payload)
}

func SendCancellationNotifications(bookingDate, startTime, endTime string, serviceID, doctorID int) {

	query := `
		SELECT DISTINCT u.fullname, u.deviceid, u.user_id
		FROM users u
		JOIN bookingtrack_tb b ON u.user_id = b.user_id
		WHERE b.booking_date = $1 
		AND b.status != 'cancelled'
		AND u.deviceid IS NOT NULL
		AND u.deviceid != ''
	`

	rows, err := handlerconn.Db.Query(query, bookingDate)
	if err != nil {
		log.Printf("Error querying users for cancellation notifications: %v", err)
		return
	}
	defer rows.Close()

	var serviceName string
	err = handlerconn.Db.QueryRow(`SELECT servicename FROM serviceavailable WHERE serv_id = $1`, serviceID).Scan(&serviceName)
	if err != nil {
		log.Printf("Error getting service name: %v", err)
		serviceName = "Medical Service"
	}

	var doctorName string
	err = handlerconn.Db.QueryRow(`SELECT doctorname FROM doctors WHERE doctor_id = $1`, doctorID).Scan(&doctorName)
	if err != nil {
		log.Printf("Error getting doctor name: %v", err)
		doctorName = "Doctor"
	}

	startTimeDisplay := startTime
	if len(startTime) > 5 {
		startTimeDisplay = startTime[:5]
	}
	endTimeDisplay := endTime
	if len(endTime) > 5 {
		endTimeDisplay = endTime[:5]
	}

	for rows.Next() {
		var fullname, deviceid string
		var userID int
		if err := rows.Scan(&fullname, &deviceid, &userID); err != nil {
			log.Printf("Error scanning user row: %v", err)
			continue
		}

		if !isValidExpoToken(deviceid) {
			log.Printf("Invalid Expo token for user %s: %s", fullname, deviceid)
			continue
		}

		// Create notification message
		message := ExpoMessage{
			To:    deviceid,
			Title: "New Appointment Slot Available!",
			Body: fmt.Sprintf("Hi %s! A slot has become available on %s at %s-%s with %s for %s. Book now before it's taken!",
				fullname, bookingDate, startTimeDisplay, endTimeDisplay, doctorName, serviceName),
			Sound: "default",
		}

		// Send the notification
		if err := SendNotification(message); err != nil {
			log.Printf("Failed to send cancellation notification to %s: %v", fullname, err)
		} else {
			log.Printf("Cancellation notification sent to %s for slot %s-%s", fullname, startTimeDisplay, endTimeDisplay)
		}
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating through users: %v", err)
	}
}

// TestPushNotification - for manual testing
func TestPushNotification(w http.ResponseWriter, r *http.Request) {
	type request struct {
		DeviceID string `json:"deviceId"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}
	msg := ExpoMessage{
		To:    req.DeviceID,
		Title: "Test Notification",
		Body:  "This is a test push notification!",
		Sound: "default",
	}
	err := SendNotification(msg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
}
