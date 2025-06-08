package booking

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	handlerconn "medquemod/db_conn"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type (
	Respond struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Data    any `json:"data"`
	}
	UserbookingRequest struct {
	   ServiceName string `json:"servicename" validate:"required"`
       Time string `json:"time" validate:"required"`
	   Dayfrom  string `json:"from" validate:"required"`
	   Dayto  string `json:"to" validate:"required"`
	}
	BookingRequest struct {
		Username   string    `json:"username" validate:"required"`
		Time       time.Time `json:"time" validate:"required"`
		Department string    `json:"department" validate:"required"`
		Day        string    `json:"day" validate:"required"`
		Diseases   string    `json:"disease" validate:"required"`
		Doctor     string    `json:"doctor" validate:"required"`
		Secretkey  string    `json:"secretekey" validate:"required"`
		Section    string    `json:"section" validate:"required"`
		DeviceId   string    `json:"deviceId"`
		Age        string    `json:"age"`
	}
	BookingHistoryStr struct {
		Username   string `json:"username" validate:"required"`
		Secretekey string `json:"secretekey" validate:"required"`
	}
	ExpoMessage struct {
		To    string `json:"to"`
		Title string `json:"title"`
		Body  string `json:"body"`
		Sound string `json:"sound"`
	}
)

func Booking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Respond{
			Message: "Method isnt Allowed",
			Success: false,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	tx, errTx := handlerconn.Db.Begin()

	if errTx != nil {
		json.NewEncoder(w).Encode(Respond{
			Message: "Something went wrong Transaction Failed",
			Success: false,
		})
		return
	}
	defer tx.Rollback()
	var BR BookingRequest
	err := json.NewDecoder(r.Body).Decode(&BR)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Respond{
			Message: "Something went wrong here",
			Success: false,
		})
		return
	}
	if BR.Section == "Guest" {
		err = HandleGeust(tx, BR.Username, BR.Secretkey, BR.Time, BR.Department, BR.Day, BR.Diseases)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Respond{
				Message: "Bad request",
				Success: false,
			})
			return
		}
	}
	if BR.Section == "Shared" {
		err = HandleshareDevice(tx, BR.Username, BR.Time, BR.Secretkey, BR.DeviceId, BR.Department, BR.Day, BR.Diseases)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Respond{
				Message: "Invalid request",
				Success: false,
			})
			return
		}
	}
	if BR.Section == "Specialgroup" {
		err = Handlespecialgroup(tx, BR.Time, BR.Department, BR.Username, BR.Secretkey, BR.Day, BR.Diseases)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Respond{
				Message: "Failed to make booking",
				Success: false,
			})
			return
		}
	}
	if BR.Section == "normal" {
		err = HandleNormal(tx, BR.Username, BR.Secretkey, BR.Time, BR.Department, BR.Day, BR.Diseases)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Respond{
				Message: "something went wrong",
				Success: false,
			})
			return
		}
	}
	if err = tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Respond{
			Message: "Transaction failed to commit",
			Success: false,
		})
		return
	}
	json.NewEncoder(w).Encode(Respond{
		Message: "Successuly made booking",
		Success: true,
	})
}
func HandleNormal(tx *sql.Tx, username string, secretekey string, Time time.Time, department string, day string, disease string) error {
	var hashedsecretekey string
	err := tx.QueryRow("SELECT secretkey FROM Users  WHERE username = $1", username).Scan(&hashedsecretekey)
	if err != nil {
		return fmt.Errorf("something went wrong here:%w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedsecretekey), []byte(secretekey))
	if err != nil {
		return fmt.Errorf("something went wrong username or secretekey not match")
	}
	err = CheckbookingRequest(tx, Time, username)
	if err != nil {
		return fmt.Errorf("something went wrong failed to validate user: %w", err)
	}
	err = HandlecheckTime(tx, day, Time)
	if err != nil {
		return fmt.Errorf("something went wrong here %w", err)
	}
	query := "INSERT INTO bookingList (username,time,department,day,disease,secretekey) VALUES($1,$2,$3,$4,$5,$6)"
	_, err = tx.Exec(query, username, Time, department, day, disease, hashedsecretekey)
	if err != nil {
		return fmt.Errorf("failed to execute query %w", err)
	}
	return nil
}
func HandleGeust(tx *sql.Tx, username string, secretkey string, Time time.Time, department string, day string, diseases string) error {
	var hashedsecretekey string
	err := tx.QueryRow("SELECT secretkey FROM Users  WHERE username = $1", username).Scan(&hashedsecretekey)
	if err != nil {
		return fmt.Errorf("something went wrong here:%w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedsecretekey), []byte(secretkey))
	if err != nil {
		return fmt.Errorf("wrong secretkey or username %w", err)
	}
	err = CheckbookingRequest(tx, Time, username)
	if err != nil {
		return fmt.Errorf("something went wrong here ")
	}
	err = HandlecheckTime(tx, day, Time)
	if err != nil {
		return fmt.Errorf("something went wrong here %w", err)
	}
	query := "INSERT INTO bookingList (username,time,department,day,disease,secretekey) VALUES($1,$2,$3,$4,$5,$6)"
	_, err = tx.Exec(query, username, Time, department, day, diseases, secretkey)
	if err != nil {
		return fmt.Errorf("something went wrong here: %w", err)
	}
	return nil

}
func HandleshareDevice(tx *sql.Tx, username string, Time time.Time, secretekey string, deviceId string, department string, day string, diseases string) error {
	var hashedsecretekey string
	err := tx.QueryRow("SELECT secretekey FROM Users WHERE username = $1", username).Scan(&hashedsecretekey)
	if err != nil {
		return fmt.Errorf("something went wrong here %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedsecretekey), []byte(secretekey))
	if err != nil {
		return fmt.Errorf("something went wrong failed to proccess secretkey")
	}
	//check if is not same user try to make request for him self
	var matching bool
	query := "SELECT EXISTS(SELECT 1 FROM Users WHERE username = $1 AND secretekey = $2 AND deviceId = $3)"
	err = tx.QueryRow(query, username, hashedsecretekey, deviceId).Scan(&matching)

	if err != nil {
		return fmt.Errorf("failed to execute query")
	}
	if matching {
		return fmt.Errorf("bad request:you can not make your own booking using by  site is only for others: %w", err)
	}
	err = CheckbookingRequest(tx, Time, username)
	if err != nil {
		return fmt.Errorf("something went wrong failed to validate user: %w", err)
	}
	err = HandlecheckTime(tx, day, Time)
	if err != nil {
		return fmt.Errorf("something went wrong here %w", err)
	}
	query = "INSERT INTO bookingList (username,time,department,day,disease,secretekey) VALUES($1,$2,$3,$4,$5,$6)"
	_, err = tx.Exec(query, username, Time, department, day, diseases, hashedsecretekey)
	if err != nil {
		return fmt.Errorf("failed to execute query %w", err)
	}
	return nil
}

// This checks whether the personnel has already booked more than once—either for the same service or different ones. It also ensures thata personnel cannot make more than two bookings before completing their required medical test.
func CheckbookingRequest(tx *sql.Tx, newtime time.Time, username string) error {
	// check if person try to make more than two booking before complete one of each
	var Countbooking int
	err := tx.QueryRow("SELECT COUNT(*) FROM bookingList WHERE username = $1 AND status = 'processing'", username).Scan(&Countbooking)
	if err != nil {
		return fmt.Errorf("something went wrong failed to fetch data: %w", err)
	}
	if Countbooking >= 2 {
		return fmt.Errorf("booking limit reached: complete at least one medical test before booking again")
	}
	// check if personal try to make more than one booking within are day
	WindowStart := newtime.Add(-24 * time.Hour)
	WindowEnd := newtime.Add(24 * time.Hour)
	var windowCount int

	query := "SELECT COUNT(*) FROM bookingList WHERE username = $1 AND time BETWEEN $2 AND $3"
	err = tx.QueryRow(query, username, WindowStart, WindowEnd).Scan(&windowCount)
	if err != nil {
		return fmt.Errorf("something went wrong %w", err)
	}
	if windowCount > 0 {
		return fmt.Errorf("booking not allowed: you can only make one booking within a 24-hour period")
	}
	return nil
}
func Handlespecialgroup(tx *sql.Tx, Time time.Time, department string, username string, secretekey string, day string, diseases string) error {
	var existuser bool
	err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM Users WHERE username = $1 AND secretekey = $2)", username, secretekey).Scan(&existuser)
	if err != nil {
		return fmt.Errorf("something went wrong here: %w", err)
	}
	if existuser {
		return fmt.Errorf("patient already exist in system")
	}
	err = HandlecheckTime(tx, day, Time)
	if err != nil {
		return fmt.Errorf("something went wrong here %w", err)
	}
	query := "INSERT INTO bookingList (username,time,department,day,disease,secretekey) VALUES($1,$2,$3,$4,$5,$6)"
	_, err = tx.Exec(query, username, Time, department, day, diseases, secretekey)
	if err != nil {
		return fmt.Errorf("failed to execute query %w", err)
	}
	return nil
}

// htpp request for the Historical booking list

func BookingHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Respond{
			Message: "Bad request",
			Success: false,
		})
		return
	}
	tx, errTx := handlerconn.Db.Begin()
	if errTx != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Respond{
			Message: "Something went wrong Transaction failed begin ",
			Success: false,
		})
		return
	}
	defer tx.Rollback()

	var U BookingHistoryStr
	err := json.NewDecoder(r.Body).Decode(&U)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Respond{
			Message: "Something went wrong failed to decode",
			Success: false,
		})
		return
	}
	list, err := ReturnAll(tx, U.Username, U.Secretekey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Respond{
			Message: "failed to return value ",
			Success: false,
		})
		return
	}
	if err := tx.Commit(); err != nil {
		json.NewEncoder(w).Encode(Respond{
			Message: "Something went wrong failed to commit transaction",
			Success: false,
		})
	}
	json.NewEncoder(w).Encode(Respond{
		Message: "Success return value",
		Success: true,
		Data:    list,
	})

}

// create function  to return all history available for this  patient
func ReturnAll(tx *sql.Tx, username string, secretekey string) (interface{}, error) {
	var validateuser bool
	var hashedsecretekey string
	err := tx.QueryRow("SELECT secretekey FROM Users WHERE username = $1", username).Scan(&hashedsecretekey)
	if err != nil {
		return "", fmt.Errorf("something went wrong or user isn`t exist  in system yet")
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedsecretekey), []byte(secretekey))
	if err != nil {
		return "", fmt.Errorf("something went wrong  username and secretekey isn`t matched")
	}
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM bookingList WHERE username = $1 AND secretekey = $2)", username, hashedsecretekey).Scan(&validateuser)
	if err != nil {
		return "", fmt.Errorf("something went wrong failed to execute query")
	}
	if !validateuser {
		return "No record exist yet", nil
	}
	rows, err := tx.Query("SELECT * FROM bookingList WHERE  username = $1 AND secretekey = $2 ORDER BY day", username, hashedsecretekey)

	if err != nil {
		return "", fmt.Errorf("failed to return booking record")
	}
	//  create slice to hold the  return column
	var Lists []map[string]interface{}
	column, _ := rows.Columns()
	count := len(column)
	for rows.Next() {
		list := make([]interface{}, count)
		ptrlist := make([]interface{}, count)
		for i := range column {
			ptrlist[i] = &list[i]
		}
		err := rows.Scan(ptrlist...)
		if err != nil {
			return "", fmt.Errorf("failed to scan rows")
		}
		List := make(map[string]interface{})
		for i, col := range column {
			val := list[i]
			List[col] = val
		}
		Lists = append(Lists, List)
	}
	if err = rows.Err(); err != nil {
		return "", fmt.Errorf("error during row iteration: %w", err)
	}
	return Lists, nil
}

// function to check if that time  is already been selected
func HandlecheckTime(tx *sql.Tx, day string, time time.Time) error {
	var exist bool
	err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM bookingList WHERE day = $1 AND time = $2)", day, time).Scan(&exist)
	if err != nil {
		return fmt.Errorf("something went wrong failed to execute query %w", err)
	}
	if exist {
		return fmt.Errorf("your try to make booking to exist slot")
	}
	return nil
}

// SendNotification sends a push notification via Expo's API with proper error handling
func SendNotification(message ExpoMessage) error {
	// 1. Validate Expo token format first
	if !isValidExpoToken(message.To) {
		return fmt.Errorf("invalid Expo token format")
	}

	// 2. Marshal with proper error handling
	payload, err := json.Marshal([]ExpoMessage{message})
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	// 3. Use correct Expo endpoint
	resp, err := http.Post(
		"https://exp.host/--/api/v2/push/send",
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	// 4. Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expo API error %d: %s", resp.StatusCode, string(body))
	}

	// 5. Parse Expo's JSON response
	var result struct {
		Data []struct {
			Status  string `json:"status"`
			ID      string `json:"id"`
			Message string `json:"message"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("invalid response format: %w", err)
	}

	// 6. Validate individual message status
	if len(result.Data) == 0 || result.Data[0].Status != "ok" {
		return fmt.Errorf("expo rejection: %s", result.Data[0].Message)
	}

	return nil
}

// storeNotificationInDB persists notifications with proper validation
func storeNotificationInDB(br BookingRequest) error {
	// 1. Always get fresh device ID with NULL checking
	var deviceID sql.NullString
	err := handlerconn.Db.QueryRow(
		"SELECT deviceId FROM Users WHERE username = $1",
		br.Username,
	).Scan(&deviceID)

	if err != nil {
		return fmt.Errorf("user query failed: %w", err)
	}
	if !deviceID.Valid {
		return fmt.Errorf("no device ID registered")
	}

	// 2. Use UTC times for consistency
	notificationTime := br.Time.UTC().Add(-10 * time.Minute)
	bookingTime := br.Time.UTC()

	// 3. Insert with explicit time zone
	_, err = handlerconn.Db.Exec(
		`INSERT INTO scheduled_notifications 
        (username, device_id, notification_time, booking_time)
        VALUES ($1, $2, $3, $4)`,
		br.Username,
		deviceID.String,
		notificationTime,
		bookingTime,
	)

	return err
}

// checkPendingNotifications finds due notifications with proper SQL escaping
func checkPendingNotifications() {
	// 1. Use parameterized query with proper time window
	query := `SELECT id, username, device_id, notification_time 
              FROM scheduled_notifications 
              WHERE status = $1 
              AND notification_time BETWEEN NOW() - INTERVAL '1 minute' AND NOW() + INTERVAL '1 minute'`

	rows, err := handlerconn.Db.Query(query, "pending")
	if err != nil {
		log.Printf("Query error: %v", err)
		return
	}
	defer rows.Close()

	// 2. Process rows with error checking
	for rows.Next() {
		var (
			id               int
			username         string
			deviceID         string
			notificationTime time.Time
		)

		if err := rows.Scan(&id, &username, &deviceID, &notificationTime); err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}

		scheduleFromDB(id, username, deviceID, notificationTime)
	}

	// 3. Check for iteration errors
	if err = rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
	}
}

// scheduleFromDB handles notification timing with retries
func scheduleFromDB(id int, username string, deviceID string, notificationTime time.Time) {
	// 1. Convert to UTC for calculation
	now := time.Now().UTC()
	notificationTime = notificationTime.UTC()

	// 2. Validate notification time
	if notificationTime.Before(now.Add(-1 * time.Minute)) {
		updateNotificationStatus(id, "expired")
		return
	}

	// 3. Calculate remaining time with buffer
	delay := notificationTime.Sub(now)
	if delay < 0 {
		delay = 0
	}

	// 4. Schedule with retry logic
	time.AfterFunc(delay, func() {
		const maxRetries = 3
		message := ExpoMessage{
			To:    deviceID,
			Title: "Booking Reminder",
			Body:  fmt.Sprintf("Your booking is soon!"),
		}

		var err error
		for attempt := 1; attempt <= maxRetries; attempt++ {
			err = SendNotification(message)
			if err == nil {
				updateNotificationStatus(id, "sent")
				return
			}

			// Exponential backoff
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		updateNotificationStatus(id, "failed")
		log.Printf("Failed after %d attempts: %v", maxRetries, err)
	})
}

// updateNotificationStatus handles status changes with timestamp
func updateNotificationStatus(id int, status string) {
	// 1. Validate allowed statuses
	validStatuses := map[string]bool{
		"pending": true,
		"sent":    true,
		"failed":  true,
		"expired": true,
	}
	if !validStatuses[status] {
		log.Printf("Invalid status: %s", status)
		return
	}

	// 2. Use parameterized query
	_, err := handlerconn.Db.Exec(
		`UPDATE scheduled_notifications 
         SET status = $1, updated_at = NOW() AT TIME ZONE 'UTC'
         WHERE id = $2`,
		status, id,
	)

	if err != nil {
		log.Printf("Update error: %v", err)
	}
}

// Helper function for Expo token validation
func isValidExpoToken(token string) bool {
	return strings.HasPrefix(token, "ExponentPushToken[") &&
		strings.HasSuffix(token, "]") &&
		len(token) > 25
}

// Modified worker functions with proper synchronization
func StartNotificationWorker(ctx context.Context) {
	log.Println("Starting notification worker")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Initial check on startup
	checkPendingNotifications()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping notification worker")
			return
		case <-ticker.C:
			checkPendingNotifications()
		}
	}
}
