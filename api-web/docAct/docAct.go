package docact

import (
	"encoding/json"
	sidefunc_test "medquemod/booking/verification"
	"net/http"
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
	var stafreg StaffRegister
	json.NewDecoder(r.Body).Decode(&stafreg)

	//    check doc is valid identification or registration number
	isValid := sidefunc_test.CheckIdentifiaction(stafreg.RegNo)
	if !isValid{
		w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(Response{
			Message: "Invalid Payload ",
			Success: false,
		})
		return
	}
	

}
