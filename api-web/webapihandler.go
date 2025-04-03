// package apiweb

// import (
// 	// "fmt"
// 	"encoding/json"
// 	"fmt"
// 	"github.com/golang-jwt/jwt"
// 	handlerconn "medquemod/db_conn"
// 	"net/http"
// 	"time"
// )

// type Return_Field struct {
// 	Message string      `json:"message"`
// 	Data    interface{} `json:"data"`
// }
// type Home_address struct {
// 	City        string `json:"city,omitempty"`
// 	PostAdddres string `json:"post_address,omitempty"`
// }
// type Staffdetatails struct {
// 	Username     string       `json:"username"`
// 	Password     string       `json:"password"`
// 	Phone        string       `json:"phone"`
// 	Email        string       `json:"email"`
// 	Home_address Home_address `json:"home_address"`
// }
// type StaffLogin struct{
// 	Username  string `json:"username"`
// 	Password  string `json:"password"`
// 	Registration string `json:"Registration"`
// }

// var secretekey = []byte("secretekey")

// func Registration(w http.ResponseWriter, r * http.Request){

// }
// func Login(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		w.WriteHeader(http.StatusMethodNotAllowed)
// 		json.NewEncoder(w).Encode(Return_Field{
// 			Message: "Invalid method is use to access resource",
// 		})
// 		return
// 	}
// 	var S StaffLogin
// 	if err := json.NewDecoder(r.Body).Decode(&S); err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(Return_Field{
// 			Message: "Something went wrong in  system",
// 		})
// 		fmt.Println("Something went wrong here", err)
// 		return
// 	}

// 	query := "SELECT password,username,regNo AdminTb WHERE  password = $1,username = $2 regNo =  $3"
// 	if err := handlerconn.Db.QueryRow(query, &S.Password, &S.Username, &S.Registration).Scan(&S.Password, &S.Username,&S.Registration); err != nil {
// 		fmt.Println("User is username or password is wrong", err)
// 		return
// 	} else {
// 		token, err := CreateToken(S.Username, S.Password)
// 		if err != nil {
// 			fmt.Println("Something went wrong", err)
// 		}
// 		fmt.Println(token)
// 		w.WriteHeader(http.StatusOK)
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	if err := json.NewEncoder(w).Encode(Return_Field{
// 		Data:    S,
// 		Message: "Success full Login",
// 	}); err != nil {
// 		fmt.Println("Something went wrong", err)
// 	}

// }

//	func CreateToken(Username string, Password string) (string, error) {
//		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
//			"Username": Username,
//			"Password": Password,
//			"exp":      time.Now().Add(time.Hour * 2).Unix(),
//		})
//		tokenstring, err := token.SignedString(secretekey)
//		if err != nil {
//			fmt.Println("Something went wrong", err)
//			return "", err
//		}
//		return tokenstring, nil
//	}
//
//	func VerifyToken(tokenstring string) error {
//		token, err := jwt.Parse(tokenstring, func(t *jwt.Token) (interface{}, error) {
//			return secretekey, nil
//		})
//		if err != nil {
//			fmt.Println("Something went wrong", err)
//			return err
//		}
//		if !token.Valid {
//			fmt.Println("Token isnt exist")
//		}
//		return nil
//	}
package apiweb

import (
	"encoding/json"
	"fmt"
	handlerconn "medquemod/db_conn"
	"net/http"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// create structure for the login
type (
	StaffLogin struct {
		Username     string `json:"username" validate:"required"`
		Password     string `json:"password" validate:"required"`
		Registration string `json:"registarion" validate:"required"`
	}
	// create structure for the Token
	Claims struct {
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
	}
	hashedPassword , err := Check_RegNo(SL.Username,SL.Registration)
	if err !=nil{
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Respond{
			Success: false,
			Message: "Failed to Sign-In Incorrect password or Username or Registration Number",
		})
	}
	 if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword),[]byte (SL.Password));err != nil{
        w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Respond{
			Success: false,
			Message: "Incorrect username or Password",
		})
	 }
}
func Check_RegNo(Registration string, Username string) (string, error) {
	// create variable that will hold the return hashpassword
	var Hash_password string
	// check for lenght of the Registration
	if len(Registration) < 6 {
		return "", fmt.Errorf("registration number too short")
	}
	if len(Registration) >= 6 {
		// select six first character from here
		check_staff := Registration[:6]
		if check_staff == "MHD/AD" {
			query := "SELECT password from Admin_tb WHERE username = 1$ AND regNo = 2$"
			if err := handlerconn.Db.QueryRow(query, Registration, Username).Scan(&Hash_password); err != nil {
				fmt.Print("User Doesnt not exist", err)
				return "", err
			} else {
				fmt.Println("Something went wrong", err)
				return "", err
			}
		} else if check_staff == "MDH/DKT" {
			query := "SELECT password from Doc_tb WHERE username = 1$ AND regNo = 2$ "
			if err := handlerconn.Db.QueryRow(query, Username, Registration).Scan(Hash_password); err != nil {
				fmt.Print("User Doesnt not exist", err)
				return "", err
			} else {
				fmt.Println("Something went wrong", err)
				return "", err
			}
		} else if check_staff == "MHD/NRS" {
			query := "SELECT password from Nurse_tb WHERE username = 1$ AND regNo = 2$"
			if err := handlerconn.Db.QueryRow(query, Registration, Username).Scan(&Hash_password); err != nil {
				fmt.Print("User Doesnt not exist", err)
				return "", err
			} else {
				return "",fmt.Errorf("invalid prefix used")
			}
		}
		if len(Registration) < 6 {
			return "", fmt.Errorf("registration number too short")
		}
	}
	return Hash_password,nil
}
