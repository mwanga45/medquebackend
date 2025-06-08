package sidefunc_test

import (
	"fmt"
	"time"
)

type Timeslot struct{
	StartTime  string
	EndTime string
}
func GenerateTimeSlote(timeInterval int, startTime string, endTime string)([]Timeslot, error) {
	layout := "15:04"
	now := time.Now()

	start, err :=  time.ParseInLocation(layout,startTime,now.Location())
	if err != nil{
	
		return nil , fmt.Errorf("Something went wrong", err)
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