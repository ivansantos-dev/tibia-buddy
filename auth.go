package main

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"

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
	googleOauthClientId := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	googleOauthClientSecret := os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
	googleOauthCallBackUrl := os.Getenv("GOOGLE_OAUTH_CALLBACK_URL")

	goth.UseProviders(
		google.New(googleOauthClientId, googleOauthClientSecret, googleOauthCallBackUrl, "email", "profile"),
	)

	return store
}

func AddUserToSession(wr http.ResponseWriter, req *http.Request, user goth.User) {
	session, err := store.Get(req, sessionName)
	if err != nil {
		log.Print("Error ", err)
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
