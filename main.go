package main

import (
	"encoding/gob"
	"net/http"
	"strings"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type IndexPageData struct {
	IsLoggedIn           bool
	VipListTableData     VipListTableData
	FormerNamesTableData FormerNamesTableData
}

type VipListTableData struct {
	Error   string
	VipList []Player
}

type FormerNamesTableData struct {
	Error       string
	FormerNames []FormerName
}
type ProfilePageData struct {
	IsLoggedIn  bool
	UserSetting UserSetting
}

var userId = "1"

func CreateFormerName(db *gorm.DB, formerName string) error {
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
	db.Where("name = ? AND user_id = ?", actualName, userId).FirstOrCreate(&FormerName{Name: actualName, Status: status, UserId: userId})

	return nil
}

func DeleteFormerName(db *gorm.DB, formerName string) error {
	log.WithField("formerName", formerName).Info("deleting former name")
	db.Unscoped().Where("name = ? AND user_id = ?", formerName, userId).Delete(&FormerName{})
	return nil
}

func CreateVipListFriend(db *gorm.DB, name string) error {
	apiChar, err := GetCharacter(name)
	if err != nil {
		log.Error(err)
	}

	characterName := apiChar.CharacterInfo.Name
	characterId := strings.ToLower(characterName)
	log.WithFields(log.Fields{"name": characterName, "userId": userId}).Info("adding vip friend")
	vipFriend := VipFriend{UserId: userId, PlayerId: characterId}
	result2 := db.Where(&vipFriend).FirstOrCreate(&vipFriend)
	log.Info(result2.RowsAffected, result2.Error)

	if result2.RowsAffected > 0 {
		player := Player{ID: characterId, Name: apiChar.CharacterInfo.Name, World: apiChar.CharacterInfo.World}
		db.FirstOrCreate(&player)

		var world = World{Name: apiChar.CharacterInfo.World}
		db.FirstOrCreate(&world)
	}
	return nil

}
func DeleteVipListFriend(db *gorm.DB, name string) error {
	playerId := strings.ToLower(name)
	log.WithField("playerId", playerId).Info("deleting vip friend")
	db.Unscoped().Where("player_id = ? AND user_id = ?", playerId, userId).Delete(&VipFriend{})

	return nil
}

func main() {
	db := initializeGorm()
	gob.Register(goth.User{})
	go CheckWorlds(db)
	go CheckFormerNames(db)

	r := gin.Default()

	r.LoadHTMLGlob("./templates/**/*")
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		isLoggedIn := false
		session, _ := store.Get(c.Request, sessionName)
		user := session.Values["user"]
		log.WithField("user", user).Info("index")

		if user != nil && user.(goth.User).IDToken != "" {
			isLoggedIn = true
		}
		log.WithFields(log.Fields{"user": user, "isLoggedIn": isLoggedIn}).Info("index")
		data := IndexPageData{
			IsLoggedIn: isLoggedIn,
		}

		c.HTML(200, "index.html", data)
	})

	r.GET("/vip-list", func(c *gin.Context) {
		c.HTML(200, "VipListTable.html", GetVipList(db, userId))
	})

	r.POST("/vip-list", func(c *gin.Context) {
		CreateVipListFriend(db, c.PostForm("name"))
		c.HTML(200, "VipListTable.html", GetVipList(db, userId))
	})

	r.DELETE("/vip-list/:name", func(c *gin.Context) {
		DeleteVipListFriend(db, c.Params.ByName("name"))
		c.HTML(200, "VipListTable.html", GetVipList(db, userId))
	})

	r.GET("/former-names", func(c *gin.Context) {
		c.HTML(200, "FormerNamesTable.html", GetFormerNames(db, userId))
	})

	r.POST("/former-names", func(c *gin.Context) {
		CreateFormerName(db, c.PostForm("name"))
		c.HTML(200, "FormerNamesTable.html", GetFormerNames(db, userId))
	})

	r.DELETE("/former-names/:formerName", func(c *gin.Context) {
		DeleteFormerName(db, c.Params.ByName("formerName"))
		c.HTML(200, "FormerNamesTable.html", GetFormerNames(db, userId))
	})

	r.GET("/login/:provider", func(c *gin.Context) {
		q := c.Request.URL.Query()
		q.Add("provider", c.Params.ByName("provider"))
		c.Request.URL.RawQuery = q.Encode()
		gothic.BeginAuthHandler(c.Writer, c.Request)
	})

	r.GET("/auth/:provider/callback", func(c *gin.Context) {
		w := c.Writer
		r := c.Request
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			log.Error(err)
		}
		AddUserToSession(w, r, user)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	r.GET("/logout", func(c *gin.Context) {
		q := c.Request.URL.Query()
		q.Add("provider", c.Params.ByName("provider"))
		c.Request.URL.RawQuery = q.Encode()

		w := c.Writer
		r := c.Request
		log.Println("Logging out")
		err := gothic.Logout(w, r)
		if err != nil {
			log.Print("Logout fail", err)
		}
		RemoveUserFromSession(w, r)
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})

	r.Run("127.0.0.1:8090")
}
