package main

import (
	"context"
	"fmt"
	"log"
	"medquemod/api"
	apiweb "medquemod/api-web"
	"medquemod/booking"
	"medquemod/bookingtimelogic"
	handler_chat "medquemod/chatbot"
	"medquemod/db_conn"
	"medquemod/handleauthentic"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	// Create context and signal channel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize database connection first
	const conn_string = "user=postgres dbname=medque password=lynx host=localhost sslmode=disable"
	if err := handlerconn.Connectionpool(conn_string); err != nil {
				log.Fatalf("something went wrong failed to connect to database %v", err)
			}

	// Start notification worker
	go booking.StartNotificationWorker(ctx)

	// Set up router and routes
	r := mux.NewRouter()
	r.HandleFunc("/register", authentic.Handler).Methods("POST")
		r.HandleFunc("/doctorinfo", api.Doctors).Methods("GET")
		r.HandleFunc("/userinfo", api.Userdetails).Methods("POST")
		r.HandleFunc("/chatbot",handler_chat.Chatbot).Methods("POST")
		r.HandleFunc("/verifyuser",api.Verifyuser).Methods("POST")
		r.HandleFunc("/registerstaff",apiweb.HandleRegisterUser).Methods("POST") //for webapplication
		r.HandleFunc("/staffsignIn",apiweb.LoginHandler).Methods("POST")//for webapplication
		r.HandleFunc("/bookinglogic", bookingtimelogic.Timelogic).Methods("GET")
		r.HandleFunc("/serviceAvailable", api.GetService).Methods("GET")

	// Configure CORS
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"POST", "PUT", "GET", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// Create HTTP server with proper configuration
	server := &http.Server{
		Addr:    ":8800",
		Handler: cors(r),
	}

	// Start server in a goroutine
	go func() {
		fmt.Println("Server listening on port 8800...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}