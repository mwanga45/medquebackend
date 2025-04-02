package apiweb

import (
	// "fmt"
	"encoding/json"
	"net/http"
)

type Return_Field struct{
	Message  string  `json:"message"`
	Data interface {} `json:"data"`
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

func Login(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost{
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Return_Field{
			Message: "Invalid method is use to access resource",
		})
		return
	}
}