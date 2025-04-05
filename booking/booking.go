package booking

import "net/http"

type (
	Respond struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
	}
)

func Booking(w http.ResponseWriter, r *http.Request) {

}
