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
	columns, _ := rows.Columns()
	count := len(columns)
	value := make([]interface{}, count)
	valueptrs := make([]interface{}, count)

	for rows.Next(){
		for i := range columns{
			valueptrs[i] = &value[i]
		}
		rows.Scan(valueptrs...)
	}

	// create another slice that will hold  single row data as value and also column name  as key after each loop  
	doctor := make( map[string]interface{})
  for i , col := range columns{
	val := value[i]
	doctor[col] = val
  }
  doctors = append(doctors, doctor)

}
