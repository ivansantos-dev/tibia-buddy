package main

import (
	"encoding/gob"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
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
		data := IndexPageData{
			IsLoggedIn: false,
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

	r.Run("127.0.0.1:8090")

	router := mux.NewRouter()
	router.HandleFunc("/login/{provider}", func(w http.ResponseWriter, r *http.Request) {
		gothic.BeginAuthHandler(w, r)
	})

	router.HandleFunc("/auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			log.Error(err)
		}
		AddUserToSession(w, r, user)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	router.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Logging out")
		err := gothic.Logout(w, r)
		if err != nil {
			log.Print("Logout fail", err)
		}
		RemoveUserFromSession(w, r)
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		isLoggedIn := true
		session, _ := store.Get(r, sessionName)
		user := session.Values["user"]
		if user == nil || user.(goth.User).IDToken == "" {
			isLoggedIn = false
		}

		tmpl, err := template.ParseFiles("templates/layout.html", "templates/index.html", "templates/vip-table.html", "templates/former-names-table.html")
		if err != nil {
			log.Error("missing file", err)
		}

		data := IndexPageData{
			IsLoggedIn:           isLoggedIn,
			VipListTableData:     VipListTableData{VipList: GetVipList(db, userId)},
			FormerNamesTableData: FormerNamesTableData{FormerNames: GetFormerNames(db, userId)},
		}
		tmpl.Execute(w, data)
	})
}
