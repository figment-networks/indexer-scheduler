package auth

import (
	"errors"
	"net/http"
)

type AuthCredentials struct {
	User     string
	Password string
}

var ErrUnauthenticated = errors.New("unauthenticated")

func BasicAuth(ac AuthCredentials, w http.ResponseWriter, r *http.Request) error {
	if ac.User == "" || ac.Password == "" {
		return nil
	}

	user, pass, _ := r.BasicAuth()

	if ac.User != user || ac.Password != pass {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized.", http.StatusUnauthorized)
		return ErrUnauthenticated
	}
	return nil
}
