package main

import (
	"flag"
	"net/http"

	"dx2.dev/werd/internal/auth"
	words "dx2.dev/werd/internal/words"
	"github.com/sirupsen/logrus"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("Debug mode enabled")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/words", words.RandomWordsHandler)
	mux.HandleFunc("/auth/login", auth.LoginHandler)
	mux.HandleFunc("/auth/createUser", auth.CreateUserHandler)

	logrus.Info("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
