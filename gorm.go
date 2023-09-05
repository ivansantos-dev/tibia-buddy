package main

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"log"
	"time"
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

type UserSetting struct {
	gorm.Model
	Name                     string
	PushNotificationEnabled  bool
	EmailNotificationEnabled bool
	NotificationEmails       string
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
	Status FormerNameStatus
	UserId string
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
	db.AutoMigrate(&UserSetting{})
	db.AutoMigrate(&FormerName{})
	db.AutoMigrate(&VipFriend{})

	return db
}

func GetVipList(db *gorm.DB, userId string) []Player {
	var vipList []Player
	db.Model(&Player{}).Where("user_id = ?", userId).Joins("left join vip_friends on players.id = vip_friends.player_id").Scan(&vipList)

	return vipList
}

func GetFormerNames(db *gorm.DB, userId string) []FormerName {
	var formerNames []FormerName
	db.Where("user_id = ?", userId).Find(&formerNames)

	return formerNames
}

func GetUserSettings(db *gorm.DB, userId string) UserSetting {
	var user UserSetting
	db.First(&user, userId)

	return user
}
