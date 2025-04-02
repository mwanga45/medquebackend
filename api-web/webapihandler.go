package apiweb

import (
	// "fmt"
	"encoding/json"
	"fmt"
	handlerconn "medquemod/db_conn"
	"net/http"
	"time"
	"github.com/golang-jwt/jwt"

)

type Return_Field struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
type Home_address struct {
	City        string `json:"city,omitempty"`
	PostAdddres string `json:"post_address,omitempty"`
}
type Staffdetatails struct {
	Username     string       `json:"username"`
	Password     string       `json:"password"`
	Phone        string       `json:"phone"`
	Email        string       `json:"email"`
	Home_address Home_address `json:"home_address"`
}
    var secretekey = []byte("secretekey")
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Return_Field{
			Message: "Invalid method is use to access resource",
		})
		return
	}

	var S Staffdetatails
	if err := json.NewDecoder(r.Body).Decode(&S); err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Return_Field{
			Message: "Something went wrong in  system",
		})
		fmt.Println("Something went wrong here",err)
		return 
	}

  query := "SELECT password , username Admin WHERE  password = $1 , username = $2"
  if  err := handlerconn.Db.QueryRow(query,&S.Password,&S.Username).Scan(&S.Password,&S.Username); err != nil{
      fmt.Println("User is username or password is wrong", err)
	  return
   }else{
	fmt.Println("Something went wrong here", err)
   }
   w.Header().Set("Content-Type", "application/json")
   if err := json.NewEncoder(w).Encode(Return_Field{
	Data: S,
	Message: "Success full Login",
   }); err != nil{
     fmt.Println("Something went wrong", err)
   }
   
}

func CreateToken(Username string,Password string)(string, error){
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,jwt.MapClaims{
		"Username": Username,
		"Password":Password,
		"exp": time.Now().Add(time.Hour * 2).Unix(),
	})
	tokenstring,err := token.SignedString(secretekey)
	if err != nil{
		fmt.Println("Something went wrong",err)
		return "",err
	}
	return tokenstring,nil
}
func verifyToken(tokenstring string)(error){
	token, err := jwt.Parse(tokenstring, func(t *jwt.Token)(interface{},error){
	   return secretekey,nil
	})
	if err != nil{
		fmt.Println("Something went wrong",err)
		return err
	}
	if !token.Valid{
		fmt.Println("Token isnt exist")
	}
	return nil
}