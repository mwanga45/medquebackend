package booking

import (
	"encoding/json"
	"net/http"
)

type (
	Bkrequest struct {
		ServId       string `json:"servid" validate:"required"`
		IntervalTime string `json:"timeInter" validate:"required"`
		Servicename  string `json:"servicename" validate:"required"`
	
	}
	Response struct{
		Message string
		Success  bool

	}

)
func ReturnSlot (w http.ResponseWriter, r* http.Request){

}
func bookinglogic(w http.ResponseWriter, r *http.Request) {
     if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid payload",
			Success: false,
		})
		return
	 }
	//  var rs Bkrequest
}
