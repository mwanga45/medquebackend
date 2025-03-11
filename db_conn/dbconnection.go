package _handler_conn

import "database/sql"

var Db *sql.DB
// initilaze connection pool
func Connectionpool(databasesourceName string) error {
	var err error
	if Db,err = sql.Open("postres", databasesourceName);err !=nil{
		return err
	}
	// Db.Ping used to verify if the connection is alive and properly configured 
	if err = Db.Ping(); err != nil{
		return err
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