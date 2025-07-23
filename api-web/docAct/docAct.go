package docact

import (
	"encoding/json"
	"fmt"
	"log"
	sidefunc_test "medquemod/booking/verification"
	handlerconn "medquemod/db_conn"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type (
	Response struct {
		Message string `json:"message"`
		Success bool   `json:"success"`
		Data    interface{}
	}
	
	Appointment struct {
		ID        int       `json:"id"`
		Username  string    `json:"username"`
		StartTime string    `json:"start_time"`
		EndTime   string    `json:"end_time"`
		Status    string    `json:"status"`
	}
	
	Patient struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
	}
	
	StaffRegister struct {
		Doctorname  string `json:"username" validate:"required"`
		RegNo       string `json:"regNo" validate:"required"`
		Password    string `json:"password" validate:"required"`
		Confirmpwrd string `json:"confirmpwrd" validate:"required"`
		Specialist  string `json:"specialist" validate:"required"`
		Phone       string `json:"phone" validate:"required"`
		Email       string `json:"email" validate:"required"`
	}
	
	Stafflogin struct {
		RegNo    string `json:"regNo" validate:"required"`
		Password string `json:"password" validate:"required"`
		Username string `json:"username" validate:"required"`
	}
	
	DoctorData struct {
		DoctorID       int    `json:"doctor_id"`
		DoctorName     string `json:"doctorname"`
		Email          string `json:"email"`
		SpecialistName string `json:"specialist_name"`
		Phone          string `json:"phone"`
		Identification string `json:"identification"`
		Role           string `json:"role"`
	}
	
	LoginResponse struct {
		Token  string     `json:"token"`
		Doctor DoctorData `json:"doctor"`
	}
)


type Claims struct {
	DoctorID int    `json:"doctor_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}


func generateToken(doctorID int, username, role string) (string, error) {

	secretKey := []byte("your-secret-key-here")

	claims := Claims{
		DoctorID: doctorID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), 
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func DoctLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Method not allowed",
			Success: false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var staffLogin Stafflogin
	if err := json.NewDecoder(r.Body).Decode(&staffLogin); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid request body",
			Success: false,
		})
		return
	}


	if staffLogin.Username == "" || staffLogin.RegNo == "" || staffLogin.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: "Username, registration number, and password are required",
			Success: false,
		})
		return
	}


	isValid := sidefunc_test.CheckIdentifiaction(staffLogin.RegNo)
	if !isValid {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid registration number format",
			Success: false,
		})
		return
	}


	var doctorData DoctorData
	var hashedPassword string

	err := handlerconn.Db.QueryRow(`
		SELECT doctor_id, doctorname, password, email, specialist_name, phone, identification, role 
		FROM doctors 
		WHERE doctorname = $1 AND identification = $2
	`, staffLogin.Username, staffLogin.RegNo).Scan(
		&doctorData.DoctorID,
		&doctorData.DoctorName,
		&hashedPassword,
		&doctorData.Email,
		&doctorData.SpecialistName,
		&doctorData.Phone,
		&doctorData.Identification,
		&doctorData.Role,
	)

	if err != nil {
		log.Printf("Database query error: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid username or registration number",
			Success: false,
		})
		return
	}

	
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(staffLogin.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid password",
			Success: false,
		})
		return
	}


	token, err := generateToken(doctorData.DoctorID, doctorData.DoctorName, doctorData.Role)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Failed to generate authentication token",
			Success: false,
		})
		return
	}

	fmt.Printf("Token %s logged in successfully\n", token)

	loginResponse := LoginResponse{
		Token:  token,
		Doctor: doctorData,
	}

	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Message: "Login successful",
		Success: true,
		Data:    loginResponse,
	})
}

func Registration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid Payload",
			Success: false,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var stafreg StaffRegister
	json.NewDecoder(r.Body).Decode(&stafreg)

	
	isValid := sidefunc_test.CheckIdentifiaction(stafreg.RegNo)
	if !isValid {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid Payload ",
			Success: false,
		})
		return
	}
	isSame := sidefunc_test.Checkpassword(stafreg.Password, stafreg.Confirmpwrd)
	if !isSame {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: "Confirm password and password are not the same",
			Success: false,
		})
		return
	}
	var isExist bool
	client, errTx := handlerconn.Db.Begin()
	if errTx != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Interna ServerError",
			Success: false,
		})
		fmt.Println("Something went wrong", errTx)
		return
	}
	defer client.Rollback()
	errCheck := client.QueryRow(`SELECT EXISTS(SELECT 1 FROM doctors WHERE email = $1 AND  doctorname = $2 )`, stafreg.Email, stafreg.Doctorname).Scan(&isExist)
	if errCheck != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Internal ServerError",
			Success: false,
		})
		fmt.Println("something went wrong", errCheck)
		return
	}
	var isSpec bool
	errspec := client.QueryRow(`SELECT EXISTS(SELECT 1 FROM specialist WHERE specialist  = $1)`, stafreg.Specialist).Scan(&isSpec)
	if errspec != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Internal ServerError",
			Success: false,
		})
		fmt.Println("something went wrong", errCheck)
		return
	}
	if !isSpec {
		json.NewEncoder(w).Encode(Response{
			Message: "Invalid Payload",
			Success: false,
		})
		return
	}
	errpaswrd := sidefunc_test.ValidateSecretkey(stafreg.Password)
	if errpaswrd != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "password must have 6 character and atleast having one Uppercase and digit ",
			Success: false,
		})
		return
	}
	hashedpwrd, err := bcrypt.GenerateFromPassword([]byte(stafreg.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Internal ServerError",
			Success: false,
		})
		log.Println("something went  wrong", err)
		return
	}
	_, errEXec := client.Exec(`INSERT INTO doctors (doctorname, password, email, specialist_name,phone, identification, role,assgn_status) VALUES($1,$2,$3,$4,$5,$6,$7,$8) `, stafreg.Doctorname, hashedpwrd, stafreg.Email, stafreg.Specialist, stafreg.Phone, stafreg.RegNo, "dkt",false)
	if errEXec != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Internal ServerError",
			Success: false,
		})
		log.Println("something went wrong", errEXec)
		return
	}

	if errCommit := client.Commit(); errCommit != nil {
		json.NewEncoder(w).Encode(Response{
			Message: "System failed to commit request try Again",
			Success: false,
		})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{
		Message: "successfuly registered",
		Success: true,
	})
}
func GetDoctorAppointments(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Stage 1: Enter GetDoctorAppointments")
    w.Header().Set("Content-Type", "application/json")

    // Stage 2: Extract JWT claims
    fmt.Println("Stage 2: Extracting JWT claims")
    claims, ok := r.Context().Value("claims").(jwt.MapClaims)
    if !ok {
        fmt.Println("Stage 2 Error: Unauthorized – no claims")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(Response{Message: "Unauthorized", Success: false})
        return
    }
    fmt.Printf("Stage 2: Claims found: %v\n", claims)

    // Stage 3: Parse doctor_id
    doctorID := int(claims["doctor_id"].(float64))
    fmt.Printf("Stage 3: doctorID = %d\n", doctorID)

    // Stage 4: Get today's date
    today := time.Now().Format("2006-01-02")
    fmt.Printf("Stage 4: today = %s\n", today)

    // Stage 5: Perform DB query
    fmt.Println("Stage 5: Querying appointments from DB")
    rows, err := handlerconn.Db.Query(`
        SELECT bt.id, u.fullname as username, bt.start_time, bt.end_time, bt.status 
        FROM bookingtrack_tb bt
        JOIN users u ON bt.user_id = u.user_id
        WHERE bt.doctor_id = $1 AND bt.booking_date = $2
        ORDER BY bt.start_time
    `, doctorID, today)
    if err != nil {
        fmt.Printf("Stage 5 Error: DB query error: %v\n", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(Response{Message: "Failed to fetch appointments", Success: false})
        return
    }
    defer rows.Close()
    fmt.Println("Stage 5: Query succeeded, iterating rows")

    // Stage 6: Scan rows
    var appointments []Appointment
    for rows.Next() {
        var apt Appointment
        fmt.Println("Stage 6: Scanning next row")
        if err := rows.Scan(&apt.ID, &apt.Username, &apt.StartTime, &apt.EndTime, &apt.Status); err != nil {
            fmt.Printf("Stage 6 Error: Row scan error: %v\n", err)
            continue
        }
        fmt.Printf("Stage 6: Scanned appointment: %+v\n", apt)
        appointments = append(appointments, apt)
    }

    // Stage 7: Write response
    fmt.Println("Stage 7: Encoding response JSON")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(Response{Message: "Successfully fetched appointments", Success: true, Data: appointments})
    fmt.Println("Stage 7: Done GetDoctorAppointments")
}

// UpdateAppointmentStatus updates the status of an appointment
func UpdateAppointmentStatus(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Stage 1: Enter UpdateAppointmentStatus")
    w.Header().Set("Content-Type", "application/json")

    var updateData struct {
        AppointmentID int    `json:"appointment_id"`
        Status        string `json:"status"`
    }

    // Stage 2: Decode body
    fmt.Println("Stage 2: Decoding request body")
    if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
        fmt.Printf("Stage 2 Error: Invalid JSON body: %v\n", err)
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(Response{Message: "Invalid request body", Success: false})
        return
    }
    fmt.Printf("Stage 2: updateData = %+v\n", updateData)

    // Stage 3: Extract JWT claims
    fmt.Println("Stage 3: Extracting JWT claims")
    claims, ok := r.Context().Value("claims").(jwt.MapClaims)
    if !ok {
        fmt.Println("Stage 3 Error: Unauthorized – no claims")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(Response{Message: "Unauthorized", Success: false})
        return
    }
    docID := int(claims["doctor_id"].(float64))
    fmt.Printf("Stage 3: doctorID = %d\n", docID)

    // Stage 4: Execute update
    fmt.Println("Stage 4: Executing DB update")
    _, err := handlerconn.Db.Exec(`
        UPDATE bookingtrack_tb 
        SET status = $1
        WHERE id = $2 AND doctor_id = $3
    `, updateData.Status, updateData.AppointmentID, docID)
    if err != nil {
        fmt.Printf("Stage 4 Error: DB update error: %v\n", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(Response{Message: "Failed to update appointment status", Success: false})
        return
    }

    // Stage 5: Write response
    fmt.Println("Stage 5: Encoding response JSON")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(Response{Message: "Successfully updated appointment status", Success: true})
    fmt.Println("Stage 5: Done UpdateAppointmentStatus")
}

// SearchPatients searches for patients by name
func SearchPatients(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Stage 1: Enter SearchPatients")
    w.Header().Set("Content-Type", "application/json")

    // Stage 2: Get query param
    query := r.URL.Query().Get("q")
    fmt.Printf("Stage 2: query = '%s'\n", query)
    if query == "" {
        fmt.Println("Stage 2 Error: Empty search query")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(Response{Message: "Search query is required", Success: false})
        return
    }

    // Stage 3: Perform DB query
    fmt.Println("Stage 3: Querying patients from DB")
    rows, err := handlerconn.Db.Query(`
        SELECT user_id, fullname, dial, email 
        FROM users 
        WHERE user_type = 'Patient' AND (fullname ILIKE $1 OR dial ILIKE $1)
        LIMIT 10
    `, "%"+query+"%")
    if err != nil {
        fmt.Printf("Stage 3 Error: DB query error: %v\n", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(Response{Message: "Failed to search patients", Success: false})
        return
    }
    defer rows.Close()
    fmt.Println("Stage 3: Query succeeded, iterating patient rows")

    // Stage 4: Scan rows
    var patients []Patient
    for rows.Next() {
        var patient Patient
        fmt.Println("Stage 4: Scanning next patient row")
        if err := rows.Scan(&patient.ID, &patient.Username, &patient.Phone, &patient.Email); err != nil {
            fmt.Printf("Stage 4 Error: Row scan error: %v\n", err)
            continue
        }
        fmt.Printf("Stage 4: Scanned patient: %+v\n", patient)
        patients = append(patients, patient)
    }

    // Stage 5: Write response
    fmt.Println("Stage 5: Encoding response JSON")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(Response{Message: "Successfully searched patients", Success: true, Data: patients})
    fmt.Println("Stage 5: Done SearchPatients")
}

// GetDoctorProfile returns doctor's profile information
func GetDoctorProfile(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Stage 1: Enter GetDoctorProfile")
    w.Header().Set("Content-Type", "application/json")

    // Stage 2: Extract JWT claims
    fmt.Println("Stage 2: Extracting JWT claims")
    claims, ok := r.Context().Value("claims").(jwt.MapClaims)
    if !ok {
        fmt.Println("Stage 2 Error: Unauthorized – no claims")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(Response{Message: "Unauthorized", Success: false})
        return
    }
    doctorID := int(claims["doctor_id"].(float64))
    fmt.Printf("Stage 2: doctorID = %d\n", doctorID)

    // Stage 3: Query profile
    fmt.Println("Stage 3: Querying doctor profile from DB")
    var doctorData DoctorData
    err := handlerconn.Db.QueryRow(`
        SELECT doctor_id, doctorname, email, specialist_name, phone, identification, role 
        FROM doctors 
        WHERE doctor_id = $1
    `, doctorID).Scan(
        &doctorData.DoctorID,
        &doctorData.DoctorName,
        &doctorData.Email,
        &doctorData.SpecialistName,
        &doctorData.Phone,
        &doctorData.Identification,
        &doctorData.Role,
    )
    if err != nil {
        fmt.Printf("Stage 3 Error: DB query error: %v\n", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(Response{Message: "Failed to fetch doctor profile", Success: false})
        return
    }
    fmt.Printf("Stage 3: doctorData = %+v\n", doctorData)

    // Stage 4: Write response
    fmt.Println("Stage 4: Encoding response JSON")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(Response{Message: "Successfully fetched doctor profile", Success: true, Data: doctorData})
    fmt.Println("Stage 4: Done GetDoctorProfile")
}
