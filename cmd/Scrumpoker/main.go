package main

import (
	"Scrumpoker/handler"
	"Scrumpoker/models"
	"context"

	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"time"

	"github.com/gorilla/mux"
)

type authenticationMiddleware struct {
	manager           *models.SessionManager
	ScrumpokerSession *handler.ScrumpokerSession
}

func (amw *authenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuid := ""
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			c, err := r.Cookie(handler.Uuid_token)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			uuid = c.Value
		} else {
			uuid = authHeader[len("Bearer "):]
		}

		session, err := amw.manager.GetSession(mux.Vars(r)["id"])
		if err != nil {
			handler.RedirectToMainandDisplayError(err.Error(), &w, r)
			return
		}

		user, err := session.GetUser(uuid)
		if err != nil {
			handler.RedirectToMainandDisplayError(err.Error(), &w, r)
			return
		}

		amw.ScrumpokerSession.Session = session
		amw.ScrumpokerSession.User = user
		amw.ScrumpokerSession.Uuid = uuid
		next.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Println("Scrumpoker REST API Testserver Init")

	var dir string
	var wait time.Duration
	var cleanupInterval time.Duration
	var maxSessionLifetime time.Duration

	flag.StringVar(&dir, "dir", "static", "the directory to serve files from. Defaults to the current dir")
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.DurationVar(&cleanupInterval, "cleanup-timeout", time.Minute*15, "the interval in which we chek that a Session is active (if not remove it from the manager)")
	flag.DurationVar(&maxSessionLifetime, "max-session-lifetime", time.Minute*60, "How lomg a Session can stay inactive")
	flag.Parse()

	sessionManager := models.InitilaizeManager(maxSessionLifetime)
	spSession := handler.ScrumpokerSession{}
	amw := authenticationMiddleware{
		manager:           sessionManager,
		ScrumpokerSession: &spSession,
	}

	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))

	// HTML-Pages Serving Endpoints
	r.HandleFunc("/", handler.HomePage).Methods("GET", "INDEX", "VIEW")
	r.HandleFunc("/join", handler.JoinSession(sessionManager)).Methods("POST")
	r.HandleFunc("/create", handler.CreateSession(sessionManager)).Methods("POST")

	//Rest-Api Endpoints Middleware
	scrumpokerRouter := r.PathPrefix("/session").Subrouter()
	scrumpokerRouter.Use(amw.Middleware)

	//Rest-Api Endpoints
	scrumpokerRouter.HandleFunc("/{id}", spSession.ViewSession(sessionManager)).Methods("GET", "VIEW")
	scrumpokerRouter.HandleFunc("/{id}", spSession.DeleteSession(sessionManager)).Methods("DELETE")
	scrumpokerRouter.HandleFunc("/{id}", spSession.AdministerSession(sessionManager)).Methods("PUT")
	scrumpokerRouter.HandleFunc("/{id}/{vote}", spSession.VoteSession(sessionManager)).Methods("PUT")
	scrumpokerRouter.HandleFunc("/{id}/status", spSession.StatusSession(sessionManager)).Methods("GET")
	scrumpokerRouter.HandleFunc("/{id}/reset", spSession.ResetSession(sessionManager)).Methods("GET")

	srv := &http.Server{
		Addr:         ":10000",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	fmt.Println("Scrumpoker REST API Testserver started")
	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	fmt.Println("Cleanup Timer started")
	ticker := time.NewTicker(cleanupInterval)
	tickerChanel := make(chan bool)
	go func() {
		for {
			select {
			case <-tickerChanel:
				return
			case <-ticker.C:
				sessionManager.Cleanup()
			}
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)

	ticker.Stop()
	tickerChanel <- true

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}
