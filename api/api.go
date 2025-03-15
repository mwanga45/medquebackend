package api

import (
	"encoding/json"
	handlerconn "medquemod/db_conn"
	"net/http"
)

type Response struct {
	Messsage string      `json:"message"`
	Success  bool        `json:"success,omitempty"`
	Data     interface{} `json:"Data" `
}

func Doctors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Messsage: "Invalid method used ",
			Success:  false,
		})
		return
	}

	rows, err := handlerconn.Db.Query("SELECT * FROM doctors")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Messsage: "database failed to fetch data",
			Success:  false,
		})
		return
	}
	// create are slice which will hold the row as value and column name as key at end
	var doctors []map[string]interface{}
	// store columns name variable
	columns, _ := rows.Columns()
	count := len(columns)
	// create slice for store temporarly  column as key and data as value with corresponding data type of column
	value := make([]interface{}, count)
	// create  slice pointer neccessary during scan value data to corresponding column name
	valueptrs := make([]interface{}, count)
	// automatic move cursor to next rows
	for rows.Next() {
		for i := range columns {
			valueptrs[i] = &value[i]
		}
		rows.Scan(valueptrs...)
		// create another slice that will hold  single row data as value and also column name  as key after each loop
		doctor := make(map[string]interface{})
		for i, col := range columns {
			val := value[i]
			doctor[col] = val
		}
		doctors = append(doctors, doctor)
	}
	// handling rows error
	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Messsage: "data failed to be proccessed",
			Success:  false,
		})
		return
	}

	w.Header().Set("content-type", "application/json")
	err = json.NewEncoder(w).Encode(Response{
		Messsage: "successfuly fetch data",
		Success:  true,
		Data:     doctors,
	})
	// handling encoding process if error occur
	if err != nil {
		json.NewEncoder(w).Encode(Response{
			Messsage: "failed to encode data ",
			Success:  false,
		})
		return
	}
}
