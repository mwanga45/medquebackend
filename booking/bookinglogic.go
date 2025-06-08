package booking


type (
	Bkrequest struct{
      ServId  string `json:"servid" validate:"required"`
	  IntervalTime  string `json:"timeInter" validate:"required"`
	  Servicename string `json:"servicename" validate:"required"`
	  Token string   `json:"token" validate:"required"`  
	}
)
func bookinglogic()  {
	
}





