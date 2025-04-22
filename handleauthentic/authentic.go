package authentic

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"medquemod/db_conn"
	"net/http"
)

// create struct for http request for registration
type reg_request struct{
	Firstname  string `json:"firstname" validate:"required"`
	Secondname string `json:"secondname" validate:"required"`
	Secretekey string `json:"secretekey" validate:"required"`
	Dial string `json:"dial" validate:"required"`
	Email string `json:"email" validate:"required"`
	DeviceId string `json:"deviceid" validate:"required"`
	Birthdate string `json:"birthdate" validate:"required"`
	HomeAddress string `json:"homeaddress"`
}
type response struct{
	Success bool `json:"success"`
	Message string `json:"messsage,omitempty"`
	
}
func Handler(w http.ResponseWriter, r* http.Request ){
	if r.Method != http.MethodPost{
    http.Error(w, "Invalid methode ", http.StatusMethodNotAllowed)
	return
	}
	var req reg_request

	if err := json.NewDecoder(r.Body).Decode(&req);err != nil{
		http.Error(w, "invalid Request payload", http.StatusBadRequest)
		return
	}

	if  req.Firstname == "" || req.Secondname == "" ||req.Secretekey == "" || req.Birthdate =="" || req.DeviceId ==""||req.Dial == "" || req.Email == "" {
		http.Error(w, "some empty please field all reqiure field ", http.StatusBadRequest)
		return
	}

	query := "SELECT   email, dial FROM Users WHERE email = $1 OR phone_number =$2"
   var emaiexist , phoneexist string
	err := handlerconn.Db.QueryRow(query,req.Email,req.Dial).Scan(&emaiexist, &phoneexist)

	if err == nil{
		http.Error(w, "email or phone number already exist", http.StatusBadRequest)
		return
	}
	if err != sql.ErrNoRows{
		http.Error(w,"server error",http.StatusInternalServerError )
		return
	}
   Fullname := fmt.Sprint(req.Firstname + " " + req.Secondname)
	insert_query := "INSERT INTO Users(fullname,Secretekey,phone_number,email,deviceId,Birthdate,home_address,user_type) VALUES($1,$2,$3,$4,$5,$6,$7,$8)"

	_,err = handlerconn.Db.Exec(insert_query,Fullname,req.Secretekey,req.Dial,req.Email,req.DeviceId,req.Birthdate,req.HomeAddress,"Patient")

	if err != nil{
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	res := response{Success: true, Message: "successfuly registered"}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(res)

}
