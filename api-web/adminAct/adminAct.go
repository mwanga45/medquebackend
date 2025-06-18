package adminact

import (
	"database/sql"
	"encoding/json"
	"fmt"
	// sidefunc_test "medquemod/booking/verification"
	handlerconn "medquemod/db_conn"
	"net/http"
)

type (
	DoctorInfo struct {
		DoctorID   int    `json:"doctor_id"`
		Doctorname string `json:"doctorname"`
	}

	ServiceInfo struct {
		ServiceID int    `json:"serv_id"`
		Servname  string `json:"servicename"`
	}

	CombinedResponse struct {
		Doctors  []DoctorInfo  `json:"doctors"`
		Services []ServiceInfo `json:"services"`
	}

	ServAssignpayload struct {
		Servname    string  `json:"servname"`
		DurationMin int     `json:"duration_time"`
		Fee         float64 `json:"fee"`
	}

	Response struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Data    interface{} `json:"data,omitempty"`
	}
	Asdocshedulepayload struct {
		DoctorID  string `json:"doctor_id" validate:"required"`
		Dayofweek string `json:"day_of_week" validate:"required"`
		StartTime string `json:"start_time" validate:"required"`
		EndTime   string `json:"end_time" validate:"required"`
	}
	AsdocServ struct {
		DoctorID  string `json:"docID"`
		ServiceID string `json:"servID"`
	}
	SpecialistPayload struct {
		Specialist  string `json:"specialist"`
		Description string `json:"description"`
	}
	DocServPayload struct {
		DoctorID   string `json:"doctor_id" validate:"required"`
		Service_id string `json:"serviceid" validate:"required"`
	}
	DocVSServLs struct {
		DoctorID   string `json:"doctor_id"`
		ServiceID  string `json:"serv_id"`
		Servname   string `json:"servname"`
		Doctorname string `json:"doctorname"`
	}
)

func AssignService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(Response{
			Message: "Invalidpayload",
			Success: false,
		})
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var Assreq ServAssignpayload
	json.NewDecoder(r.Body).Decode(&Assreq)
	var servname string
	checkifexisterr := handlerconn.Db.QueryRow(`SELECT servicename FROM serviceAvailable  WHERE servicename = $1`, Assreq.Servname).Scan(&servname)
	if checkifexisterr != nil {
		if checkifexisterr == sql.ErrNoRows {
			_, queryerr := handlerconn.Db.Exec(`INSERT INTO serviceAvailable (servicename,duration_minutes, fee) VALUES($1,$2,$3)`, Assreq.Servname, Assreq.DurationMin, Assreq.Fee)
			if queryerr != nil {
				json.NewEncoder(w).Encode(Response{
					Message: "Internal serverError",
					Success: false,
				})
				fmt.Println("failed to insertdata", queryerr)
				return
			}
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "failed to assign new service",
			Success: false,
		})
		return
	}
	json.NewEncoder(w).Encode(Response{
		Message: "Servecename is already exist",
		Success: false,
		Data:    Assreq.Servname,
	})

}

func Asdocshedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Badrequest",
			Success: false,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var shereq Asdocshedulepayload

	json.NewDecoder(r.Body).Decode(&shereq)
	var isExist bool

	errcheck := handlerconn.Db.QueryRow(`SELECT EXISTS(SELECT 1 FROM doctors WHERE doctor_id = $1)`, shereq.DoctorID).Scan(&isExist)
	if errcheck != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Internal serverError",
			Success: false,
		})
		fmt.Println("failed to execute query", errcheck)
	}
	if !isExist {
		json.NewEncoder(w).Encode(Response{
			Message: " Staff(doctor) is not yet exist in the system",
			Success: false,
			Data:    shereq.DoctorID,
		})
		return
	}
	_, errexec := handlerconn.Db.Exec(`INSERT INTO doctorshedule (doctor_id, day_of_week, start_time, end_time) VALUES($1,$2,$3,$4)`, shereq.DoctorID, shereq.Dayofweek, shereq.StartTime, shereq.EndTime)
	if errexec != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Something went wrong",
			Success: false,
		})
		fmt.Println("something went wrong", errexec)
		return
	}
	json.NewEncoder(w).Encode(Response{
		Message: "Successfuly assign new staff shedule",
		Success: true,
		Data:    shereq.DoctorID,
	})
}

func AssignDocService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Badrequest",
			Success: false,
		})
		return
	}
	var asservreq AsdocServ
	var isdocExist, isservExist bool
	json.NewDecoder(r.Body).Decode(&asservreq)

	checker := handlerconn.Db.QueryRow(`SELECT 1 FROM doctors WHERE doctor_id = $1`, asservreq.DoctorID).Scan(&isdocExist)
	if checker != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Internal ServerError",
			Success: false,
		})
		fmt.Println("Failed to check if doc exist", checker)
		return
	}

	checker = handlerconn.Db.QueryRow(`SELECT 1 FROM serviceAvailable WHERE serv_id = $1`, asservreq.ServiceID).Scan(&isservExist)
	if checker != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Internal ServerError",
			Success: false,
		})
		fmt.Println("Failed to check if serv exist", checker)
		return
	}
	if isdocExist || isservExist {
		json.NewEncoder(w).Encode(Response{
			Message: "doctoer or service is not exist",
			Success: false,
		})
		return
	}
	_, errorExec := handlerconn.Db.Exec(`INSERT INTO doctorServ_tb (doctor_id, service_id) VALUES($1,$2)`, asservreq.DoctorID, asservreq.ServiceID)
	if errorExec != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Internal ServerError",
			Success: false,
		})
		fmt.Println("something went wrong", checker)
		return
	}
	json.NewEncoder(w).Encode(Response{
		Message: "Successfuly assign  doctor service",
		Success: true,
		Data:    asservreq,
	})

}

func RegSpecialist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid payload",
			Success: false,
		})
		return
	}
	w.Header().Set("content-type", "application/json")
	var specpayload SpecialistPayload

	json.NewDecoder(r.Body).Decode(&specpayload)

	var specialist string
	queryCheck := handlerconn.Db.QueryRow(`SELECT specialist FROM specialist WHERE specialist = $1 `, specpayload.Specialist).Scan(&specialist)
	if queryCheck != nil {
		if queryCheck == sql.ErrNoRows {

			_, err := handlerconn.Db.Exec(`INSERT INTO specialist (specialist, description) VALUES($1,$2)`, specpayload.Specialist, specpayload.Description)
			if err != nil {
				json.NewEncoder(w).Encode(Response{
					Message: "Internal Service Error",
					Success: false,
				})
				fmt.Println("Something went wrong", err)
				return
			}
		}
		json.NewEncoder(w).Encode(Response{
			Message: "Something went wrong try Again",
			Success: false,
			Data:    specpayload.Specialist,
		})
		fmt.Println("Something went  wrong", queryCheck)
	}

	json.NewEncoder(w).Encode(Response{
		Message: "Your try  to register specilaist  which is already exists",
		Success: false,
		Data:    specpayload.Specialist,
	})

}
func ReturnSpec(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid payload",
			Success: false,
		})
		return
	}
	row, errfetch := handlerconn.Db.Query(`SELECT * FROM specialist `)
	if errfetch != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Internal serverError",
			Success: false,
		})
		fmt.Println("Something went wrong ", errfetch)
	}
	defer row.Close()
	var specialists []SpecialistPayload
	for row.Next() {
		var s SpecialistPayload
		errscan := row.Scan(&s.Specialist, &s.Description)
		if errscan != nil {
			json.NewEncoder(w).Encode(Response{
				Message: "Internal error",
				Success: false,
			})
			fmt.Println("Something went wrong here", errscan)
		}

		specialists = append(specialists, s)
	}
	if err := row.Err(); err != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Internal serverError",
			Success: false,
		})
		fmt.Println("Something went wrong", err)
		return
	}

	json.NewEncoder(w).Encode(Response{
		Message: "Successfuly",
		Success: true,
		Data:    specialists,
	})
}

func DocServAssign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Bad request",
			Success: false,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var reqAsgn DocServPayload

	json.NewDecoder(r.Body).Decode(&reqAsgn)
	query := "SELECT EXISTS(SELECT 1 FROM doctor_services WHERE doctor_id = $1 AND service_id = $2 )"
	var isExist bool
	errcheck := handlerconn.Db.QueryRow(query, reqAsgn.DoctorID, reqAsgn.Service_id).Scan(&isExist)
	if errcheck != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "payload is Already exists",
			Success: false,
		})
		return
	}
	_, execErr := handlerconn.Db.Exec(`INSERT INTO doctor_services (doctor_id, service_id) VALUES($1,$2) `, reqAsgn.DoctorID, reqAsgn.Service_id)
	if execErr != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Something went wrong",
			Success: false,
		})
		fmt.Println("something went wrong", execErr)
		return
	}

	json.NewEncoder(w).Encode(Response{
		Message: "Successfuly assign doctor to service of id",
		Success: true,
		Data:    reqAsgn.Service_id,
	})
}
func DocVsServ(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Only GET is allowed",
			Success: false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")


	doctorRows, err := handlerconn.Db.Query(`SELECT doctor_id, doctorname FROM doctors`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "Error fetching doctors", Success: false})
		return
	}
	defer doctorRows.Close()

	var doctors []DoctorInfo
	for doctorRows.Next() {
		var d DoctorInfo
		if err := doctorRows.Scan(&d.DoctorID, &d.Doctorname); err != nil {
			json.NewEncoder(w).Encode(Response{Message: "Error scanning doctor", Success: false})
			return
		}
		doctors = append(doctors, d)
	}

	serviceRows, err := handlerconn.Db.Query(`SELECT serv_id, servicename FROM serviceAvailable`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Message: "Error fetching services", Success: false})
		return
	}
	defer serviceRows.Close()

	var services []ServiceInfo
	for serviceRows.Next() {
		var s ServiceInfo
		if err := serviceRows.Scan(&s.ServiceID, &s.Servname); err != nil {
			json.NewEncoder(w).Encode(Response{Message: "Error scanning service", Success: false})
			return
		}
		services = append(services, s)
	}


	finalResponse := CombinedResponse{
		Doctors:  doctors,
		Services: services,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Message: "Success",
		Success: true,
		Data:    finalResponse,
	})
}
