package main

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var sessionName = "auth"
var store = InitializeSessionStore()

func InitializeSessionStore() *sessions.CookieStore {
	key := "update me at some point"
	store := sessions.NewCookieStore([]byte(key))
	store.Options.Path = "/"
	store.Options.HttpOnly = true

	gothic.Store = store
	GOOGLE_OAUTH_CLIENT_ID := "470889723490-621rbcg3hk06stb2ubrujnu70anf1ugr.apps.googleusercontent.com"
	GOOGLE_OAUTH_CLIENT_SECRET := "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"
	goth.UseProviders(
		google.New(GOOGLE_OAUTH_CLIENT_ID, GOOGLE_OAUTH_CLIENT_SECRET, "http://localhost:8090/auth/google/callback", "email", "profile"),
	)

	return store
}

func AddUserToSession(wr http.ResponseWriter, req *http.Request, user goth.User) {
	session, err := store.Get(req, sessionName)
	if err != nil {
		logrus.Print("Error ", err)
	}

	// Remove the raw data to reduce the size
	user.RawData = map[string]interface{}{}

	session.Values["user"] = user
	err = session.Save(req, wr)
	if err != nil {
		log.Error("failed to save session", err)
	}
}

func RemoveUserFromSession(wr http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, sessionName)
	session.Values["user"] = goth.User{}
	session.Save(req, wr)
}
