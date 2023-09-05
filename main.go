package main

import (
	"encoding/gob"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	log "github.com/sirupsen/logrus"

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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db := initializeGorm()
	formerNameService := &FormerNameService{db: db}
	vipListService := &VipListService{db: db}
	gob.Register(goth.User{})
	stream := NewServer()

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
		vipListService.CreateVipListFriend(c.PostForm("name"))
		c.HTML(200, "VipListTable.html", GetVipList(db, userId))
	})

	r.DELETE("/vip-list/:name", func(c *gin.Context) {
		vipListService.DeleteVipListFriend(c.Params.ByName("name"))
		c.HTML(200, "VipListTable.html", GetVipList(db, userId))
	})

	r.GET("/former-names", func(c *gin.Context) {
		c.HTML(200, "FormerNamesTable.html", GetFormerNames(db, userId))
	})

	r.POST("/former-names", func(c *gin.Context) {
		formerNameService.CreateFormerName(c.PostForm("name"))
		c.HTML(200, "FormerNamesTable.html", GetFormerNames(db, userId))
	})

	r.DELETE("/former-names/:formerName", func(c *gin.Context) {
		formerNameService.DeleteFormerName(c.Params.ByName("formerName"))
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

	r.GET("/stream", HeadersMiddleware(), stream.serveHTTP(), func(c *gin.Context) {
		v, ok := c.Get("clientChan")
		if !ok {
			return
		}
		clientChan, ok := v.(ClientChan)
		if !ok {
			return
		}
		c.Stream(func(w io.Writer) bool {
			// Stream message to client from message channel
			if msg, ok := <-clientChan; ok {
				c.SSEvent("message", msg)
				return true
			}

			return false
		})
	})

	go func() {
		for {
			time.Sleep(time.Second * 10)
			now := time.Now().Format("2006-01-02 15:04:05")
			currentTime := fmt.Sprintf("The Current Time Is %v", now)

			// Send current time to clients message channel
			stream.Message <- currentTime
		}
	}()

	r.Run(os.Getenv("SERVER_URL"))
}
