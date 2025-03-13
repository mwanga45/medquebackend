package authentic

import "net/http"

// create struct for http request for registration
type reg_request struct{
	fullname string `json:"fullname"`
	phone_num string `json:"phone_num"`
	email_address string `json:"email_address`
	home_address string `json:"home_address"`
}
type response struct{
	Success bool `json:"success"`
	Message string `json:"messsage, omitempty`
	
}
func Handler(w http.ResponseWriter, r* http.Request ){
	if r.Method != http.MethodPost{
    http.Error(w, "Invalid methode ", http.StatusMethodNotAllowed)
	return
	}
	var req reg_request
	
}
func Reg_authentic()  {
}