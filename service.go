package main

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strings"
)

type FormerNameService struct {
	db *gorm.DB
}

func (s *FormerNameService) CreateFormerName(formerName string) error {
	apiChar, err := GetCharacter(formerName)
	if err != nil {
		log.Error(err)
	}

	status := expiring
	if strings.EqualFold(apiChar.CharacterInfo.Name, formerName) {
		status = claimed
	}

	if apiChar.CharacterInfo.World == "" {
		status = available
	} else {
		formerName = apiChar.CharacterInfo.Name
	}

	if status == expiring {
		for _, actualFormerName := range apiChar.CharacterInfo.FormerNames {
			if strings.EqualFold(actualFormerName, formerName) {
				break
			}
		}
	}

	log.WithFields(log.Fields{"name": formerName}).Info("add former name")
	s.db.Where("name = ?", formerName).FirstOrCreate(&FormerName{Name: formerName, Status: status})

	return nil
}

type VipListService struct {
	db *gorm.DB
}

func (s *VipListService) CreateVipListFriend(name string) error {
	apiChar, err := GetCharacter(name)
	if err != nil {
		log.Error(err)
	}

	characterName := apiChar.CharacterInfo.Name

	if characterName == "" {
		return errors.New("character not found")
	}

	characterId := strings.ToLower(characterName)
	log.WithFields(log.Fields{"name": characterName}).Info("adding vip friend")

	player := Player{ID: characterId, Name: apiChar.CharacterInfo.Name, World: apiChar.CharacterInfo.World}
	s.db.FirstOrCreate(&player)

	var world = World{Name: apiChar.CharacterInfo.World}
	s.db.FirstOrCreate(&world)

	return nil
}
