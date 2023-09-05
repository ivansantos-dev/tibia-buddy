package main

import (
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
	}

	var actualName = apiChar.CharacterInfo.Name
	if status == expiring {
		for _, actualFormerName := range apiChar.CharacterInfo.FormerNames {
			if strings.EqualFold(actualFormerName, formerName) {
				actualName = actualFormerName
				break
			}
		}
	}

	log.WithFields(log.Fields{"name": actualName, "userId": userId}).Info("add former name")
	s.db.Where("name = ? AND user_id = ?", actualName, userId).FirstOrCreate(&FormerName{Name: actualName, Status: status, UserId: userId})

	return nil
}

func (s *FormerNameService) DeleteFormerName(formerName string) error {
	log.WithField("formerName", formerName).Info("deleting former name")
	s.db.Unscoped().Where("name = ? AND user_id = ?", formerName, userId).Delete(&FormerName{})
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
	characterId := strings.ToLower(characterName)
	log.WithFields(log.Fields{"name": characterName, "userId": userId}).Info("adding vip friend")
	vipFriend := VipFriend{UserId: userId, PlayerId: characterId}
	result2 := s.db.Where(&vipFriend).FirstOrCreate(&vipFriend)
	log.Info(result2.RowsAffected, result2.Error)

	if result2.RowsAffected > 0 {
		player := Player{ID: characterId, Name: apiChar.CharacterInfo.Name, World: apiChar.CharacterInfo.World}
		s.db.FirstOrCreate(&player)

		var world = World{Name: apiChar.CharacterInfo.World}
		s.db.FirstOrCreate(&world)
	}

	return nil
}

func (s *VipListService) DeleteVipListFriend(name string) error {
	playerId := strings.ToLower(name)
	log.WithField("playerId", playerId).Info("deleting vip friend")
	s.db.Unscoped().Where("player_id = ? AND user_id = ?", playerId, userId).Delete(&VipFriend{})

	return nil
}
