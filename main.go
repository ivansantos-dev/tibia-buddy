package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Player struct {
	gorm.Model
	Name          string
	LowerCaseName string
	World         string
	IsOnline      bool
}

type World struct {
	gorm.Model
	Name string
}

type Profile struct {
	gorm.Model
	Name                    string
	EnablePushNotification  bool
	EnableEmailNotification bool
}

type FormerNameStatus = string

const (
	available FormerNameStatus = "available"
	expiring  FormerNameStatus = "expiring"
	claimed   FormerNameStatus = "claimed"
)

type FormerName struct {
	gorm.Model
	Name   string
	Status string
}

func initializeGorm() *gorm.DB {
	dbName := "test.db"
	log.Println("Initialize Gorm in " + dbName)

	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database", err)
	}

	db.AutoMigrate(&Player{})
	db.AutoMigrate(&World{})
	db.AutoMigrate(&Profile{})
	db.AutoMigrate(&FormerName{})

	return db
}

type IndexPageData struct {
	VipList []Player
}

func main() {
	db := initializeGorm()
	go CheckWorlds(db)
	go CheckFormerNames(db)

	http.Handle("/static/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			log.Panic("missing file", err)
		}

		data := IndexPageData{
			VipList: []Player{
				{Name: "Aragorn", World: "Optera", IsOnline: true},
			},
		}
		tmpl.Execute(w, data)

	})
	http.HandleFunc("/vip-list/add", func(w http.ResponseWriter, r *http.Request) {
		characterName := r.PostFormValue("name")
		log.Println(characterName)
		player := Player{}
		result := db.Where("lower_case_name == ?", strings.ToLower(characterName)).First(&player)
		if result.Error != nil {
			log.Println(result.Error)
			log.Println("Retrieving char")
			apiChar, err := GetCharacter(characterName)
			log.Println("finished get character")
			if err != nil {
				// TODO return error
				log.Println("failed to get character")
			}
			log.Println(apiChar)
			player = Player{Name: apiChar.CharacterInfo.Name, LowerCaseName: strings.ToLower(apiChar.CharacterInfo.Name), World: apiChar.CharacterInfo.World}
			db.Create(&player)
		}

		tmpl, err := template.ParseFiles("templates/vip-table.html")
		if err != nil {
			log.Println("[ERROR] missing file", err)
		}

		data := IndexPageData{
			VipList: []Player{
				{Name: "Aragorn", World: "Optera", IsOnline: true},
				{Name: player.Name, World: player.World, IsOnline: false},
			},
		}
		tmpl.Execute(w, data)
	})

	port := "127.0.0.1:8090"
	log.Println("Listening to port: " + port)
	log.Fatal(http.ListenAndServe(port, nil))
}
