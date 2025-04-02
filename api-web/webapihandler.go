package apiweb
 type Home_address struct{
	City string `json:"city,omitempty"`
	PostAdddres string `json:"post_address,omitempty"`
 }
type Staffdetatails struct {
 Username  string `json:"username"`
 Password string `json:"password"`
 Phone string `json:"phone"`
 Email string `json:"email"`
 Home_address Home_address `json:"home_address"`
}