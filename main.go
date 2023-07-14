package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Player struct {
	gorm.Model
	Name     string
	World    string
	IsOnline bool
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
	http.HandleFunc("/vip-list/add/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World")
		time.Sleep(2 * time.Second)
		log.Println("I am here")
	})

	port := ":8090"
	log.Println("Listening to port: " + port)
	log.Fatal(http.ListenAndServe(port, nil))
}
