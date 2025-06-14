package sidefunc_test

import (
	"database/sql"
	"fmt"
	"time"
)

type (
	Timeslot struct{
	StartTime  string
	EndTime string
}
    Username struct{
		Usrname string 
	}
   

)
func GenerateTimeSlote(timeInterval int, startTime string, endTime string)([]Timeslot, error) {
	layout := "15:04"
	now := time.Now()

	start, err :=  time.ParseInLocation(layout,startTime,now.Location())
	if err != nil{
	
		return nil , fmt.Errorf("Something went wrong ", err)
	}
   end, err :=  time.ParseInLocation(layout, endTime, now.Location())
   if err != nil{
	return nil , fmt.Errorf("Something  went wrong ", err)
   }
   if end.Before(start){
	  end = end.Add(24 *time.Hour)
   }
   var slot []Timeslot
   current  :=  start

   for {
	next :=  current.Add(time.Duration(timeInterval) *time.Minute)
	if current.After(end){
		break
	}
	slot = append(slot, Timeslot{
		StartTime: current.Format(layout),
		EndTime: next.Format(layout),
	})
	current = next
   }
   return  slot , nil

}
func Dayofweek(day int)(string, error){
	daysname  :=  []string{
		"Sunday",
		"Monday",
		"Tuesday",
		"Wensday",
		"Thursday",
		"Friday",
		"Saturday",
	}
	if day < 0 ||day > 6{
		return "", fmt.Errorf("the day value should be   range from 0-6")
	}
	return daysname[day], nil
}
func DaytimeofToday(dayoftoday string,dayname string ){

}

// func CheckBookingforWhom(isforMe bool , tx *sql.DB )  {
    
// }
// func checkalreadybookedToday(userid string,start_time string, end_time string )  {
	
// }

func Checklimit (userid string , tx *sql.DB)(string){
	checkrow , err := tx.Query(`SELECT username FROM Specialgroup WHERE managedby_id = $1 `, userid)
	if err != nil{
		return fmt.Sprint("Something went wrong",err)
	}
	specgroup :=  map[int]string{}
	var usr Username
	defer checkrow.Close()
    checkrow.Scan(usr.Usrname)
	var count int
	for checkrow.Next(){
      count ++ 
	  if count > 5{
		return "Sorry you have Already reach limit"
	  }
	}
	
	return ""
}