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
    room_number VARCHAR(20),
	profile_picture VARCHAR(200),
    consultation_fee DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
` 
	if _, err = Db.Exec(doctor_tb); err != nil {
		log.Fatalf("failed to create new table %v", err)
	}
	query_test := `INSERT INTO doctors (
		full_name,
		specialty,
		years_experience,
		department,
		phone_number,
		email,
		availability,
		room_number,
		profile_picture,
		consultation_fee
	  )
	  VALUES
		('Dr. John Doe', 'Cardiology', 15, 'Cardiology', '123-456-7890', 'johndoe@example.com', 'Mon-Fri 9AM-5PM', '101', 'https://example.com/images/johndoe.jpg', 150.00),
		('Dr. Jane Smith', 'Neurology', 12, 'Neurology', '987-654-3210', 'janesmith@example.com', 'Mon-Fri 8AM-4PM', '102', 'https://example.com/images/janesmith.jpg', 200.00),
		('Dr. Emily Davis', 'Pediatrics', 8, 'Pediatrics', '555-123-4567', 'emilydavis@example.com', 'Tue-Thu 10AM-6PM', '103', 'https://example.com/images/emilydavis.jpg', 100.00);`
	  if _, err := Db.Exec(query_test);err !=nil{
		log.Fatal("failed to insert data")
	  }

	patient_tb := `CREATE TABLE IF NOT EXISTS Patients (
		user_id  SERIAL  PRIMARY KEY ,
		full_name VARCHAR(150) NOT NULL,
		home_address VARCHAR(150),
		email VARCHAR(100) UNIQUE NOT NULL,
		phone_number VARCHAR(20) UNIQUE,
		deviceId VARCHAR(200)UNIQUE,
		user_type VARCHAR(20) CHECK (user_type IN ('Patient')),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _,err = Db.Exec(patient_tb);err !=nil{
		log.Fatalf("failed to create table patient_tb %v", err)
	}

	return nil

}

// create function to terminate connection
// func Closeconn() error {
// 	if Db != nil {
// 		return Db.Close()
// 	}
// 	return nil

// }
