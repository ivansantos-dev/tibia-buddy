package main

import (
	"html/template"
	"net/http"
	"strings"

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

	http.Handle("/static/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/index.html", "templates/vip-table.html")
		if err != nil {
			log.Error("missing file", err)
		}

		data := IndexPageData{
			VipList: GetVipList(db, userId)}
		tmpl.Execute(w, data)
	})

	http.HandleFunc("/vip-list/add", func(w http.ResponseWriter, r *http.Request) {
		characterName := r.PostFormValue("name")
		characterId := strings.ToLower(characterName)
		apiChar, err := GetCharacter(characterName)
		if err != nil {
			log.Error(err)
		}
		player := Player{ID: characterId, Name: apiChar.CharacterInfo.Name, World: apiChar.CharacterInfo.World}
		db.FirstOrCreate(&player)

		var world = World{Name: apiChar.CharacterInfo.World}
		db.FirstOrCreate(&world)

		log.WithFields(log.Fields{"name": player.Name, "userId": userId}).Info("adding vip friend")
		vipFriend := VipFriend{UserId: userId, PlayerName: player.Name}
		result2 := db.Where(&vipFriend).FirstOrCreate(&vipFriend)
		log.Info(result2.RowsAffected, result2.Error)


		tmpl, err := template.ParseFiles("templates/vip-table.html")
		if err != nil {
			log.Error("missing file", err)
		}

		tmpl.ExecuteTemplate(w, "vip-list-table", GetVipList(db, userId))
	})

	port := "127.0.0.1:8090"
	log.Info("Listening to port: " + port)
	log.Fatal(http.ListenAndServe(port, nil))
}
