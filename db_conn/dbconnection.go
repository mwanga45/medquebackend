package handlerconn

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var Db *sql.DB
// initilaze connection pool
func Connectionpool(databasesourceName string) error {
	var err error
	if Db,err = sql.Open("postgres", databasesourceName);err !=nil{
		return err
	}
	// Db.Ping used to verify if the connection is alive and properly configured 
	if err = Db.Ping(); err != nil{
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
    consultation_fee DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`
	if _,err = Db.Exec(doctor_tb);err !=nil{
		log.Fatalf("failed to create new table %v", err)
	}

	return nil

}
// create fi=unction to terminate connection 
func Closeconn()error {
	if Db != nil{
		return Db.Close()
	}
	return nil

	
}