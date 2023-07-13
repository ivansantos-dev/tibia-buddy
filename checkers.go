package main

import (
	"gorm.io/gorm"
	"log"
	"time"
)

var worldCheckSleep = 30 * time.Second
var formerNameSleep = 1 * time.Minute

func CheckWorlds(db *gorm.DB) {
	var worlds []World
	db.Find(&worlds)

	for {
		for _, world := range worlds {
			log.Printf("Checking world: %s\n", world.Name)
			worldInfo, err := GetWorld(world.Name)
			if err != nil {
				log.Println("[ERROR] retrieving world from tibia data api", err)
			}

			worldName := worldInfo.Name

			playerNames := make([]string, len(worldInfo.OnlinePlayers))
			players := make([]Player, len(worldInfo.OnlinePlayers))
			for i, player := range worldInfo.OnlinePlayers {
				playerNames[i] = player.Name
				players[i] = Player{Name: player.Name, World: worldName}
			}

			log.Printf("World: %s, Online Player Count: %d\n", worldName, len(players))
			tx := db.Begin()
			tx.Create(&players)
			tx.Model(&Player{}).Where("world = ? AND is_online = ?", worldInfo.Name, false).Update("status", false)
			tx.Model(&Player{}).Where("name IN ?", playerNames).Update("status", true)
			tx.Commit()
		}

		time.Sleep(worldCheckSleep)
	}
}

func CheckFormerNames(db *gorm.DB) {
	var formerNames []FormerName
	db.Where("status = ?", expiring).Find(&formerNames)

	availableList := make([]string, len(formerNames))
	claimedList := make([]string, len(formerNames))
	for {
		for _, formerName := range formerNames {
			expiringName := formerName.Name
			log.Printf("Checking former name: %s\n", expiringName)
			apiCharacter, err := GetCharacter(expiringName)
			if err != nil {
				log.Println("[ERROR] retrieving former name from tibia data api", err)
			}

			if expiringName == apiCharacter.CharacterInfo.Name {
				claimedList = append(claimedList, expiringName)
			}

			if apiCharacter.CharacterInfo.World == "" {
				availableList = append(availableList, expiringName)
			}

			time.Sleep(1 * time.Second)
		}

		db.Model(&FormerName{}).Where("name IN ?", availableList).Update("status", available)
		db.Model(&FormerName{}).Where("name IN ?", claimedList).Update("status", claimed)

		time.Sleep(worldCheckSleep)
	}

}
