package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"internal/models"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

type indexPage struct {
	Title   string
	Error   string
	IsAdmin bool
}

type ScrumpokerSession struct {
	User    *models.User
	Session *models.ScrumpokerSession
	Uuid    string
}

const (
	Uuid_token  = "uuid_token"
	Error_token = "error_token"
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	indexPage := indexPage{Title: "Scrumpoker Example"}
	c, _ := r.Cookie(Error_token)
	if c != nil {
		indexPage.Error = c.Value
	}

	t, _ := template.ParseFiles("web/template/index.html")
	t.Execute(w, indexPage)
}

func JoinSession(manager *models.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type createSessionStruct struct {
			Username  string
			SessionID string
		}

		err := r.ParseForm()
		if err != nil {
			RedirectToMainandDisplayError(err.Error(), &w, r)
			return
		}

		formData := new(createSessionStruct)
		decoder := schema.NewDecoder()
		err = decoder.Decode(formData, r.Form)
		if err != nil {
			RedirectToMainandDisplayError(err.Error(), &w, r)
			return
		}

		user, err := models.InitilaizeUser(formData.Username)
		if err != nil {
			RedirectToMainandDisplayError(err.Error(), &w, r)
			return
		}
		session, err := manager.JoinSession(formData.SessionID, *user)
		if err != nil {
			RedirectToMainandDisplayError(err.Error(), &w, r)
			return
		}

		setCookie(&w, *user)
		http.Redirect(w, r, "session/"+session.SessionID, http.StatusMovedPermanently)
	}
}

func CreateSession(manager *models.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type createSessionStruct struct {
			Username string
			Cards    string
		}

		err := r.ParseForm()
		if err != nil {
			RedirectToMainandDisplayError(err.Error(), &w, r)
			return
		}

		formData := new(createSessionStruct)
		decoder := schema.NewDecoder()
		err = decoder.Decode(formData, r.Form)
		if err != nil {
			RedirectToMainandDisplayError(err.Error(), &w, r)
			return
		}

		user, err := models.InitilaizeUser(formData.Username)
		if err != nil {
			RedirectToMainandDisplayError(err.Error(), &w, r)
			return
		}

		session, err := manager.AddSession(*user, strings.Split(formData.Cards, ","))
		if err != nil {
			RedirectToMainandDisplayError(err.Error(), &w, r)
			return
		}

		setCookie(&w, *user)
		http.Redirect(w, r, "session/"+session.SessionID, http.StatusMovedPermanently)
	}
}

func (s *ScrumpokerSession) ViewSession(manager *models.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type sessionStruct struct {
			indexPage
			Username  string
			Bearer    string
			SessionID string
			Intervall int

			Users      []models.User
			CardsArray []string
		}

		data := sessionStruct{
			Username:   s.User.Name,
			Bearer:     s.User.UUID,
			CardsArray: s.Session.Cards,
			Users:      s.Session.Users,
			SessionID:  s.Session.SessionID,
			Intervall:  2500,
		}

		data.Title = "Scrumpoker"
		data.IsAdmin = s.Session.IsAdmin(s.Uuid)

		t, err := template.ParseFiles("web/template/poker.html")
		if err != nil {
			RedirectToMainandDisplayError(err.Error(), &w, r)
			return
		}
		t.Execute(w, data)
	}
}

func (s *ScrumpokerSession) DeleteSession(manager *models.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := manager.DeleteSession(s.Session.SessionID, *s.User)
		if err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}
}

func (s *ScrumpokerSession) VoteSession(manager *models.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Session.Vote(s.User.UUID, mux.Vars(r)["vote"])
	}
}

func (s *ScrumpokerSession) ResetSession(manager *models.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Session.ResetVotes()
	}
}

func (s *ScrumpokerSession) StatusSession(manager *models.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonResponse, jsonError := json.Marshal(s.Session.Status(s.Uuid))
		if jsonError != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
	}
}

func (s *ScrumpokerSession) AdministerSession(manager *models.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type administerStruct struct {
			ShouldShow bool
		}

		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = s.Session.GetAdmin(s.Uuid)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		formData := administerStruct{}
		if err := json.NewDecoder(r.Body).Decode(&formData); err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		s.Session.AdministerSession(formData.ShouldShow)
	}
}

func setCookie(w *http.ResponseWriter, user models.User) {
	http.SetCookie(*w, &http.Cookie{
		Name:    Uuid_token,
		Value:   user.UUID,
		Expires: time.Now().Add(6 * time.Second),
	})
}

func RedirectToMainandDisplayError(err string, w *http.ResponseWriter, r *http.Request) {
	http.SetCookie(*w, &http.Cookie{
		Name:    Error_token,
		Value:   err,
		Expires: time.Now().Add(6 * time.Second),
		Path:    "/",
	})
	http.Redirect(*w, r, "/", http.StatusMovedPermanently)
}
