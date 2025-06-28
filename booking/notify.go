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
	query := `SELECT sn.id, u.fullname, u.deviceid, sn.notification_time
		FROM scheduled_notifications sn
		JOIN bookingtrack_tb b ON sn.booking_id = b.id
		JOIN users u ON b.user_id = u.user_id
		WHERE (sn.status IS NULL OR sn.status = 'pending')
		AND sn.notification_time > NOW()`
	rows, err := handlerconn.Db.Query(query)
	if err != nil {
		log.Printf("Query error: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id               int
			username         string
			deviceID         string
			notificationTime time.Time
		)
		if err := rows.Scan(&id, &username, &deviceID, &notificationTime); err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}
		msg := ExpoMessage{
			To:    deviceID,
			Title: "Booking Reminder",
			Body:  fmt.Sprintf("Hi %s, you have a booking coming up soon!", username),
			Sound: "default",
		}
		if notificationTime.After(time.Now()) {
			if err := SendNotification(msg); err != nil {
				log.Printf("Failed to send notification: %v", err)
				updateNotificationStatus(id, "failed")
				continue
			}
			updateNotificationStatus(id, "sent")
		} else {
			updateNotificationStatus(id, "expired")
		}
	}
	if err = rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
	}
}

func StartNotificationWorker(ctx context.Context) {
	log.Println("Starting notification worker")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
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
