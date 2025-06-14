package profile

import (
	"database/sql"
	"encoding/json"
	"fmt"
	sidefunc_test "medquemod/booking/verification"
	handlerconn "medquemod/db_conn"
	"medquemod/middleware"
	"net/http"
)

type (

	Response struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Data    interface{} `json:"data,omitempty"`
	}
	Createpayload struct{
		Age int `json:"age"`
		Firstname string `json:"firstname"`
		Secondname string `json:"Secondname"`
		Dial string `json:"dial"`
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
	client , errTx := handlerconn.Db.Begin()
	if errTx != nil{
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
	
	var phone string 
	errcheckId := client.QueryRow(`SELECT dial FROM Users WHERE user_id = $1 `,claims.ID).Scan(phone)
	if errcheckId != nil{
		if errcheckId == sql.ErrNoRows{
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
	if errDec != nil{
		json.NewEncoder(w).Encode(Response{
			Message: "Internal serverError please try again",
			Success: false,
		})
		return
	} 
	specialnames, err := sidefunc_test.CheckLimit(claims.ID,client)
	if err != nil{
		json.NewEncoder(w).Encode(Response{
			Message: "Internal Server Error",
			Success: false,
		})
		fmt.Errorf("something went wrong: %w", err)
		return
	}
	



}
