package main

import (
	"encoding/gob"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	log "github.com/sirupsen/logrus"
)

type IndexPageData struct {
	IsLoggedIn  bool
	VipList     []Player
	FormerNames []FormerName
}

type ProfilePageData struct {
	IsLoggedIn  bool
	UserSetting UserSetting
}

var userId = "1"
var sessionName = "auth"
var store = initializeSessionStore()

func initializeSessionStore() *sessions.CookieStore {
	key := "update me at some point"
	store := sessions.NewCookieStore([]byte(key))
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	return store
}

func addUserToSession(wr http.ResponseWriter, req *http.Request, user goth.User) {
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

func removeUserFromSession(wr http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, sessionName)
	session.Values["user"] = goth.User{}
	session.Save(req, wr)
}

func main() {
	db := initializeGorm()
	gob.Register(goth.User{})
	go CheckWorlds(db)
	go CheckFormerNames(db)

	gothic.Store = store
	GOOGLE_OAUTH_CLIENT_ID := "470889723490-621rbcg3hk06stb2ubrujnu70anf1ugr.apps.googleusercontent.com"
	GOOGLE_OAUTH_CLIENT_SECRET := "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"
	goth.UseProviders(
		google.New(GOOGLE_OAUTH_CLIENT_ID, GOOGLE_OAUTH_CLIENT_SECRET, "http://localhost:8090/auth/google/callback", "email", "profile"),
	)

	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	router.HandleFunc("/login/{provider}", func(w http.ResponseWriter, r *http.Request) {
		gothic.BeginAuthHandler(w, r)
	})

	router.HandleFunc("/auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			log.Error(err)
		}
		addUserToSession(w, r, user)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	router.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Logging out")
		err := gothic.Logout(w, r)
		if err != nil {
			log.Print("Logout fail", err)
		}
		removeUserFromSession(w, r)
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		isLoggedIn := true
		session, _ := store.Get(r, sessionName)
		user := session.Values["user"]
		if user == nil || user.(goth.User).IDToken == "" {
			isLoggedIn = false

		}

		tmpl, err := template.ParseFiles("templates/layout.html", "templates/index.html", "templates/vip-table.html", "templates/former-names-table.html")
		if err != nil {
			log.Error("missing file", err)
		}

		data := IndexPageData{
			IsLoggedIn:  isLoggedIn,
			VipList:     GetVipList(db, userId),
			FormerNames: GetFormerNames(db, userId),
		}
		tmpl.Execute(w, data)
	})

	router.HandleFunc("/settings", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/layout.html", "templates/profile.html")
		if err != nil {
			log.Error("missing file", err)
		}

		tmpl.Execute(w, ProfilePageData{IsLoggedIn: true, UserSetting: GetUserSettings(db, userId)})
	})

	router.HandleFunc("/former-names", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/former-names-table.html")
		if err != nil {
			log.Error("missing file", err)
		}

		tmpl.ExecuteTemplate(w, "former-names-table", GetFormerNames(db, userId))
	}).Methods("GET")

	router.HandleFunc("/former-names", func(w http.ResponseWriter, r *http.Request) {
		formerName := r.PostFormValue("name")
		apiChar, err := GetCharacter(formerName)
		if err != nil {
			log.Error(err)
		}

		status := expiring
		if strings.EqualFold(apiChar.CharacterInfo.Name, formerName) {
			status = claimed
		}

		if apiChar.CharacterInfo.World == "" {
			status = available
		}

		var actualName string
		if status == expiring {
			for _, actualFormerName := range apiChar.CharacterInfo.FormerNames {
				if strings.EqualFold(actualFormerName, formerName) {
					actualName = actualFormerName
					break
				}
			}
		}

		log.WithFields(log.Fields{"name": actualName, "userId": userId}).Info("add former name")
		db.Where("name = ? AND user_id = ?", actualName, userId).FirstOrCreate(&FormerName{Name: actualName, Status: status, UserId: userId})

		tmpl, err := template.ParseFiles("templates/former-names-table.html")
		if err != nil {
			log.Error("missing file", err)
		}

		tmpl.ExecuteTemplate(w, "former-names-table", GetFormerNames(db, userId))
	}).Methods("POST")

	router.HandleFunc("/former-names/{name}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		formerName := vars["name"]

		log.WithField("formerName", formerName).Info("deleting former name")
		db.Unscoped().Where("name = ? AND user_id = ?", formerName, userId).Delete(&FormerName{})

		tmpl, err := template.ParseFiles("templates/former-names-table.html")
		if err != nil {
			log.Error("missing file", err)
		}

		tmpl.ExecuteTemplate(w, "former-names-table", GetFormerNames(db, userId))
	})

	router.HandleFunc("/vip-list", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/vip-table.html")
		if err != nil {
			log.Error("missing file", err)
		}

		tmpl.ExecuteTemplate(w, "vip-list-table", GetVipList(db, userId))
	}).Methods("GET")

	router.HandleFunc("/vip-list", func(w http.ResponseWriter, r *http.Request) {
		characterName := r.PostFormValue("name")
		apiChar, err := GetCharacter(characterName)
		if err != nil {
			log.Error(err)
		}

		characterName = apiChar.CharacterInfo.Name
		characterId := strings.ToLower(characterName)
		log.WithFields(log.Fields{"name": characterName, "userId": userId}).Info("adding vip friend")
		vipFriend := VipFriend{UserId: userId, PlayerId: characterId}
		result2 := db.Where(&vipFriend).FirstOrCreate(&vipFriend)
		log.Info(result2.RowsAffected, result2.Error)

		if result2.RowsAffected > 0 {
			player := Player{ID: characterId, Name: apiChar.CharacterInfo.Name, World: apiChar.CharacterInfo.World}
			db.FirstOrCreate(&player)

			var world = World{Name: apiChar.CharacterInfo.World}
			db.FirstOrCreate(&world)
		}

		tmpl, err := template.ParseFiles("templates/vip-table.html")
		if err != nil {
			log.Error("missing file", err)
		}

		tmpl.ExecuteTemplate(w, "vip-list-table", GetVipList(db, userId))
	}).Methods("POST")

	router.HandleFunc("/vip-list/{player}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		playerId := strings.ToLower(vars["player"])

		log.WithField("playerId", playerId).Info("deleting vip friend")
		db.Unscoped().Where("player_id = ? AND user_id = ?", playerId, userId).Delete(&VipFriend{})

		tmpl, err := template.ParseFiles("templates/vip-table.html")
		if err != nil {
			log.Error("missing file", err)
		}

		tmpl.ExecuteTemplate(w, "vip-list-table", GetVipList(db, userId))
	})

	port := "127.0.0.1:8090"
	log.Info("Listening to port: " + port)

	log.Fatal(http.ListenAndServe(port, router))
}
