package handlerconn

import (
	"database/sql"
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
		rating varchar(50),
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err = Db.Exec(doctors_status); err !=nil{
		log.Fatalf("failed to create new table %v",err)
	}
	
	
	user_tb := `CREATE TABLE IF NOT EXISTS Users (
		user_id  SERIAL  PRIMARY KEY ,
		fullname VARCHAR(150) NOT NULL,
		Secretekey VARCHAR(200) NOT NULL, 
		home_address VARCHAR(150),
		email VARCHAR(100) UNIQUE NOT NULL,
		dial VARCHAR(20) UNIQUE,
		deviceId VARCHAR(200)UNIQUE NOT NULL,
		Birthdate VARCHAR(200) NOT NULL,
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
	serviceAvailable := `CREATE TABLE IF NOT EXISTS serviceavalable(
	    id SERIAL PRIMARY KEY,
		disease VARCHAR(255) NOT NULL,
		doctor_id INT REFERENCES doctors(doctor_id) ON DELETE CASCADE,
		fullname VARCHAR(255)

	)`
	if _,err = Db.Exec(serviceAvailable); err !=nil{
       log.Fatalf("Failed to create table serviceAvailable :%v ", err)
	}
	

	// data instert it  for sample test  

// doctorsDetails := `INSERT INTO doctors (full_name, specialty, years_experience, department, phone_number, email, availability, time_available, room_number, profile_picture, consultation_fee)
// VALUES 
// ('Dr. Sarah Johnson', 'Cardiologist', 12, 'Cardiology', '123-456-7890', 'sarah.johnson@hospital.com', 'Yes', '09:00 AM - 03:00 PM', 'Room 101', '/images/sarah.jpg', 150.00),
// ('Dr. James Lee', 'Dermatologist', 8, 'Dermatology', '234-567-8901', 'james.lee@hospital.com', 'Yes', '10:00 AM - 04:00 PM', 'Room 102', '/images/james.jpg', 120.00),
// ('Dr. Amina Yusuf', 'Neurologist', 15, 'Neurology', '345-678-9012', 'amina.yusuf@hospital.com', 'Yes', '11:00 AM - 05:00 PM', 'Room 103', '/images/amina.jpg', 200.00),
// ('Dr. David Smith', 'Pediatrician', 10, 'Pediatrics', '456-789-0123', 'david.smith@hospital.com', 'Yes', '08:00 AM - 12:00 PM', 'Room 104', '/images/david.jpg', 100.00),
// ('Dr. Leila Ali', 'Orthopedic', 9, 'Orthopedics', '567-890-1234', 'leila.ali@hospital.com', 'Yes', '01:00 PM - 06:00 PM', 'Room 105', '/images/leila.jpg', 180.00);`
// _, err = Db.Exec(doctorsDetails)
// if err != nil {
// 	log.Fatalf("failed to insert sample doctor data: %v", err)
// }
// status_test := `INSERT INTO doctor_status (doctor_id, full_name, specialty, time_available, rating)
// VALUES 
// (1, 'Dr. Sarah Johnson', 'Cardiologist', '09:00 AM - 03:00 PM','1.2'),
// (2, 'Dr. James Lee', 'Dermatologist', '10:00 AM - 04:00 PM','3.4'),
// (3, 'Dr. Amina Yusuf', 'Neurologist', '11:00 AM - 05:00 PM','4.5'),
// (4, 'Dr. David Smith', 'Pediatrician', '08:00 AM - 12:00 PM','3.6'),
// (5, 'Dr. Leila Ali', 'Orthopedic', '01:00 PM - 06:00 PM','2.9');`

// _,err = Db.Query(status_test)
// if err != nil{
// 	log.Fatalf("failedtto insert data %v", err)
// } 
	return nil

}


