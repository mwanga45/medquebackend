package docact

import (
	"encoding/json"
	"fmt"
	"log"
	sidefunc_test "medquemod/booking/verification"
	handlerconn "medquemod/db_conn"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type (
	Response struct {
		Message string `json:"message"`
		Success bool   `json:"success"`
		Data    interface{}
	}
	StaffRegister struct {
		Doctorname  string `json:"username" validate:"required"`
		RegNo       string `json:"regNo" validate:"required"`
		Password    string `json:"password" validate:"required"`
		Confirmpwrd string `json:"confirmpwrd" validate:"required"`
		Specialist  string `json:"specialist" validate:"required"`
		Phone       string `json:"phone" validate:"required"`
		Email       string `json:"email" validate:"required"`
	}
	Stafflogin struct {
		RegNo    string `json:"regNo" validate:"required"`
		Password string `json:"password" validate:"required"`
		Username string `json:"username" validate:"required"`
	}
	DoctorData struct {
		DoctorID       int    `json:"doctor_id"`
		DoctorName     string `json:"doctorname"`
		Email          string `json:"email"`
		SpecialistName string `json:"specialist_name"`
		Phone          string `json:"phone"`
		Identification string `json:"identification"`
		Role           string `json:"role"`
	}
	LoginResponse struct {
		Token  string     `json:"token"`
		Doctor DoctorData `json:"doctor"`
	}
)


type Claims struct {
	DoctorID int    `json:"doctor_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}


func generateToken(doctorID int, username, role string) (string, error) {

	secretKey := []byte("your-secret-key-here")

	claims := Claims{
		DoctorID: doctorID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), 
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func DoctLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Method not allowed",
			Success: false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var staffLogin Stafflogin
	if err := json.NewDecoder(r.Body).Decode(&staffLogin); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid request body",
			Success: false,
		})
		return
	}


	if staffLogin.Username == "" || staffLogin.RegNo == "" || staffLogin.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: "Username, registration number, and password are required",
			Success: false,
		})
		return
	}


	isValid := sidefunc_test.CheckIdentifiaction(staffLogin.RegNo)
	if !isValid {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid registration number format",
			Success: false,
		})
		return
	}


	var doctorData DoctorData
	var hashedPassword string

	err := handlerconn.Db.QueryRow(`
		SELECT doctor_id, doctorname, password, email, specialist_name, phone, identification, role 
		FROM doctors 
		WHERE doctorname = $1 AND identification = $2
	`, staffLogin.Username, staffLogin.RegNo).Scan(
		&doctorData.DoctorID,
		&doctorData.DoctorName,
		&hashedPassword,
		&doctorData.Email,
		&doctorData.SpecialistName,
		&doctorData.Phone,
		&doctorData.Identification,
		&doctorData.Role,
	)

	if err != nil {
		log.Printf("Database query error: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid username or registration number",
			Success: false,
		})
		return
	}

	
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(staffLogin.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid password",
			Success: false,
		})
		return
	}


	token, err := generateToken(doctorData.DoctorID, doctorData.DoctorName, doctorData.Role)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Failed to generate authentication token",
			Success: false,
		})
		return
	}


	loginResponse := LoginResponse{
		Token:  token,
		Doctor: doctorData,
	}

	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Message: "Login successful",
		Success: true,
		Data:    loginResponse,
	})
}

func Registration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid Payload",
			Success: false,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var stafreg StaffRegister
	json.NewDecoder(r.Body).Decode(&stafreg)

	
	isValid := sidefunc_test.CheckIdentifiaction(stafreg.RegNo)
	if !isValid {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid Payload ",
			Success: false,
		})
		return
	}
	isSame := sidefunc_test.Checkpassword(stafreg.Password, stafreg.Confirmpwrd)
	if !isSame {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: "Confirm password and password are not the same",
			Success: false,
		})
		return
	}
	var isExist bool
	client, errTx := handlerconn.Db.Begin()
	if errTx != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Interna ServerError",
			Success: false,
		})
		fmt.Println("Something went wrong", errTx)
		return
	}
	defer client.Rollback()
	errCheck := client.QueryRow(`SELECT EXISTS(SELECT 1 FROM doctors WHERE email = $1 AND  doctorname = $2 )`, stafreg.Email, stafreg.Doctorname).Scan(&isExist)
	if errCheck != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Internal ServerError",
			Success: false,
		})
		fmt.Println("something went wrong", errCheck)
		return
	}
	var isSpec bool
	errspec := client.QueryRow(`SELECT EXISTS(SELECT 1 FROM specialist WHERE specialist  = $1)`, stafreg.Specialist).Scan(&isSpec)
	if errspec != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Internal ServerError",
			Success: false,
		})
		fmt.Println("something went wrong", errCheck)
		return
	}
	if !isSpec {
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid Payload",
			Success: false,
		})
		return
	}
	errpaswrd := sidefunc_test.ValidateSecretkey(stafreg.Password)
	if errpaswrd != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "password must have 6 character and atleast having one Uppercase and digit ",
			Success: false,
		})
		return
	}
	hashedpwrd, err := bcrypt.GenerateFromPassword([]byte(stafreg.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Internal ServerError",
			Success: false,
		})
		log.Println("something went  wrong", err)
		return
	}
	_, errEXec := client.Exec(`INSERT INTO doctors (doctorname, password, email, specialist_name,phone, identification, role,assgn_status) VALUES($1,$2,$3,$4,$5,$6,$7,$8) `, stafreg.Doctorname, hashedpwrd, stafreg.Email, stafreg.Specialist, stafreg.Phone, stafreg.RegNo, "dkt",false)
	if errEXec != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Internal ServerError",
			Success: false,
		})
		log.Println("something went wrong", errEXec)
		return
	}

	if errCommit := client.Commit(); errCommit != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "System failed to commit request try Again",
			Success: false,
		})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{
		Message: "successfuly registered",
		Success: true,
	})
}
