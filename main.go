package main

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type IndexPageData struct {
	VipList []Player
}

var userId = "1"

func main() {
	db := initializeGorm()
	go CheckWorlds(db)
	go CheckFormerNames(db)

	router := mux.NewRouter()

	router.Handle("/static/", http.FileServer(http.Dir(".")))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/index.html", "templates/vip-table.html")
		if err != nil {
			log.Error("missing file", err)
		}

		data := IndexPageData{
			VipList: GetVipList(db, userId)}
		tmpl.Execute(w, data)
	})

	router.HandleFunc("/vip-list", func(w http.ResponseWriter, r *http.Request) {
		characterName := r.PostFormValue("name")
		characterId := strings.ToLower(characterName)
		apiChar, err := GetCharacter(characterName)
		if err != nil {
			log.Error(err)
		}
		log.WithFields(log.Fields{"name": characterId, "userId": userId}).Info("adding vip friend")
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
	})

	router.HandleFunc("/vip-list/{player}", func(w http.ResponseWriter, r *http.Request) {
		log.Info("here")
		vars := mux.Vars(r)

		playerId := strings.ToLower(vars["player"])

		log.WithField("playerId", playerId).Info("deleting vip friend")
		db.Unscoped().Where("player_id = ? AND user_id",playerId, userId).Delete(&VipFriend{})


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
