package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
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

func main() {
	db := initializeGorm()
	go CheckWorlds(db)
	go CheckFormerNames(db)

	port := ":8090"
	log.Println("Listening to port: " + port)
	// log.Fatal(http.ListenAndServe(port, nil))
}
