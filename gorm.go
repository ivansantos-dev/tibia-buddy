package main

import (
	"log"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Player struct {
	ID        string `gorm:"primaryKey"`
	Name      string
	World     string
	IsOnline  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type World struct {
	Name      string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Profile struct {
	gorm.Model
	Name                    string
	EnablePushNotification  bool
	EnableEmailNotification bool
	NotificationEmail       string
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

type VipFriend struct {
	gorm.Model
	UserId     string
	PlayerName string
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
	db.AutoMigrate(&VipFriend{})

	return db
}

func GetVipList(db *gorm.DB, userId string) []Player {
	vipFriends := make([]VipFriend, 0)
	db.Where("user_id = ?", userId).Find(&vipFriends)

	var vipNames []string
	for _, friend := range vipFriends {
		vipNames = append(vipNames, strings.ToLower( friend.PlayerName))
	}

	var vipList []Player
	db.Where("id in ?", vipNames).Find(&vipList)

	return vipList
}

// func GetWorld() World, error {
// 	var world World
// 	db.Find()
// 	return nil, nil
// }
