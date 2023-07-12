package main

import (
  "gorm.io/gorm"
  "gorm.io/driver/sqlite"
  "net/http"
  "log"
)

type Status string
const (
  Online Status = "online"
  Offline Status = "offline" 
)

type Player struct { 
  gorm.Model
  Name string
  World string
  Status  Status
}

type World struct {
  gorm.Model
  Name string
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

  return db

}

func main() {
  initializeGorm()

  log.Println("Listening to port: 8090")
  log.Fatal(http.ListenAndServe(":8090", nil))
}
