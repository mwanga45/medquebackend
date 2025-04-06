package booking

// here  one device will be used to make booking to day for adult  for child  it will be almost 3 child
import (
	"database/sql"
	"encoding/json"
	"fmt"
	handlerconn "medquemod/db_conn"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type (
	Respond struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
	}
	BookingRequest struct {
		Username   string    `json:"username" validate:"required"`
		Time       time.Time `json:"time" validate:"required"`
		Department string    `json:"department" validate:"required"`
		Day        string    `json:"day" validate:"required"`
		Diseases   string    `json:"desease" validate:"required"`
		Doctor     string    `json:"doctor" validate:"required"`
		Secretkey  string    `json:"secretekey" validate:"required"`
	}
	Secretkey struct {
		Secretkey string `json:"deviceId"`
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
	err = HandleGeust(tx, BR.Username, BR.Secretkey, BR.Time,BR.Department,BR.Day,BR.Diseases)
	if err !=nil{
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Respond{
			Message: "Bad request",
			Success: false,
		})
	}
	
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
	errFunc := CheckbookingRequest(Time, username)
	if errFunc != nil {
		return fmt.Errorf("somthing  went wrong %w", errFunc)
	}
	query := "INSERT INTO bookingList (username,time,department,day,disease,secretekey) VALUES($1,$2,$3,$4,$5,$6)"
	_, err = tx.Exec(query, username, Time, department, day, diseases, secretkey)
	if err != nil {
		return fmt.Errorf("something went wrong here: %w", err)
	}
	return nil

}

// func Handlechild(){

// }
// func HandleshareDevice() {

// }
// This checks whether the personnel has already booked more than once—either for the same service or different ones. It also ensures thata personnel cannot make more than two bookings before completing their required medical test.
func CheckbookingRequest(newtime time.Time, username string) error {
	tx, errTx := handlerconn.Db.Begin()
	defer tx.Rollback()
	if errTx != nil {
		return fmt.Errorf("something went wrong transaction failed:%w", errTx)
	}
	// check if person try to make more than two booking before complete one of each
	var Countbooking int
	err := tx.QueryRow("SELECT COUNT(*) FROM bookingList WHERE username = $1 AND status = 'processing'", username).Scan(&Countbooking)
	if err != nil {
		return fmt.Errorf("something went wrong failed to fetch data : %w", err)
	}
	if Countbooking == 2 {
		return fmt.Errorf("failed to make booking please complete atleast one of the medical test")
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
		return fmt.Errorf("your will be  allowed to make another  booking after 24hrs ")
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("something went wrong here %w", err)
	}
	return nil

}
