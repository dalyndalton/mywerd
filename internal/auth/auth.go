package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"log"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func initDB() {
	var err error
	db_url := "postgres://postgres:password@localhost:5432/werd?sslmode=disable"
	if envPath := os.Getenv("DB_URL"); envPath != "" {
		db_url = envPath
	}

	db, err = sql.Open("postgres", db_url)
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(1 * time.Minute)

	// Test connection
	err = db.Ping()
	if err != nil {
		logrus.Fatal("Failed to connect to the database:", err)
	}
	logrus.Info("Connected to the database :rocket:")
}

func WithAuthUser(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Use the session token from the cookie
		sessionToken := cookie.Value

		// Get the session from the database
		var userID string
		var expiresAt time.Time
		err = db.QueryRow("SELECT userID, expires_at FROM sessions WHERE session_token = $1", sessionToken).Scan(&userID, &expiresAt)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if the session has expired
		if time.Now().After(expiresAt) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add the user ID to the request context
		ctx := context.WithValue(r.Context(), "userID", userID)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	logrus.WithField("username", username).Info("Creating user")

	var existingUsername string
	err := db.QueryRow("SELECT username FROM users WHERE username = $1", username).Scan(&existingUsername)
	if err == nil {
		http.Error(w, "Username already taken", http.StatusBadRequest)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logrus.Error("Failed to hash the password:", err)
		return
	}

	// Insert the user into the database
	_, err = db.Exec("INSERT INTO users(username, password_hash) VALUES($1, $2)", username, hashedPassword)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logrus.Error("Failed to insert the user into the database:", err)
		return
	}

	// Redirect to the home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Get the hashed password from the database
	var userID, hashedPassword string
	err := db.QueryRow("SELECT user_id, password_hash FROM users WHERE username = $1", username).Scan(&userID, &hashedPassword)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		logrus.Error("Failed to get the hashed password from the database:", err)
		return
	}

	// Compare the hashed password with the password from the form
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Generate a new session token
	sessionToken, _ := generateSessionToken()
	expiration := time.Now().Add(24 * time.Hour)

	// If the user already has a session, delete it
	_, err = db.Exec("DELETE FROM sessions WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logrus.Error("Failed to delete the session from the database:", err)
		return
	}

	// Insert the session into the database
	_, err = db.Exec("INSERT INTO sessions(user_id, session_token, expires_at) VALUES($1, $2, $3)", userID, sessionToken, expiration)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logrus.Error("Failed to insert the session into the database:", err)
		return
	}

	// Set the session token in a cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiration,
	})

	// Redirect to the home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	sessionToken := cookie.Value

	_, err = db.Exec("DELETE FROM sessions WHERE session_token = $1", sessionToken)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Delete the session token cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		MaxAge: -1,
	})

	// Redirect to the home page
	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}

func generateSessionToken() (string, error) {
	bytes := make([]byte, 32) // 256-bit token
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func init() {
	initDB()

}
