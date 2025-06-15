package profile

import (
	"database/sql"
	"encoding/json"
	"fmt"
	sidefunc_test "medquemod/booking/verification"
	handlerconn "medquemod/db_conn"
	"medquemod/middleware"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type (
	Response struct {
		Message   string      `json:"message"`
		Success   bool        `json:"success"`
		Erroruser string      `json:"error"`
		Data      interface{} `json:"data,omitempty"`
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
	hashsecretkey,_ := bcrypt.GenerateFromPassword([]byte(reqpayload.Secretkey), bcrypt.DefaultCost)
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
