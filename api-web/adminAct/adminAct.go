package adminact

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	handlerconn "medquemod/db_conn"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
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

	ServiceAssignPayload struct {
		ServiceName string  `json:"serviceName"`
		DurationMin int     `json:"durationMin"`
		Fee         float64 `json:"fee"`
	}
	ServiceAssignPayload2 struct {
		ServiceName   string  `json:"serviceName"`
		InitialNumber int     `json:"initialNumber"`
		Fee           float64 `json:"fee"`
	}
	Service struct {
		ID              int     `json:"id"`
		Servicename     string  `json:"servicename"`
		ConsultationFee float64 `json:"consultationFee"`
		CreatedAt       string  `json:"created_at"`
		DurationMin     int     `json:"duration_minutes"`
	}
	Response struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Data    interface{} `json:"data,omitempty"`
	}
	AsdocschedulePayload struct {
		DoctorID  string `json:"doctor_id"`
		DayOfWeek string `json:"day_of_week"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
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
	BookingToday struct {
		UserId      string `json:"userID"`
		Servicename string `json:"serv_name"`
		Patientname string `json:"patientname"`
		Start_time  string `json:"start"`
		End_time    string `json:"end"`
		Status      string `json:"status"`
		DayOfWeek   string `json:"dayofweek"`
		Service_Id  string
	}

	// Admin login structures
	AdminLoginRequest struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}

	AdminData struct {
		AdminID  int    `json:"admin_id"`
		Username string `json:"username"`
		Role     string `json:"role"`
		Email    string `json:"email"`
	}

	AdminLoginResponse struct {
		Token string    `json:"token"`
		Admin AdminData `json:"admin"`
	}

	DoctorInfoResponse struct {
		DoctorID    int    `json:"doctor_id"`
		Doctorname  string `json:"doctorname"`
		Email       string `json:"email"`
		Specialist  string `json:"specialist_name"`
		Phone       string `json:"phone"`
		AssgnStatus bool   `json:"assgn_status"`
	}
)


type AdminClaims struct {
	AdminID  int    `json:"admin_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}


func generateAdminToken(adminID int, username, role, email string) (string, error) {
	secretKey := []byte("admin-secret-key-here")

	claims := AdminClaims{
		AdminID:  adminID,
		Username: username,
		Role:     role,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

const (
	ADMIN_USERNAME = "admin"
	ADMIN_PASSWORD = "admin123"
	ADMIN_EMAIL    = "admin@medque.com"
	ADMIN_ID       = 1
	ADMIN_ROLE     = "admin"
)

func AdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Method not allowed",
			Success: false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var adminLogin AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&adminLogin); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid request body",
			Success: false,
		})
		return
	}

	if adminLogin.Username == "" || adminLogin.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: "Username and password are required",
			Success: false,
		})
		return
	}

	if adminLogin.Username != ADMIN_USERNAME || adminLogin.Password != ADMIN_PASSWORD {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid username or password",
			Success: false,
		})
		return
	}

	token, err := generateAdminToken(ADMIN_ID, ADMIN_USERNAME, ADMIN_ROLE, ADMIN_EMAIL)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Failed to generate authentication token",
			Success: false,
		})
		return
	}

	adminData := AdminData{
		AdminID:  ADMIN_ID,
		Username: ADMIN_USERNAME,
		Role:     ADMIN_ROLE,
		Email:    ADMIN_EMAIL,
	}

	loginResponse := AdminLoginResponse{
		Token: token,
		Admin: adminData,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Message: "Admin login successful",
		Success: true,
		Data:    loginResponse,
	})
}

func AssignNonTimeserv(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"message":"Invalid method","success":false}`, http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[AssignNonTimeserv] read body error: %v\n", err)
		http.Error(w, `{"message":"Cannot read body","success":false}`, http.StatusBadRequest)
		return
	}
	log.Printf("[AssignNonTimeserv] raw JSON: %s\n", string(body))

	var req ServiceAssignPayload2
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("[AssignNonTimeserv] JSON decode error: %v\n", err)
		http.Error(w, `{"message":"Bad JSON","success":false}`, http.StatusBadRequest)
		return
	}
	log.Printf("[AssignNonTimeserv] decoded struct: %+v\n", req)

	// --- 3) Check for missing/invalid fields yourself ---
	if strings.TrimSpace(req.ServiceName) == "" {
		log.Println("[AssignNonTimeserv] validation error: servname empty")
		http.Error(w, `{"message":"Service name required","success":false}`, http.StatusBadRequest)
		return
	}
	if req.InitialNumber <= 0 {
		log.Println("[AssignNonTimeserv] validation error: initial_number <= 0")
		http.Error(w, `{"message":"Initial number must >0","success":false}`, http.StatusBadRequest)
		return
	}
	if req.Fee <= 0 {
		log.Println("[AssignNonTimeserv] validation error: fee <= 0")
		http.Error(w, `{"message":"Fee must >0","success":false}`, http.StatusBadRequest)
		return
	}

	// --- 4) Check existence and insert ---
	var existing string
	err = handlerconn.Db.QueryRow(`SELECT servicename FROM serviceavailable_tb WHERE servicename=$1`, req.ServiceName).Scan(&existing)

	switch {
	case err == sql.ErrNoRows:
		// does not exist → try insert
		if _, err := handlerconn.Db.Exec(
			`INSERT INTO serviceavailable_tb (servicename, initial_number, fee) VALUES ($1,$2,$3)`,
			req.ServiceName, req.InitialNumber, req.Fee,
		); err != nil {
			log.Printf("[AssignNonTimeserv] DB insert error: %v\n", err)
			http.Error(w, `{"message":"Internal server error: `+err.Error()+`","success":false}`, http.StatusInternalServerError)
			return
		}
		log.Printf("[AssignNonTimeserv] inserted: %q\n", req.ServiceName)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Response{
			Message: "Non-time service created",
			Success: true,
			Data:    req.ServiceName,
		})
		return

	case err != nil:
		// some other query error
		log.Printf("[AssignNonTimeserv] DB query error: %v\n", err)
		http.Error(w, `{"message":"Database error","success":false}`, http.StatusInternalServerError)
		return

	default:
		// already exists
		log.Printf("[AssignNonTimeserv] already exists: %q\n", existing)
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(Response{
			Message: "Service already exists",
			Success: false,
			Data:    existing,
		})
		return
	}
}

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
	var Assreq ServiceAssignPayload
	json.NewDecoder(r.Body).Decode(&Assreq)
	var servname string
	checkifexisterr := handlerconn.Db.QueryRow(`SELECT servicename FROM serviceavailable WHERE servicename = $1`, Assreq.ServiceName).Scan(&servname)
	if checkifexisterr != nil {
		if checkifexisterr == sql.ErrNoRows {
			_, queryerr := handlerconn.Db.Exec(`INSERT INTO serviceavailable (servicename,duration_minutes, fee) VALUES($1,$2,$3)`, Assreq.ServiceName, Assreq.DurationMin, Assreq.Fee)
			if queryerr != nil {
				json.NewEncoder(w).Encode(Response{
					Message: "Internal serverError: " + queryerr.Error(),
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
		Data:    Assreq.ServiceName,
	})

}

func Asdocschedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Badrequest",
			Success: false,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var shereq AsdocschedulePayload

	json.NewDecoder(r.Body).Decode(&shereq)
	fmt.Printf("Decoded schedule: %+v\n", shereq)
	if strings.TrimSpace(shereq.DoctorID) == "" ||
		strings.TrimSpace(shereq.DayOfWeek) == "" ||
		strings.TrimSpace(shereq.StartTime) == "" ||
		strings.TrimSpace(shereq.EndTime) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: "All fields are required and must not be empty",
			Success: false,
		})
		return
	}

	doctorIDInt, err := strconv.Atoi(shereq.DoctorID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: "DoctorID must be a valid integer",
			Success: false,
		})
		return
	}
	dayOfWeekInt, err := strconv.Atoi(shereq.DayOfWeek)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: "DayOfWeek must be a valid integer",
			Success: false,
		})
		return
	}

	var isExist bool
	errcheck := handlerconn.Db.QueryRow(`SELECT EXISTS(SELECT 1 FROM doctors WHERE doctor_id = $1)`, doctorIDInt).Scan(&isExist)
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
	_, errexec := handlerconn.Db.Exec(`INSERT INTO doctorshedule (doctor_id, day_of_week, start_time, end_time) VALUES($1,$2,$3,$4)`, doctorIDInt, dayOfWeekInt, shereq.StartTime, shereq.EndTime)
	if errexec != nil {
		fmt.Println("SQL Insert Error:", errexec)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: fmt.Sprintf("Something went wrong: %v", errexec),
			Success: false,
		})
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
	fmt.Println(specialists)
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
	if isExist {
		json.NewEncoder(w).Encode(Response{
			Message: "Request is Already exist in table",
			Success: false,
		})
		return
	} else {
		_, execErr := handlerconn.Db.Exec(`INSERT INTO doctor_services (doctor_id, service_id) VALUES($1,$2) `, reqAsgn.DoctorID, reqAsgn.Service_id)
		if execErr != nil {
			json.NewEncoder(w).Encode(Response{
				Message: "Something went wrong",
				Success: false,
			})
			fmt.Println("something went wrong", execErr)
			return
		}
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

// add another function that will get doctor information
func GetDoctorInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Method not allowed",
			Success: false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Query to get doctor information with required fields
	rows, err := handlerconn.Db.Query(`
		SELECT doctor_id, doctorname, email, specialist_name, phone, assgn_status 
		FROM doctors 
		WHERE assgn_status = false
		ORDER BY doctor_id
	`)
	if err != nil {
		log.Printf("Database query error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Failed to fetch doctor information",
			Success: false,
		})
		return
	}
	defer rows.Close()

	var doctors []DoctorInfoResponse
	for rows.Next() {
		var doctor DoctorInfoResponse
		err := rows.Scan(
			&doctor.DoctorID,
			&doctor.Doctorname,
			&doctor.Email,
			&doctor.Specialist,
			&doctor.Phone,
			&doctor.AssgnStatus,
		)
		if err != nil {
			log.Printf("Error scanning doctor row: %v", err)
			continue
		}
		doctors = append(doctors, doctor)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Error processing doctor data",
			Success: false,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Message: "Doctor information retrieved successfully",
		Success: true,
		Data:    doctors,
	})
}

func GetsevAvailable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Method not allowed",
			Success: false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	rows, err := handlerconn.Db.Query("SELECT serv_id, servicename, duration_minutes, fee, created_at FROM  serviceAvailable")
	if err != nil {
		log.Printf("Database query error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Server Error: Failed to fetch data",
			Success: false,
		})
		return
	}
	defer rows.Close()

	var services []Service

	for rows.Next() {
		var service Service
		err := rows.Scan(
			&service.ID,
			&service.Servicename,
			&service.DurationMin,
			&service.ConsultationFee,
			&service.CreatedAt,
		)
		if err != nil {
			log.Printf("Row scanning error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{
				Message: "Server Error: Failed to process data",
				Success: false,
			})
			return
		}
		services = append(services, service)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Server Error: Data processing failed",
			Success: false,
		})
		return
	}

	json.NewEncoder(w).Encode(Response{
		Message: "Successfully returned data",
		Success: true,
		Data:    services,
	})
}
func GetbookingToday(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Invaild payload",
			Success: false,
		})
		return
	}
	now := time.Now()
	year, month, day := now.Date()
	datestring := fmt.Sprintf("%04d-%02d-%02d", year, month, day)

	query := `SELECT u.fullname, b.start_time, b.end_time, s.servicename, b.status
		FROM bookingtrack_tb b
		JOIN users u ON b.user_id = u.user_id
		JOIN serviceavailable s ON b.service_id = s.serv_id
		WHERE b.booking_date = $1`

	rows, err := handlerconn.Db.Query(query, datestring)
	if err != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "Internal serverError",
			Success: false,
		})
		return
	}
	defer rows.Close()

	var bookings []map[string]interface{}
	for rows.Next() {
		var patient, servicename, status string
		var start_time, end_time string
		err := rows.Scan(&patient, &start_time, &end_time, &servicename, &status)
		if err != nil {
			continue
		}
		bookings = append(bookings, map[string]interface{}{
			"patient":     patient,
			"start_time":  start_time,
			"end_time":    end_time,
			"servicename": servicename,
			"status":      status,
		})
	}

	json.NewEncoder(w).Encode(Response{
		Message: "Successfully returned today's bookings",
		Success: true,
		Data:    bookings,
	})
}
