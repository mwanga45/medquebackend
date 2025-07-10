package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	handlerconn "medquemod/db_conn"
	"net/http"
	"strings"
	"time"
)

type ExpoMessage struct {
	To    string `json:"to"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Sound string `json:"sound"`
}

func SendNotification(message ExpoMessage) error {
	if !isValidExpoToken(message.To) {
		return fmt.Errorf("invalid Expo token format")
	}
	payload, err := json.Marshal([]ExpoMessage{message})
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	resp, err := http.Post(
		"https://exp.host/--/api/v2/push/send",
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expo API error %d: %s", resp.StatusCode, string(body))
	}
	var result struct {
		Data []struct {
			Status  string `json:"status"`
			ID      string `json:"id"`
			Message string `json:"message"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("invalid response format: %w", err)
	}
	if len(result.Data) == 0 || result.Data[0].Status != "ok" {
		return fmt.Errorf("expo rejection: %s", result.Data[0].Message)
	}
	return nil
}

func updateNotificationStatus(id int, status string) {
	validStatuses := map[string]bool{
		"pending": true,
		"sent":    true,
		"failed":  true,
		"expired": true,
	}
	if !validStatuses[status] {
		log.Printf("Invalid status: %s", status)
		return
	}
	_, err := handlerconn.Db.Exec(
		`UPDATE scheduled_notifications 
         SET status = $1, updated_at = NOW() AT TIME ZONE 'UTC'
         WHERE id = $2`,
		status, id,
	)
	if err != nil {
		log.Printf("Update error: %v", err)
	}
}

func isValidExpoToken(token string) bool {
	return strings.HasPrefix(token, "ExponentPushToken[") &&
		strings.HasSuffix(token, "]") &&
		len(token) > 25
}

func checkPendingNotifications() {
	query := `SELECT sn.id, u.fullname, u.deviceid, b.booking_date, b.start_time, b.id as booking_id
		FROM scheduled_notifications sn
		JOIN bookingtrack_tb b ON sn.booking_id = b.id
		JOIN users u ON b.user_id = u.user_id
		WHERE (sn.status IS NULL OR sn.status = 'pending')
		AND (b.booking_date || ' ' || b.start_time)::timestamp > NOW()`

	rows, err := handlerconn.Db.Query(query)
	if err != nil {
		log.Printf("Query error: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id          int
			username    string
			deviceID    string
			bookingDate time.Time
			startTime   string
			bookingID   int
		)

		if err := rows.Scan(&id, &username, &deviceID, &bookingDate, &startTime, &bookingID); err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}

		// Parse the start time
		startTimeParsed, err := time.Parse("15:04:05", startTime)
		if err != nil {
			log.Printf("Time parse error: %v", err)
			continue
		}

		bookingDateTime := time.Date(
			bookingDate.Year(), bookingDate.Month(), bookingDate.Day(),
			startTimeParsed.Hour(), startTimeParsed.Minute(), startTimeParsed.Second(),
			0, time.UTC,
		)

		notificationTime := bookingDateTime.Add(-5 * time.Minute)
		currentTime := time.Now()

		if currentTime.After(notificationTime) && currentTime.Before(notificationTime.Add(1*time.Minute)) {
			msg := ExpoMessage{
				To:    deviceID,
				Title: "Booking Reminder",
				Body:  fmt.Sprintf("Hi %s, your booking is in 5 minutes at %s!", username, startTime),
				Sound: "default",
			}

			if err := SendNotification(msg); err != nil {
				log.Printf("Failed to send notification: %v", err)
				updateNotificationStatus(id, "failed")
				continue
			}

			updateNotificationStatus(id, "sent")
			log.Printf("Notification sent for booking ID %d", bookingID)
		} else if currentTime.After(bookingDateTime) {
			// Booking time has passed, mark as expired
			updateNotificationStatus(id, "expired")
			log.Printf("Notification expired for booking ID %d", bookingID)
		}
		// If notification time hasn't come yet, leave as pending
	}

	if err = rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
	}
}

func StartNotificationWorker(ctx context.Context) {
	log.Println("Starting notification worker - checking every minute for pending notifications")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Check immediately on startup
	checkPendingNotifications()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping notification worker")
			return
		case <-ticker.C:
			checkPendingNotifications()
		}
	}
}

// GetNotificationStatus - for debugging and monitoring
func GetNotificationStatus() (map[string]int, error) {
	query := `SELECT status, COUNT(*) as count 
		FROM scheduled_notifications 
		GROUP BY status`

	rows, err := handlerconn.Db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	statusCounts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		statusCounts[status] = count
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return statusCounts, nil
}

// ManualTriggerNotifications - for testing purposes
func ManualTriggerNotifications() {
	log.Println("Manually triggering notification check...")
	checkPendingNotifications()
}

// RegisterPushToken handles registration of Expo push tokens from the frontend
func RegisterPushToken(w http.ResponseWriter, r *http.Request) {
	type request struct {
		DeviceID string `json:"deviceId"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}
	if req.DeviceID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing deviceId"})
		return
	}
	// TODO: Associate deviceId with the authenticated user in the database
	log.Printf("Received push token: %s", req.DeviceID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
