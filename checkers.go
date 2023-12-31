package main

import (
	"time"
)

var worldCheckSleep = 30 * time.Second
var formerNameSleep = 1 * time.Minute

//func CheckWorlds(db *gorm.DB) {
//	var worlds []World
//	db.Find(&worlds)
//
//	for {
//		for _, world := range worlds {
//			log.WithField("world", world.Name).Info("Checking world")
//			worldInfo, err := GetWorld(world.Name)
//			if err != nil {
//				log.Error("Failed to retrieve world from tibia data api", err)
//			}
//
//			worldName := worldInfo.Name
//
//			playerNames := make([]string, len(worldInfo.OnlinePlayers))
//			players := make([]Player, len(worldInfo.OnlinePlayers))
//			for i, player := range worldInfo.OnlinePlayers {
//				playerNames[i] = strings.ToLower(player.Name)
//				players[i] = Player{Name: player.Name, ID: strings.ToLower(player.Name), World: worldName}
//			}
//
//			tx := db.Begin()
//			tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&players)
//			tx.Model(&Player{}).Where("world = ?", worldName).Update("is_online", false)
//			tx.Model(&Player{}).Where("id IN ?", playerNames).Update("is_online", true)
//			tx.Commit()
//
//			log.WithFields(log.Fields{"world": worldName, "player count": len(players)}).Info("Finish processing world.")
//			time.Sleep(3 * time.Second)
//		}
//
//		time.Sleep(worldCheckSleep)
//	}
//}
//
//func CheckFormerNames(db *gorm.DB) {
//	var formerNames []FormerName
//	db.Where("status = ?", expiring).Find(&formerNames)
//
//	availableList := make([]string, len(formerNames))
//	claimedList := make([]string, len(formerNames))
//	for {
//		for _, formerName := range formerNames {
//			expiringName := formerName.Name
//			log.WithField("name", expiringName).Info("Checking former name")
//			apiCharacter, err := GetCharacter(expiringName)
//			if err != nil {
//				log.Error("Failed to retrieve former name from tibia data api", err)
//			}
//
//			if expiringName == apiCharacter.CharacterInfo.Name {
//				claimedList = append(claimedList, expiringName)
//			}
//
//			if apiCharacter.CharacterInfo.World == "" {
//				availableList = append(availableList, expiringName)
//			}
//
//			time.Sleep(1 * time.Second)
//		}
//
//		db.Model(&FormerName{}).Where("name IN ?", availableList).Update("status", available)
//		db.Model(&FormerName{}).Where("name IN ?", claimedList).Update("status", claimed)
//
//		time.Sleep(worldCheckSleep)
//	}
//
//}
