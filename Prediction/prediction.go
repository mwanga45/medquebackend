package prediction

import (
	"encoding/json"
	handlerconn "medquemod/db_conn"
	"net/http"
)

type ServiceDayPrediction struct {
	ServiceID   int     `json:"service_id"`
	ServiceName string  `json:"service_name"`
	DayOfWeek   int     `json:"day_of_week"`
	DayName     string  `json:"day_name"`
	Predicted   float64 `json:"predicted"`
	Actual      int     `json:"actual"`
}

type PredictionResponse struct {
	Message string                 `json:"message"`
	Success bool                   `json:"success"`
	Data    []ServiceDayPrediction `json:"data"`
}

var dayNames = []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}


func PredictionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Query all services
	rows, err := handlerconn.Db.Query(`SELECT serv_id, servicename FROM serviceavailable`)
	if err != nil {
		json.NewEncoder(w).Encode(PredictionResponse{
			Message: "Failed to fetch services",
			Success: false,
		})
		return
	}
	defer rows.Close()

	var predictions []ServiceDayPrediction
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		for day := 0; day <= 6; day++ {

			var lastDate, prevDate string
			errDates := handlerconn.Db.QueryRow(`
				SELECT MAX(booking_date) FROM bookingtrack_tb WHERE dayofweek = $1
			`, day).Scan(&lastDate)
			if errDates != nil || lastDate == "" {
				predictions = append(predictions, ServiceDayPrediction{
					ServiceID:   id,
					ServiceName: name,
					DayOfWeek:   day,
					DayName:     dayNames[day],
					Predicted:   0,
					Actual:      0,
				})
				continue
			}
			errPrev := handlerconn.Db.QueryRow(`
				SELECT MAX(booking_date) FROM bookingtrack_tb WHERE dayofweek = $1 AND booking_date < $2
			`, day, lastDate).Scan(&prevDate)
			if errPrev != nil {
				prevDate = ""
			}

			var actual int
			errActual := handlerconn.Db.QueryRow(`
				SELECT COUNT(*) FROM bookingtrack_tb
				WHERE service_id = $1 AND dayofweek = $2 AND booking_date = $3
			`, id, day, lastDate).Scan(&actual)
			if errActual != nil {
				actual = 0
			}
			// Predicted: average bookings for this service and day between prevDate (exclusive) and lastDate (inclusive)
			var predicted float64
			if prevDate != "" {
				errPred := handlerconn.Db.QueryRow(`
					SELECT AVG(cnt) FROM (
						SELECT COUNT(*) as cnt
						FROM bookingtrack_tb
						WHERE service_id = $1 AND dayofweek = $2 AND booking_date > $3 AND booking_date <= $4
						GROUP BY booking_date
					) as daily_counts
				`, id, day, prevDate, lastDate).Scan(&predicted)
				if errPred != nil {
					predicted = 0
				}
			} else {
				predicted = float64(actual)
			}
			predictions = append(predictions, ServiceDayPrediction{
				ServiceID:   id,
				ServiceName: name,
				DayOfWeek:   day,
				DayName:     dayNames[day],
				Predicted:   predicted,
				Actual:      actual,
			})
		}
	}

	json.NewEncoder(w).Encode(PredictionResponse{
		Message: "Prediction data fetched successfully",
		Success: true,
		Data:    predictions,
	})
}
