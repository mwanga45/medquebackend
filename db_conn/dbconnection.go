package handlerconn

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var Db *sql.DB

// initilaze connection pool
func Connectionpool() error {
	var err error
	mode := os.Getenv("DB_MODE")
	if mode == "cloud" {
		databaseURL := os.Getenv("DATABASE_URL_CLOUD")
		if databaseURL == "" {
			return fmt.Errorf("DATABASE_URL_CLOUD environment variable is not set")
		}
		log.Printf("[DB] Using cloud database: %s", databaseURL)
		if Db, err = sql.Open("postgres", databaseURL); err != nil {
			log.Printf("Failed to open cloud database connection: %v", err)
			return err
		}
		if err = Db.Ping(); err != nil {
			log.Printf("Failed to ping cloud database: %v", err)
			return err
		}
		log.Println("Successfully connected to cloud database")
	} else {
		databaseURL := os.Getenv("DATABASE_URL")
		if databaseURL == "" {
			return fmt.Errorf("DATABASE_URL environment variable is not set")
		}
		log.Printf("[DB] Using local database: %s", databaseURL)
		if Db, err = sql.Open("postgres", databaseURL); err != nil {
			log.Printf("Failed to open local database connection: %v", err)
			return err
		}
		if err = Db.Ping(); err != nil {
			log.Printf("Failed to ping local database: %v", err)
			return err
		}
		log.Println("Successfully connected to local database")
	}

	Db.SetMaxOpenConns(25)
	Db.SetConnMaxIdleTime(25)
	Db.SetConnMaxLifetime(5 * time.Minute)

	specialist := `CREATE TABLE IF NOT EXISTS specialist (
  specialist VARCHAR(200) PRIMARY KEY,
  description TEXT
);`
	if _, err = Db.Exec(specialist); err != nil {
		log.Printf("Failed to create specialist table: %v", err)
		return err
	}
	log.Println("Table 'specialist' created or already exists.")

	doctor_tb := `
  CREATE TABLE IF NOT EXISTS doctors (
    doctor_id SERIAL PRIMARY KEY,
    doctorname VARCHAR(250) NOT NULL UNIQUE,
    password VARCHAR(250) NOT NULL,
    email VARCHAR(250) NOT NULL UNIQUE,
    specialist_name VARCHAR(200),
    phone VARCHAR(20),
    identification VARCHAR(250) NOT NULL UNIQUE,
    role VARCHAR(20) DEFAULT 'dkt',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_specialist_name
      FOREIGN KEY (specialist_name)
      REFERENCES specialist(specialist)
      ON UPDATE CASCADE
      ON DELETE SET NULL
  )`
	if _, err = Db.Exec(doctor_tb); err != nil {
		log.Fatalf("failed to create new table %v", err)
	}
	log.Println("Table 'doctors' created or already exists.")

	const doctorShedule = `
  CREATE TABLE IF NOT EXISTS doctorshedule (
    shedule_id SERIAL PRIMARY KEY,
    doctor_id INTEGER REFERENCES doctors(doctor_id),
    day_of_week INTEGER NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  )`
	if _, err = Db.Exec(doctorShedule); err != nil {
		log.Fatalf("failed to create new table %v", err)
	}
	log.Println("Table 'doctorshedule' created or already exists.")

	user_tb := `CREATE TABLE IF NOT EXISTS users (
  user_id SERIAL PRIMARY KEY,
  fullname VARCHAR(150) NOT NULL,
  secretekey VARCHAR(200) NOT NULL,
  home_address VARCHAR(150),
  email VARCHAR(100) UNIQUE NOT NULL,
  dial VARCHAR(20) UNIQUE,
  deviceid VARCHAR(200) UNIQUE NOT NULL,
  birthdate VARCHAR(200) NOT NULL,
  user_type VARCHAR(20) CHECK (user_type IN ('Patient')),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`
	if _, err = Db.Exec(user_tb); err != nil {
		log.Fatalf("failed to create table users %v", err)
	}
	log.Println("Table 'users' created or already exists.")

	serviceAvailable := `
  CREATE TABLE IF NOT EXISTS serviceavailable (
    serv_id SERIAL PRIMARY KEY,
    servicename VARCHAR(250) NOT NULL UNIQUE,
    duration_minutes INTEGER NOT NULL,
    fee DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  )`
	if _, err = Db.Exec(serviceAvailable); err != nil {
		log.Fatalf("Failed to create table serviceavailable :%v ", err)
	}
	log.Println("Table 'serviceavailable' created or already exists.")

	serviceAvailable2 := `
	CREATE TABLE IF NOT EXISTS serviceavailable_tb (
	  serv2_id SERIAL PRIMARY KEY,
	  servicename VARCHAR(250) NOT NULL UNIQUE,
	  initial_number INTEGER NOT NULL,
	  fee DECIMAL(10,2) NOT NULL,
	  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err = Db.Exec(serviceAvailable2); err != nil {
		log.Fatalf("Failed to create table serviceavailable_tb :%v ", err)
	}
	log.Println("Table 'serviceavailable_tb' created or already exists.")

	specialgroup := `
	CREATE TABLE IF NOT EXISTS specialgroup (
	  spec_id SERIAL PRIMARY KEY,
	  username VARCHAR(200),
	  secretkey VARCHAR(200),
	  age INTEGER,
	  managedby_id INTEGER REFERENCES users(user_id),
	  dialforcreator VARCHAR(20),
	  dialforuser VARCHAR(20),
	  reason TEXT NOT NULL,
	  FOREIGN KEY (dialforcreator) REFERENCES users(dial),
	  FOREIGN KEY (dialforuser) REFERENCES users(dial)
	)`
	if _, err = Db.Exec(specialgroup); err != nil {
		log.Fatal("Failed to create table specialgroup", err)
	}
	log.Println("Table 'specialgroup' created or already exists.")

	bookingtracking := `CREATE TABLE IF NOT EXISTS bookingtrack_tb (
	  id SERIAL PRIMARY KEY,
	  user_id INTEGER REFERENCES users(user_id),
	  spec_id INTEGER REFERENCES specialgroup(spec_id),
	  doctor_id INTEGER REFERENCES doctors(doctor_id),
	  service_id INTEGER REFERENCES serviceavailable(serv_id),
	  service2_id INTEGER REFERENCES serviceavailable_tb(serv2_id),
	  booking_date DATE NOT NULL,
	  dayofweek INTEGER NOT NULL,
	  start_time TIME NOT NULL,
	  end_time TIME NOT NULL,
	  status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed')),
	  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err = Db.Exec(bookingtracking); err != nil {
		log.Fatalf("Failed to create table bookingtrack_tb :%v", err)
	}
	log.Println("Table 'bookingtrack_tb' created or already exists.")
	scheduled_notificationstb := `CREATE TABLE IF NOT EXISTS scheduled_notifications (
		id SERIAL PRIMARY KEY,
		booking_id INTEGER REFERENCES bookingtrack_tb(id),
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	  );`
	if _, err = Db.Exec(scheduled_notificationstb); err != nil {
		log.Fatalf("failed to create table scheduled_notifications: %v", err)
	}
	log.Println("Table 'scheduled_notifications' created or already exists.")

	doctorServ_tb := `
	   CREATE TABLE IF NOT EXISTS doctor_services (
		id SERIAL PRIMARY KEY,
		doctor_id INTEGER REFERENCES doctors(doctor_id),
		service_id INTEGER REFERENCES serviceavailable(serv_id),
		service2_id INTEGER REFERENCES serviceavailable_tb(serv2_id),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	  )`
	if _, err = Db.Exec(doctorServ_tb); err != nil {
		log.Fatalf("Failed to create table doctor_services: %v", err)
	}
	log.Println("Table 'doctor_services' created or already exists.")

	return nil
}
