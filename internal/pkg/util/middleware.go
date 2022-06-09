package util

import (
	"internal/handler"
	"internal/models"
	"net/http"

	"github.com/gorilla/mux"
)

type AuthenticationMiddleware struct {
	Manager           *models.SessionManager
	ScrumpokerSession *handler.ScrumpokerSession
}

func (amw *AuthenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuid := ""
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			//Auth via Cookie (used bacause the Scrumpoker-page is reached via redirect)
			c, err := r.Cookie(handler.Uuid_token)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			uuid = c.Value
		} else { //"Normal" Auth via Bearer-Token
			uuid = authHeader[len("Bearer "):]
		}

		session, err := amw.Manager.GetSession(mux.Vars(r)["id"])
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
