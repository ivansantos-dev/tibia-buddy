package main

import (
	"encoding/gob"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	log "github.com/sirupsen/logrus"
)

type IndexPageData struct {
	IsLoggedIn           bool
	VipListTableData     VipListTableData
	FormerNamesTableData FormerNamesTableData
}

type VipListTableData struct {
	Error   string
	VipList []Player
}

type FormerNamesTableData struct {
	Error       string
	FormerNames []FormerName
}
type ProfilePageData struct {
	IsLoggedIn  bool
	UserSetting UserSetting
}

var userId = "1"


func main() {
	db := initializeGorm()
	gob.Register(goth.User{})
	go CheckWorlds(db)
	go CheckFormerNames(db)


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
		AddUserToSession(w, r, user)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	router.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Logging out")
		err := gothic.Logout(w, r)
		if err != nil {
			log.Print("Logout fail", err)
		}
		RemoveUserFromSession(w, r)
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
			IsLoggedIn:           isLoggedIn,
			VipListTableData:     VipListTableData{VipList: GetVipList(db, userId)},
			FormerNamesTableData: FormerNamesTableData{FormerNames: GetFormerNames(db, userId)},
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

		tmpl.ExecuteTemplate(w, "former-names-table", FormerNamesTableData{FormerNames: GetFormerNames(db, userId)})
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

		tmpl.ExecuteTemplate(w, "former-names-table", FormerNamesTableData{FormerNames: GetFormerNames(db, userId)})
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

		tmpl.ExecuteTemplate(w, "former-names-table", FormerNamesTableData{FormerNames: GetFormerNames(db, userId)})
	})

	router.HandleFunc("/vip-list", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/vip-table.html")
		if err != nil {
			log.Error("missing file", err)
		}

		tmpl.ExecuteTemplate(w, "vip-list-table", VipListTableData{VipList: GetVipList(db, userId)})
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

		tmpl.ExecuteTemplate(w, "vip-list-table", VipListTableData{VipList: GetVipList(db, userId)})
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

		tmpl.ExecuteTemplate(w, "vip-list-table", VipListTableData{VipList: GetVipList(db, userId)})
	})

	port := "127.0.0.1:8090"
	log.Info("Listening to port: " + port)

	log.Fatal(http.ListenAndServe(port, router))
}
