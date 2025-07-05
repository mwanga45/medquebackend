package profile

import (
	"database/sql"
	"encoding/json"
	"fmt"
	sidefunc_test "medquemod/booking/verification"
	handlerconn "medquemod/db_conn"
	"medquemod/middleware"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type (
	Response struct {
		Message   string      `json:"message"`
		Success   bool        `json:"success"`
		Erroruser string      `json:"error"`
		Data      interface{} `json:"data,omitempty"`
	}
	Historyforme struct {
		BookingID   int    `json:"booking_id"`
		UserId      int    `json:"user_id"`
		Service_id  int    `json:"service_id"`
		Spec_id     int    `json:"spec_id"`
		DayofWeek   int    `json:"dayofweek"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
		BookingDate string `json:"booking_date"`
		Status      string `json:"status"`
	}
	Createpayload struct {
		Age        int    `json:"age"`
		Firstname  string `json:"firstname"`
		Secondname string `json:"Secondname"`
		Dial       string `json:"dial"`
		Secretkey  string `json:"secretkey"`
		Reason     string `json:"reason"`
	}
)

func BookingHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid method",
			Success: false,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")

	claims, ok := r.Context().Value("user").(*middleware.CustomClaims)
	if !ok {
		json.NewEncoder(w).Encode(Response{
			Message: "Unauthorized",
			Success: false,
		})
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if user exists
	var isExist bool
	errcheckId := handlerconn.Db.QueryRow(`SELECT EXISTS(SELECT 1 FROM Users WHERE user_id = $1)`, claims.ID).Scan(&isExist)
	if errcheckId != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Internal Server Error",
			Success: false,
		})
		return
	}

	if !isExist {
		json.NewEncoder(w).Encode(Response{
			Message: "User doesn't exist",
			Success: false,
			Data:    nil,
		})
		return
	}

	// Get all bookings for the user - using COALESCE to handle NULL spec_id
	rows, err := handlerconn.Db.Query(`SELECT id, user_id, COALESCE(spec_id, 0) as spec_id, service_id, dayofweek, start_time, end_time, booking_date, status FROM bookingtrack_tb WHERE user_id = $1 ORDER BY created_at DESC`, claims.ID)
	if err != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Internal Server Error",
			Success: false,
		})
		return
	}
	defer rows.Close()

	historyList := make([]Historyforme, 0)
	for rows.Next() {
		var h Historyforme
		err := rows.Scan(&h.BookingID, &h.UserId, &h.Spec_id, &h.Service_id, &h.DayofWeek, &h.StartTime, &h.EndTime, &h.BookingDate, &h.Status)
		if err != nil {
			fmt.Printf("Scan error: %v\n", err)
			json.NewEncoder(w).Encode(Response{
				Message: "Error scanning row: " + err.Error(),
				Success: false,
			})
			return
		}
		historyList = append(historyList, h)
	}

	if errRow := rows.Err(); errRow != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Something went wrong",
			Success: false,
		})
		return
	}

	json.NewEncoder(w).Encode(Response{
		Message: "Booking history fetched successfully",
		Success: true,
		Data:    historyList,
	})
}
func UserAct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid payload",
			Success: false,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	client, errTx := handlerconn.Db.Begin()
	if errTx != nil {

		json.NewEncoder(w).Encode(Response{
			Message: "Internal ServerError",
			Success: false,
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer client.Rollback()
	claims, ok := r.Context().Value("user").(*middleware.CustomClaims)
	if !ok {
		json.NewEncoder(w).Encode(Response{
			Message: "Unauthorized",
			Success: false,
		})
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var Phone string
	errcheckId := client.QueryRow(`SELECT dial FROM Users WHERE user_id = $1 `, claims.ID).Scan(&Phone)
	if errcheckId != nil {
		if errcheckId == sql.ErrNoRows {
			json.NewEncoder(w).Encode(Response{
				Message: "User doesn`t exist ",
				Success: false,
			})
			return
		}
		json.NewEncoder(w).Encode(Response{
			Message: "Interna ServerError",
			Success: false,
		})
		return
	}
	var reqpayload Createpayload

	errDec := json.NewDecoder(r.Body).Decode(&reqpayload)
	if errDec != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Internal serverError please try again",
			Success: false,
		})
		return
	}
	specialnames, err := sidefunc_test.CheckLimit(claims.ID, client)
	if err != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Internal Server Error",
			Success: false,
			Data:    err.Error(),
		})
		fmt.Println("something went wrong", err)
		return
	}

	Username := fmt.Sprint(reqpayload.Firstname + " " + reqpayload.Secondname)
	if errsecretkey := sidefunc_test.ValidateSecretkey(reqpayload.Secretkey); errsecretkey != nil {
		json.NewEncoder(w).Encode(Response{
			Erroruser: errsecretkey.Error(),
			Success:   false,
		})
		return
	}
	hashsecretkey, _ := bcrypt.GenerateFromPassword([]byte(reqpayload.Secretkey), bcrypt.DefaultCost)
	var newspecId int

	InsertError := client.QueryRow(`INSERT INTO Specialgroup (Username,secretkey, Age, managedby_id, dialforCreator, dialforUser, reason ) VALUES($1,$2,$3,$4,$5,$6) RETURNING spec_id`, Username, hashsecretkey, reqpayload.Age, Phone, reqpayload.Dial, reqpayload.Reason).Scan(&newspecId)

	if InsertError != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Failed to Create new user, Internal ServerERROR",
			Success: false,
		})
		fmt.Println("something went wrong", err)
		return
	}

	if errcommit := client.Commit(); errcommit != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Internal ServerError",
			Success: false,
		})
		fmt.Println("something went wrong", errcommit)
		return
	}
	specialnames[newspecId] = Username

	json.NewEncoder(w).Encode(Response{
		Message: "successfully create new user",
		Success: true,
		Data:    specialnames,
	})
}
func PendingBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid Method",
			Success: false,
		})
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	claim, ok := r.Context().Value("user").(*middleware.CustomClaims)
	if !ok {
		json.NewEncoder(w).Encode(Response{
			Message: "Unauthorized user",
			Success: false,
		})
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if user exists
	var isExist bool
	errcheckId := handlerconn.Db.QueryRow(`SELECT EXISTS(SELECT 1 FROM Users WHERE user_id = $1)`, claim.ID).Scan(&isExist)
	if errcheckId != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Internal Server Error",
			Success: false,
		})
		return
	}

	if !isExist {
		json.NewEncoder(w).Encode(Response{
			Message: "User doesn't exist",
			Success: false,
			Data:    nil,
		})
		return
	}

	// Get current time and date
	timeNow := time.Now()
	currentDate := timeNow.Format("2006-01-02")
	currentTime := timeNow.Format("15:04:05")

	// Query for pending bookings that haven't reached their time and date yet
	query := `
		SELECT id, user_id, COALESCE(spec_id, 0) as spec_id, service_id, dayofweek, 
		       start_time, end_time, booking_date, status
		FROM bookingtrack_tb 
		WHERE user_id = $1 
		AND status != 'cancelled'
		AND (
			(booking_date > $2) 
			OR (booking_date = $2 AND start_time > $3)
		)
		ORDER BY booking_date ASC, start_time ASC
	`

	rows, err := handlerconn.Db.Query(query, claim.ID, currentDate, currentTime)
	if err != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Internal Server Error",
			Success: false,
		})
		return
	}
	defer rows.Close()

	pendingList := make([]Historyforme, 0)
	for rows.Next() {
		var h Historyforme
		err := rows.Scan(&h.BookingID, &h.UserId, &h.Spec_id, &h.Service_id, &h.DayofWeek, &h.StartTime, &h.EndTime, &h.BookingDate, &h.Status)
		if err != nil {
			fmt.Printf("Scan error: %v\n", err)
			json.NewEncoder(w).Encode(Response{
				Message: "Error scanning row: " + err.Error(),
				Success: false,
			})
			return
		}
		pendingList = append(pendingList, h)
	}

	if errRow := rows.Err(); errRow != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Something went wrong",
			Success: false,
		})
		return
	}

	json.NewEncoder(w).Encode(Response{
		Message: "Pending bookings fetched successfully",
		Success: true,
		Data:    pendingList,
	})
}
