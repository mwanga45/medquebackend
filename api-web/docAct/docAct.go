package docact

import (
	"encoding/json"
	"fmt"
	"log"
	sidefunc_test "medquemod/booking/verification"
	handlerconn "medquemod/db_conn"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type (
	Response struct {
		Message string `json:"message"`
		Success bool   `json:"success"`
		Data    interface{}
	}
	StaffRegister struct {
		Doctorname string `json:"username" validate:"required"`
		RegNo      string `json:"regNo" validate:"required"`
		Password   string `json:"password" validate:"required"`
		Confirmpwrd string `json:"confirmpwrd" validate:"required"`
		Specialist string `json:"specialist" validate:"required"`
		Phone      string `json:"phone" validate:"required"`
		Email      string `json:"email" validate:"required"`

	}
)

func Registration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid Payload",
			Success: false,
		})
		return
	}
	w.Header().Set("Content-Type","application/json")
	var stafreg StaffRegister
	json.NewDecoder(r.Body).Decode(&stafreg)

	//    check doc is valid identification or registration number
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
	if !isSame{
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
	if !isSpec{
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid Payload",
			Success: false,
		})
		return
	}
	errpaswrd := sidefunc_test.ValidateSecretkey(stafreg.Password)
	if errpaswrd != nil{
		json.NewEncoder(w).Encode(Response{
			Message: "password must have 6 character and atleast having one Uppercase and digit ",
			Success: false,
		})
		return
	}
	hashedpwrd, err := bcrypt.GenerateFromPassword([]byte(stafreg.Password),bcrypt.DefaultCost)
	 if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Internal ServerError",
			Success: false,
		})
		log.Println("something went  wrong", err)
		return
	 }
	_,errEXec := client.Exec(`INSERT INTO doctors (doctorname, password, email, specialist_name,phone, identification, role) VALUES($1,$2,$3,$4,$5,$6,$7) `, stafreg.Doctorname, hashedpwrd,stafreg.Email,stafreg.Specialist,stafreg.Phone,stafreg.RegNo,"dkt")
	if errEXec != nil{
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Internal ServerError",
			Success: false,
		})
		log.Println("something went wrong", errEXec)
		return
	}
	
	if errCommit := client.Commit(); errCommit != nil{
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
