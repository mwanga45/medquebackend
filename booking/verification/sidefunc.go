package sidefunc_test

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	// "golang.org/x/crypto/bcrypt"
)

type (
	Timeslot struct {
		StartTime string
		EndTime   string
	}
)

func GenerateTimeSlote(timeInterval int, startTime string, endTime string) ([]Timeslot, error) {
	layout := "15:04"
	now := time.Now()

	start, err := time.ParseInLocation(layout, startTime, now.Location())
	if err != nil {

		return nil, fmt.Errorf("something went wrong: %w", err)
	}
	end, err := time.ParseInLocation(layout, endTime, now.Location())
	if err != nil {
		return nil, fmt.Errorf("something  went wrong: %w ", err)
	}
	if end.Before(start) {
		end = end.Add(24 * time.Hour)
	}
	var slot []Timeslot
	current := start

	for {
		next := current.Add(time.Duration(timeInterval) * time.Minute)
		if current.After(end) {
			break
		}
		slot = append(slot, Timeslot{
			StartTime: current.Format(layout),
			EndTime:   next.Format(layout),
		})
		current = next
	}
	return slot, nil

}

func Dayofweek(day int) (string, error) {
	daysname := []string{
		"Sunday",
		"Monday",
		"Tuesday",
		"Wednesday",
		"Thursday",
		"Friday",
		"Saturday",
	}
	if day < 0 || day > 6 {
		return "", fmt.Errorf("the day value should be   range from 0-6")
	}
	return daysname[day], nil
}
func DayOfWeekReverse(day string) (int, error) {
    switch day {
    case "Sunday":
        return 0, nil
    case "Monday":
        return 1, nil
    case "Tuesday":
        return 2, nil
    case "Wednesday":
        return 3, nil
    case "Thursday":
        return 4, nil
    case "Friday":
        return 5, nil
    case "Saturday":
        return 6, nil
    default:
        return -1, fmt.Errorf("invalid weekday: %q", day)
    }
}
func DaytimeofToday(dayoftoday string, dayname string) {

}


func CheckalreadybookedToday(userid string, date string, tx *sql.Tx) (bool, error) {
	const query = `
        SELECT user_id, start_time, end_time, dayofweek, spec_id 
        FROM bookingTrack_tb  
        WHERE booking_date = $1 AND user_id = $2
    `

	rows, err := tx.Query(query, date, userid)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, fmt.Errorf("database query failed: %w", err)
	}

	defer rows.Close()

	hasNormalBooking := false
	for rows.Next() {
		var (
			user_id    int
			start_time string
			end_time   string
			dayofweek  int
			spec_id    int
		)

		if err := rows.Scan(&user_id, &start_time, &end_time, &dayofweek, &spec_id); err != nil {
			return false, fmt.Errorf("row scan failed: %w", err)
		}

		if spec_id == 0 {
			hasNormalBooking = true
		}
	}

	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("rows iteration error: %w", err)
	}
	return hasNormalBooking, nil
}

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

func ValidateSecretkey(key string) error {
	if len(key) < 6 {
		return errors.New("please secretekey must have atleast 6 character")

	}
	var hasUppercase, hasNumber bool
	for _, r := range key {
		switch {
		case unicode.IsUpper(r):
			hasUppercase = true
		case unicode.IsDigit(r):
			hasNumber = true

		}   
	}
	if !hasUppercase {
		return errors.New("please make sure your Secretkey having atleast one Uppercase character")
	}
	if !hasNumber {
		return errors.New("please make sure your your Secretkey having atleast one digit")
	}
	return nil

}

func HandleREGprocess(table string,username string, regNo string,password []byte, phone_number string, email string, home_address string, tx *sql.Tx)error{
	var exist bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE regNo = $1)",table)
	errExist :=tx.QueryRow(query,regNo).Scan(&exist)
	if errExist != nil {
		return fmt.Errorf("something went wrong")
	}
	if exist{
		return fmt.Errorf("staff is already registered with this registration number")
	}
   _, errExec := tx.Exec(fmt.Sprintf("INSERT INTO %s (username,regNO,password,phone_number,email,home_address)  VALUES ($1, $2, $3, $4, $5,$6)",table),username,regNo,password,phone_number,email,home_address)
   if errExec !=nil {
	return fmt.Errorf("something went wwrong here")
   }
  return nil
}

func CheckIdentifiaction(RegNo string)(bool)  {
    if strings.HasPrefix(RegNo,"MHD/DKT"){
		return true
	}else{
		return false
	}
}
func Checkpassword(pwrd , cpwrd string)(bool){
   if pwrd == cpwrd {
	return true
   }else{
	return false
   }
}
