package apiweb

import (
	"encoding/json"
	"fmt"
	handlerconn "medquemod/db_conn"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// create structure for the login
type (
	StaffRegister struct{
		Username string `json:"username" validate:"required"`
		RegNo string `json:"regNo" validate:"required"`
		Password string `json:"password" validate:"required"`
		Phone string `json:"phone" validate:"required"`
		Email string `json:"email" validate:"required"`
		Home_address string `json:"home_address" validate:"required"`
	}
	StaffLogin struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
		RegNo    string `json:"registration" validate:"required"`
	}
	// create structure for the Token
	JwtClaims struct {
		Username string
		jwt.StandardClaims
	}
	// create  struct to return response
	Respond struct {
		Message string `json:"message"`
		Success bool   `json:"success"`
		Data    interface{}
	}
)
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Respond{
			Success: false,
			Message: "Invalid Method",
		})
		return
	}
	// create an instance for the  StaffLogin struct
	var SL StaffLogin
	if err := json.NewDecoder(r.Body).Decode(&SL); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Respond{
			Success: false,
			Message: "Bad Request",
		})
		return
	}
	hashedPassword, err := Check_RegNo(SL.Username, SL.RegNo)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Respond{
			Success: false,
			Message: "Failed to Sign-In Incorrect password or Username or Registration Number",
		})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(SL.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Respond{
			Success: false,
			Message: "Incorrect username or Password",
		})
		return
	}
	tokenstring, err := CreateToken(SL.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Respond{
			Success: false,
			Message: "Something went wrong",
		})
		return
	}
	json.NewEncoder(w).Encode(Respond{
		Success: true,
		Message: "Login Successfuly",
		Data:    map[string]string{"token": tokenstring},
	})
}

// create function that will create an token for session
var secretekey = []byte("secrete-key")

func CreateToken(username string) (string, error) {
	Claims := JwtClaims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 2).Unix(),
			Issuer:    "medque",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims)
	return token.SignedString(secretekey)

}
func VerifyToken(tokenstring string) error {
	tokenvalid, err := jwt.Parse(tokenstring, func(t *jwt.Token) (interface{}, error) {
		return secretekey, nil
	})
	if err != nil {
		fmt.Println("Something went wrong")
		return err
	}
	if !tokenvalid.Valid {
		return fmt.Errorf("page is alreadty expire")
	}
	return nil

}

func Check_RegNo(username string, regNo string)(string,error)  {
	// create variable that will hold the hashpassword
	var hashedPassword string

	if len(regNo) < 7{
		return "", fmt.Errorf("registratio number is too short")	
	}
//   select  first seven  character from the  regno
	check_regno := regNo[:7]

	switch check_regno{
	case"MHD/DKT":
		query := "SELECT password from Dkt_tb WHERE username = $1 AND regNo =$2"
		err := handlerconn.Db.QueryRow(query,username,regNo).Scan(&hashedPassword)
		if err != nil{
			fmt.Println("Something went wrong here, or  User don`t exist",err)
			return "",err
		}
		return hashedPassword,nil
	case "MHD/NRS":
		query := "SELECT password from Nrs_tb WHERE username = $1 AND regNo = $2"
		err := handlerconn.Db.QueryRow(query,username,regNo).Scan(&hashedPassword)
		if err !=nil{
			fmt.Println("Something went wrong here, or  User don`t exist",err)
			return "",err
		}
		return hashedPassword,nil
	case "MHD/ADM":
		query := "SELECT password from Admin_tb WHERE username = $1 AND regNo = $2"
		err := handlerconn.Db.QueryRow(query,username,regNo).Scan(&hashedPassword)
		if err != nil{
          fmt.Println("Something went wrong here, or  User don`t exist",err)
		  return "",err
		}
		return hashedPassword, nil

	default:
	return "",fmt.Errorf("invalid registration number ")
	}
}
func HandleRegisterUser(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost{
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Respond{
			Success: false,
			Message: "Invalid Method",
		})
		return
	}
	var SR StaffRegister
	err := json.NewDecoder(r.Body).Decode(&SR)
	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Respond{
			Success: false,
			Message: "bad request",
		})
		return
	}
	notExist := Staffexist(SR.RegNo)
	if notExist != nil{
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Respond{
			Success: false,
			Message: "Staff is not exist yet",
		})
		return
	}
	if errAssign := Check_Identification(SR.Username,SR.RegNo,SR.Password,SR.Phone,SR.Email,SR.Home_address);errAssign !=nil{
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Respond{
			Success: false,
			Message: "Something went wrong or User isn`t exist yet in system",
		})
		return
	}
	json.NewEncoder(w).Encode(Respond{
		Success: true,
		Message: "Successfully registered",
	})
	
}
// check if the staff registration Number is exist already in system
func Staffexist(regNo string)(error){
	query := "SELECT regNo from Staff_tb WHERE regNo = $1"
	var RegNo string
	err := handlerconn.Db.QueryRow(query,regNo).Scan(&RegNo)
	if err != nil{
		return  fmt.Errorf("something went wrong Or user does`nt exist yet in system: %v",err)
	}
	return nil
}
// check there  identification  to assign them to appropiate table

func Check_Identification(username string, regNo string,password string,phone_number string, email string,home_address string)(error){
	check_reg := regNo[:7]

	switch check_reg{
	case"MHD/DKT":
		query := "INSERT INTO Dkt_tb username,regNO,password,phone_number,email,home_address  VALUES username = $1, regNo = $2, password = $3, phone_number =$4, email = $5,home_address = $6"
		_,err:= handlerconn.Db.Exec(query, username,regNo,password,phone_number,email,home_address)
		if err != nil{
			return fmt.Errorf("something went wrong: %v",err)
		}
		return nil
	case "MHD/ADM":
		query := "INSERT INTO Admin_tb username,regNO,password,phone_number,email,home_address  VALUES username = $1, regNo = $2, password = $3, phone_number =$4, email = $5,home_address = $6"
		_,err:= handlerconn.Db.Exec(query, username,regNo,password,phone_number,email,home_address)
		if err != nil{
			return fmt.Errorf("something went wrong: %v",err)
		}
		return nil
	case "MHD/NRS":	
	query := "INSERT INTO Nrs_tb username,regNO,password,phone_number,email,home_address  VALUES username = $1, regNo = $2, password = $3, phone_number =$4, email = $5,home_address = $6"
		_,err:= handlerconn.Db.Exec(query, username,regNo,password,phone_number,email,home_address)
		if err != nil{
			return fmt.Errorf("something went wrong: %v",err)
		}
		return nil
	}
	return fmt.Errorf("something went wrong failed to proccess data")
   
}

