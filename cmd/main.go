package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jonathaninc/backend/internal/handlers"
)

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// requestLogger middleware adds request logging
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Request started: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		log.Printf("Request completed: %s %s in %v", r.Method, r.URL.Path, duration)
	})
}

func main() {
	log.Println("Starting application...")

	// Get Redis configuration from environment variables
	redisHost := os.Getenv("CONNECTION_REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
		log.Println("CONNECTION_REDIS_HOST not set, using default:", redisHost)
	} else {
		log.Println("Redis host:", redisHost)
	}

	redisPort := os.Getenv("CONNECTION_REDIS_PORT")
	if redisPort == "" {
		redisPort = "6380"
		log.Println("CONNECTION_REDIS_PORT not set, using default:", redisPort)
	} else {
		log.Println("Redis port:", redisPort)
	}

	redisAddr := redisHost + ":" + redisPort
	log.Printf("Connecting to Redis at %s", redisAddr)

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Successfully connected to Redis")

	// Create counterService with Redis client
	counterService := handlers.NewCounterService(redisClient)

	// Create a new mux for routing
	mux := http.NewServeMux()

	// Hello endpoint - shows current count
	mux.HandleFunc("/hello", handlers.HelloHandler(counterService))

	// Increment endpoint
	mux.HandleFunc("/increment", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Printf("Method not allowed: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Printf("Processing increment request from %s", r.RemoteAddr)
		newCount, err := counterService.Increment(r.Context())
		if err != nil {
			log.Printf("Error incrementing counter: %v", err)
			http.Error(w, "Error incrementing counter", http.StatusInternalServerError)
			return
		}

		log.Printf("Counter incremented to %d", newCount)
		w.Write([]byte("Counter incremented. New value: " + strconv.Itoa(newCount)))
	})

	// Decrement endpoint
	mux.HandleFunc("/decrement", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Printf("Method not allowed: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Printf("Processing decrement request from %s", r.RemoteAddr)
		newCount, err := counterService.Decrement(r.Context())
		if err != nil {
			log.Printf("Error decrementing counter: %v", err)
			http.Error(w, "Error decrementing counter", http.StatusInternalServerError)
			return
		}

		log.Printf("Counter decremented to %d", newCount)
		w.Write([]byte("Counter decremented. New value: " + strconv.Itoa(newCount)))
	})

	// Status endpoint - returns current count
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Processing status request from %s", r.RemoteAddr)
		currentCount, err := counterService.GetCurrentCount(r.Context())
		if err != nil {
			log.Printf("Error retrieving counter: %v", err)
			http.Error(w, "Error retrieving counter", http.StatusInternalServerError)
			return
		}

		log.Printf("Returning current count: %d", currentCount)
		w.Write([]byte("Current counter value: " + strconv.Itoa(currentCount)))
	})

	// Add a health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Check Redis connection
		if _, err := redisClient.Ping(ctx).Result(); err != nil {
			log.Printf("Health check failed: Redis connection error: %v", err)
			http.Error(w, "Redis connection error", http.StatusServiceUnavailable)
			return
		}

		log.Printf("Health check passed")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply middleware chain: CORS + Request logging
	handler := corsMiddleware(requestLogger(mux))

	serverPort := ":8080"
	log.Printf("Starting server on %s", serverPort)
	if err := http.ListenAndServe(serverPort, handler); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
