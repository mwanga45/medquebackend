package apiweb

import (
	// "fmt"
	"encoding/json"
	"fmt"
	handlerconn "medquemod/db_conn"
	"net/http"
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
  const secretekey = []btyes("secretekey")
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
  rows, err := handlerconn.Db.QueryRow(query,).Scan()
}
func CreateToken(Username string , Passoword string) (string error){
	token := jwt.NewWithClaims(jwt.SigningClamsHs256,jwt.NewClaims{
		"Username": Username,
		"Password":Password,
		"exp": time.Now().Add(time.Hour * 2).Unix(),
	})
	tokenstring , err := token.Signedstring(secretekey)
}