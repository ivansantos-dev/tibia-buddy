package main

import (
	"log"
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
	UserId   string
	PlayerId string
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
	var vipList []Player
	db.Model(&Player{}).Where("user_id = ?", userId).Joins("left join vip_friends on players.id = vip_friends.player_id").Scan(&vipList)

	return vipList
}

// func GetWorld() World, error {
// 	var world World
// 	db.Find()
// 	return nil, nil
// }
