package session

import (
	"time"

	"github.com/gorilla/sessions"
)

var Store = sessions.NewCookieStore([]byte("SESSION_KEY"))

func SessionOptions() {
	Store.MaxAge(int(12 * 24 * time.Hour / time.Second))
}
