package booking

import (
	"net/http"
)

type (
	Respond struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
	}
	BookingRequest struct {
		Time       int64  `json:"time" validate:"required"`
		Department string `json:"department" validate:"required"`
		Day        string `json:"day" validate:"required"`
		Diseases   string `json:"desease" validate:"required"`
	}
)

func Booking(w http.ResponseWriter, r *http.Request) {

}
