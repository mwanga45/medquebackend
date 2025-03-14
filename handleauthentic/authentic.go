package authentic

import (
	"database/sql"
	"encoding/json"
	"medquemod/db_conn"
	"net/http"
)

// create struct for http request for registration
type reg_request struct{
	Fullname string `json:"fullname"`
	Phone_num string `json:"phone_num"`
	Email_address string `json:"email_address"`
	Home_address string `json:"home_address"`
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

	if  req.Fullname == "" || req.Email_address == "" ||req.Phone_num == "" {
		http.Error(w, "some empty please field all reqiure field ", http.StatusBadRequest)
		return
	}

	query := "SELECT   email, phone_number FROM Patients WHERE email = $1 OR phone_number =$2"
   var emaiexist , phoneexist string
	err := handlerconn.Db.QueryRow(query,req.Email_address,req.Phone_num).Scan(&emaiexist, &phoneexist)

	if err == nil{
		http.Error(w, "email or phone number already exist", http.StatusBadRequest)
		return
	}
	if err != sql.ErrNoRows{
		http.Error(w,"server error",http.StatusInternalServerError )
		return
	}

	insert_query := "INSERT INTO Patients(full_name, home_address, email,phone_number,user_type) VALUES($1,$2,$3,$4,$5)"

	_,err = handlerconn.Db.Exec(insert_query,req.Fullname,req.Home_address,req.Email_address,req.Phone_num,"Patient")

	if err != nil{
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	res := response{Success: true, Message: "successfuly registered"}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(res)

}
