package main

import (
	"context"
	"fmt"
	"log"

	// "medquemod/api"
	// apiweb "medquemod/api-web"
	"medquemod/booking"
	Routesf_test "medquemod/routes"

	// "medquemod/bookingtimelogic"
	// handler_chat "medquemod/chatbot"
	handlerconn "medquemod/db_conn"
	// "medquemod/handleauthentic"
	"net/http"
	"os"
	"os/signal"

	// "os/user"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	
	envPath := ".env"
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Fatal: .env file not found at %s or failed to load: %v", envPath, err)
	}

	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	
	if err := handlerconn.Connectionpool(); err != nil {
		log.Fatalf("something went wrong failed to connect to database %v", err)
	}

	go booking.StartNotificationWorker(ctx)

	// Set up router and routes
	r := mux.NewRouter()
	Routesf_test.HandleRoutes(r)

	// Configure CORS
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"POST", "PUT", "GET", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// Create HTTP server with proper configuration
	server := &http.Server{
		Addr:    ":8801",
		Handler: cors(r),
	}

	
	go func() {
		fmt.Println("Server listening on port 8801...")
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
