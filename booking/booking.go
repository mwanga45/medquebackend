package booking

// here  one device will be used to make booking to day for adult  for child  it will be almost 3 child
import (
	"encoding/json"
	"fmt"
	handlerconn "medquemod/db_conn"
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
		Doctor     string `json:"doctor" validate:"required"`
	}
	DeviceId struct{
		DeviceId string `json:"deviceId"`
	}
)

func Booking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Respond{
			Message: "Method isnt Allowed",
			Success: false,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	tx, errTx := handlerconn.Db.Begin()

	if errTx != nil {
		json.NewEncoder(w).Encode(Respond{
			Message: "Something went wrong Transaction Failed",
			Success: false,
		})
	}
	defer tx.Rollback()

}

func handleGeust(username string, nationality string, deviceid string, time int, department string, day string, doctor string, diseases string)error {
	tx, errTx := handlerconn.Db.Begin()
	if errTx != nil {
		return fmt.Errorf("something went wrong here Transaction failed: %w",errTx)
	}
	var idExist bool
    err:= tx.QueryRow("SELECT EXISTS(SELECT 1 FROM Users WHERE deviceid = $1)",deviceid).Scan(&idExist)

	if err !=nil{
		return fmt.Errorf("something went wrong: %w",err)
	}
	if !idExist {
		return fmt.Errorf("user doesnt exist on system")
	}
	
	return nil

}
func handlechild() {

}
func handleshareDevice() {

}
