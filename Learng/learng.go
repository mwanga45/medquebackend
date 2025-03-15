package learng

import (
	"encoding/json"
	handlerconn "medquemod/db_conn"
	"net/http"
)

type Response struct {
	Message string      `json:"message,omitempty"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

func Doctors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet{
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "The method used is invalid",
			Success: false,
		})
		return
	}
	row,err := handlerconn.Db.Query("SELECT * FROM doctors")
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "failed to fetch data to the database",
			Success: false,
		})
		return
	}

	var doctors []map[string]interface {}
	columns,_ := row.Columns()
	count := len(columns)
	value := make([]interface{}, count)
	valueptrs := make ([]interface{},count)
   

	for row.Next(){
		for i := range columns{
			valueptrs[i] = &value[i]
		}
		row.Scan(valueptrs...)
		 doctor :=make(map[string]interface{})
		for i,col:= range columns{
			val := value[i]
			doctor[col] = val
		}
		doctors = append(doctors, doctor)
	}

}