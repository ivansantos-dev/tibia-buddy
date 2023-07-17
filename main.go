package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"strings"

	"gorm.io/gorm"
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
			log.Panic("missing file", err)
		}

		data := IndexPageData{
			VipList: GetVipList(db, userId)}
		tmpl.Execute(w, data)
	})

	http.HandleFunc("/vip-list/add", func(w http.ResponseWriter, r *http.Request) {
		characterName := r.PostFormValue("name")	
		characterId := strings.ToLower(characterName)
		player := Player{ID: characterId}
		result := db.Limit(1).Find(&player)
		if result.RowsAffected = {
			apiChar, err := GetCharacter(characterName)
			if err != nil {
				log.Println(err)
			}

			player = Player{ID: characterId, Name: apiChar.CharacterInfo.Name, World: apiChar.CharacterInfo.World} 
			db.FirstOrCreate(&player)
			var world = World{Name: apiChar.CharacterInfo.World}
			db.FirstOrCreate(&world)
		}

		vipFriend := VipFriend{UserId: userId, PlayerName: player.Name}
		db.FirstOrCreate(&vipFriend)

		tmpl, err := template.ParseFiles("templates/vip-table.html")
		if err != nil {
			log.Println("[ERROR] missing file", err)
		}

		tmpl.ExecuteTemplate(w, "vip-list-table", GetVipList(db, userId))
	})

	port := "127.0.0.1:8090"
	log.Println("Listening to port: " + port)
	log.Fatal(http.ListenAndServe(port, nil))
}
