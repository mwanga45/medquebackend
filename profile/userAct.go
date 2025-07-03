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
	Historyforme struct {
		UserId      int    `json:"userId"`
		ServiceId   int    `json:"serviceId"`
		SpecId      int    `json:"specId"`
		DayOfWeek   int    `json:"dayOfWeek"`
		StartTime   string `json:"startTime"`
		EndTime     string `json:"endTime"`
		BookingDate string `json:"bookingDate"`
	}
	// HistoryNotforme struct {
	// 	UserId   int    `json:"user_id"`
	// 	Spec_id int    `json:"spec_id"`
	// 	Service_id int	`json:"service_id"`
	// 	DayofWeek int `json:"dayofweek"`
	// }
	CreatePayload struct {
		Age        int    `json:"age"`
		FirstName  string `json:"firstName"`
		SecondName string `json:"secondName"`
		Dial       string `json:"dial"`
		SecretKey  string `json:"secretKey"`
		Reason     string `json:"reason"`
	}
)

func BookingHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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
	var isExist bool
	errcheckId := client.QueryRow(`SELECT EXISTS(SELECT 1 FROM Users WHERE user_id = $1)`, claims.ID).Scan(&isExist)
	if errcheckId != nil {
		if errcheckId == sql.ErrNoRows {
			json.NewEncoder(w).Encode(Response{
				Message: "User doesn`t exist ",
				Success: false,
				Data:    nil,
			})
			return
		}
		json.NewEncoder(w).Encode(Response{
			Message: "Internal Server Error",
			Success: false,
		})
		return
	}
	var history Historyforme
	json.NewDecoder(r.Body).Decode(&history)
	rows, err := client.Query(`SELECT user_id, spec_id, service_id, dayofweek, starttime, endtime, bookingdate FROM Bookingtrack_tb WHERE user_id = $1 AND spec_id = $2 AND service_id = $3`, claims.ID, history.SpecId, history.ServiceId)
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
		err := rows.Scan(&h.UserId, &h.SpecId, &h.ServiceId, &h.DayOfWeek, &h.StartTime, &h.EndTime, &h.BookingDate)
		if err != nil {
			json.NewEncoder(w).Encode(Response{
				Message: "Error scanning row",
				Success: false,
			})
			return
		}
		historyList = append(historyList, h)
	}
	if errCommit := client.Commit(); errCommit != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Internal ServerError",
			Success: false,
		})
		fmt.Println("something went wrong", errCommit)
		return
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
	var reqpayload CreatePayload

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

	Username := fmt.Sprint(reqpayload.FirstName + " " + reqpayload.SecondName)
	if errsecretkey := sidefunc_test.ValidateSecretkey(reqpayload.SecretKey); errsecretkey != nil {
		json.NewEncoder(w).Encode(Response{
			Erroruser: errsecretkey.Error(),
			Success:   false,
		})
		return
	}
	hashsecretkey, _ := bcrypt.GenerateFromPassword([]byte(reqpayload.SecretKey), bcrypt.DefaultCost)
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
