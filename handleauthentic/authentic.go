package authentic

import (
	"database/sql"
	"encoding/json"
	"fmt"
	handlerconn "medquemod/db_conn"
	"medquemod/middleware"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type (
	loginRequest struct {
		Email     string `json:"email" validate:"required"`
		SecretKey string `json:"secretKey" validate:"required"`
	}
	regRequest struct {
		FirstName   string `json:"firstname" validate:"required"`
		SecondName  string `json:"secondname" validate:"required"`
		SecretKey   string `json:"secreteKey" validate:"required"`
		Dial        string `json:"dial" validate:"required"`
		Email       string `json:"email" validate:"required"`
		DeviceId    string `json:"deviceId" validate:"required"`
		Birthdate   string `json:"birthdate" validate:"required"`
		HomeAddress string `json:"homeaddress"`
	}
	response struct {
		Success bool   `json:"success"`
		Message string `json:"message,omitempty"`
		Token   string `json:"token,omitempty"`
	}
)

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid Methode used", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("content-type", "application/json")
	var log loginRequest
	if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
		http.Error(w, "Invalide payload", http.StatusBadRequest)
		json.NewEncoder(w).Encode(response{
			Message: "Invalid payload",
			Success: false,
		})
		return
	}
	var (
		hashsecretekey string
		userRole       string
		id             string
		username       string
	)

	errcheck := handlerconn.Db.QueryRow("SELECT Secretekey, user_type,fullname, user_id FROM Users WHERE email= $1", log.Email).Scan(&hashsecretekey, &userRole, &username, &id)
	if errcheck != nil {
		if errcheck == sql.ErrNoRows {
			json.NewEncoder(w).Encode(response{
				Message: "Wrong Secretekey or  Email",
				Success: false,
			})
			return
		}
		json.NewEncoder(w).Encode(response{
			Success: false,
			Message: "Something went wrong",
		})
		fmt.Println("Error in login", errcheck)
		return
	}
	if comparepassword := bcrypt.CompareHashAndPassword([]byte(hashsecretekey), []byte(log.SecretKey)); comparepassword != nil {
		http.Error(w, "Invalid Payload", http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response{
			Message: "Wrong password or Email",
			Success: false,
		})
		return
	}
	token, err := middleware.GenerateJWT(userRole, id, username)
	if err != nil {
		http.Error(w, "Something went wrong ", http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response{
			Message: "Internal Server Error",
			Success: false,
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response{
		Message: "succesdfully login",
		Success: true,
		Token:   token,
	})
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid methode ", http.StatusMethodNotAllowed)
		return
	}

	var req regRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid Request payload", http.StatusBadRequest)
		fmt.Print(err)
		return
	}

	if req.FirstName == "" || req.SecondName == "" || req.SecretKey == "" || req.Birthdate == "" || req.DeviceId == "" || req.Dial == "" || req.Email == "" {
		http.Error(w, "some empty please field all reqiure field ", http.StatusBadRequest)
		return
	}

	query := "SELECT   email, dial FROM Users WHERE email = $1 OR  dial =$2"
	var emaiexist, phoneexist string
	err := handlerconn.Db.QueryRow(query, req.Email, req.Dial).Scan(&emaiexist, &phoneexist)

	if err == nil {
		http.Error(w, "email or phone number already exist", http.StatusBadRequest)
		return
	}
	if err != sql.ErrNoRows {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	Fullname := fmt.Sprint(req.FirstName + " " + req.SecondName)
	hashedsecretekey, err := bcrypt.GenerateFromPassword([]byte(req.SecretKey), bcrypt.DefaultCost)
	if err != nil {
		json.NewEncoder(w).Encode(response{
			Message: "failed to hash secret key",
			Success: false,
		})
		return
	}
	insert_query := "INSERT INTO Users(fullname, secretekey, dial, email, deviceid, birthdate, home_address, user_type) VALUES($1,$2,$3,$4,$5,$6,$7,$8)"

	_, err = handlerconn.Db.Exec(insert_query, Fullname, hashedsecretekey, req.Dial, req.Email, req.DeviceId, req.Birthdate, req.HomeAddress, "Patient")

	if err != nil {
		fmt.Println("DB Insert Error:", err) 
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	res := response{Success: true, Message: "successfully registered"}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(res)

}
