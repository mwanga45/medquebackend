package handlerconn

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var Db *sql.DB

// initilaze connection pool
func Connectionpool(databasesourceName string) error {
	var err error
	if Db, err = sql.Open("postgres", databasesourceName); err != nil {
		return err
	}

	// configuration of new connection pool
	Db.SetMaxOpenConns(25)
	Db.SetConnMaxIdleTime(25)
	Db.SetConnMaxLifetime(5 * time.Minute)

	// Db.Ping used to verify if the connection is alive and properly configured
	if err = Db.Ping(); err != nil {
		return err
	}

	doctor_tb := `CREATE TABLE IF NOT EXISTS doctors (
		doctor_id SERIAL PRIMARY KEY,
		full_name VARCHAR(100) NOT NULL,
		specialty VARCHAR(100) NOT NULL,
		years_experience INT,
		department VARCHAR(100),
		phone_number VARCHAR(20),
		email VARCHAR(100),
		availability VARCHAR(50), 
		time_available VARCHAR(50), -- newly added column
		room_number VARCHAR(20),
		profile_picture VARCHAR(200),
		consultation_fee DECIMAL(10,2),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	
	if _, err = Db.Exec(doctor_tb); err != nil {
		log.Fatalf("failed to create new table %v", err)
	}
	doctors_status := `CREATE TABLE IF NOT EXISTS doctor_status (
		status_id SERIAL PRIMARY KEY,
		doctor_id INT REFERENCES doctors(doctor_id) ON DELETE CASCADE,
		full_name VARCHAR(100) NOT NULL,
		specialty VARCHAR(100) NOT NULL,
		time_available VARCHAR(50),
		availability BOOLEAN DEFAULT FALSE,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err = Db.Exec(doctors_status); err !=nil{
		log.Fatalf("failed to create new table %v",err)
	}
	
	doctors := []struct {
        fullName        string
        specialty       string
        experience      int
        department      string
        phoneNumber     string
        email           string
        availability    string
        timeAvailable   string
        roomNumber      string
        profilePicture  string
        consultationFee float64
    }{
        {
            "Dr. John Doe", "Cardiology", 10, "Heart Department",
            "123-456-7890", "johndoe@example.com", "Available", "09:00 AM - 05:00 PM",
            "Room 101", "profile1.jpg", 150.00,
        },
        {
            "Dr. Sarah Smith", "Neurology", 8, "Neuro Dept",
            "987-654-3210", "sarahsmith@example.com", "Not Available", "10:00 AM - 04:00 PM",
            "Room 202", "profile2.jpg", 180.00,
        },
        {
            "Dr. Ahmed Karim", "Pediatrics", 5, "Children Ward",
            "555-222-1111", "ahmedkarim@example.com", "Available", "08:00 AM - 02:00 PM",
            "Room 303", "profile3.jpg", 130.00,
        },
        {
            "Dr. Emily Zhang", "Dermatology", 12, "Skin Dept",
            "444-666-8888", "emilyzhang@example.com", "Available", "11:00 AM - 06:00 PM",
            "Room 404", "profile4.jpg", 160.00,
        },
        {
            "Dr. Michael Lee", "Orthopedics", 15, "Ortho Dept",
            "333-777-9999", "michaellee@example.com", "Not Available", "07:00 AM - 03:00 PM",
            "Room 505", "profile5.jpg", 170.00,
        },
    }
    insertQuery := `INSERT INTO doctors (full_name, specialty, experience, department, phone_number, email, availability, time_available, room_number, profile_picture, consultation_fee) VALUES (:full_name, :specialty, :experience, :department, :phone_number, :email, :availability, :time_available, :room_number, :profile_picture, :consultation_fee)`
    for _, doc := range doctors {
        _, err := Db.Exec(insertQuery, doc.fullName, doc.specialty, doc.experience, doc.department, doc.phoneNumber, doc.email, doc.availability, doc.timeAvailable, doc.roomNumber, doc.profilePicture, doc.consultationFee)
        if err != nil {
            fmt.Printf("Error inserting %s: %v\n", doc.fullName, err)
        } else {
            fmt.Printf("Inserted %s successfully.\n", doc.fullName)
        }
    }
	statusInsertQuery := `
    INSERT INTO doctor_status (
        doctor_id, full_name, specialty, time_available, availability
    ) VALUES ($1, $2, $3, $4, $5)
`

statusData := []struct {
    fullName      string
    specialty     string
    timeAvailable string
    available     bool
}{
    {"Dr. John Doe", "Cardiology", "09:00 AM - 05:00 PM", true},
    {"Dr. Sarah Smith", "Neurology", "10:00 AM - 04:00 PM", false},
    {"Dr. Ahmed Karim", "Pediatrics", "08:00 AM - 02:00 PM", true},
    {"Dr. Emily Zhang", "Dermatology", "11:00 AM - 06:00 PM", true},
    {"Dr. Michael Lee", "Orthopedics", "07:00 AM - 03:00 PM", false},
}
for _, d := range statusData {
    // Fetch doctor_id using full_name
    var doctorID int
    err := Db.QueryRow("SELECT doctor_id FROM doctors WHERE full_name = $1", d.fullName).Scan(&doctorID)
    if err != nil {
        log.Printf("Failed to find doctor_id for %s: %v", d.fullName, err)
        continue
    }

    // Insert into doctor_status
    _, err = Db.Exec(statusInsertQuery, doctorID, d.fullName, d.specialty, d.timeAvailable, d.available)
    if err != nil {
        log.Printf("Failed to insert status for %s: %v", d.fullName, err)
    } else {
        fmt.Printf("Inserted status for %s successfully.\n", d.fullName)
    }
}
	user_tb := `CREATE TABLE IF NOT EXISTS Users (
		user_id  SERIAL  PRIMARY KEY ,
		full_name VARCHAR(150) NOT NULL,
		home_address VARCHAR(150),
		email VARCHAR(100) UNIQUE NOT NULL,
		phone_number VARCHAR(20) UNIQUE,
		deviceId VARCHAR(200)UNIQUE,
		Age VARCHAR(20),
		user_type VARCHAR(20) CHECK (user_type IN ('Patient')),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _,err = Db.Exec(user_tb);err !=nil{
		log.Fatalf("failed to create table patient_tb %v", err)
	}

	scheduled_notificationstb := `CREATE TABLE IF NOT EXISTS scheduled_notifications (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) NOT NULL,
		device_id VARCHAR(255) NOT NULL,
		notification_time TIMESTAMPTZ NOT NULL,
		booking_time TIMESTAMPTZ NOT NULL,
		status VARCHAR(20) DEFAULT 'pending',
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);`
	if _,err = Db.Exec(scheduled_notificationstb); err != nil{
		log.Fatalf("failed to create table sheduled notification table:%v",err)
	}
	return nil

}


