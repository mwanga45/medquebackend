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
   

)
func GenerateTimeSlote(timeInterval int, startTime string, endTime string)([]Timeslot, error) {
	layout := "15:04"
	now := time.Now()

	start, err :=  time.ParseInLocation(layout,startTime,now.Location())
	if err != nil{
	
		return nil , fmt.Errorf("something went wrong: %w", err)
	}
   end, err :=  time.ParseInLocation(layout, endTime, now.Location())
   if err != nil{
	return nil , fmt.Errorf("something  went wrong: %w ", err)
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
func CheckLimit(userID string, client *sql.Tx) (map[int]string, error) {
	rows, err := client.Query(
		`SELECT fullname, spec_id FROM Specialgroup WHERE managedby_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	
	specials := make(map[int]string)

	
	var count int
	for rows.Next() {
		var name string
		var specID int
		if err := rows.Scan(&name, &specID); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		specials[specID] = name
		count++

		if count > 5 {
			return specials, fmt.Errorf("limit exceeded: you have already reached the limit of 5 people")
		}
	}
	if err := rows.Err(); err != nil {
		return specials, fmt.Errorf("iteration error: %w", err)
	}

	return specials, nil
}
