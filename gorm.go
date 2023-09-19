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
}

type DB struct {
	db *gorm.DB
}

func (c *DB) InitializeGorm() {
	dbName := "test.db"
	log.Println("Initialize Gorm in " + dbName)

	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database", err)
	}

	db.AutoMigrate(&Player{})
	db.AutoMigrate(&World{})
	db.AutoMigrate(&FormerName{})

	c.db = db
}

func (c *DB) GetVipList(names []string) []Player {
	var vipList []Player
	c.db.Where("name IN ?", names).Find(&vipList)
	return vipList
}

func (c *DB) GetFormerNames(names []string) []FormerName {
	var formerNames []FormerName
	c.db.Where("name IN ?", names).Find(&formerNames)
	return formerNames
}
