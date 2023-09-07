package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
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

var userId = "1"
var defaultNames = []string{"Aragorn"}

func filterEmpty(slice []string) []string {
	var result []string
	for _, str := range slice {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}
func GetVipList(c *gin.Context) []string {
	listStr, err := c.Cookie("vip-list")
	if err != nil {
		log.Error("Failed to get cookie", err)
	}
	return filterEmpty(strings.Split(listStr, ","))
}

func AddVipList(c *gin.Context, name string) []string {
	list := GetVipList(c)
	// do not appemd of name is already in list
	for _, n := range list {
		if n == name {
			return list
		}
	}
	list = append(list, name)
	SetVipListCookie(c, list)
	return list
}

func RemoveVipList(c *gin.Context, name string) []string {
	list := GetVipList(c)
	for i, n := range list {
		if n == name {
			if len(list) == 1 {
				list = []string{}
				break
			}
			list = append(list[:i], list[i+1:]...)
		}
	}
	SetVipListCookie(c, list)
	return list
}

func SetVipListCookie(c *gin.Context, names []string) {
	setCookie(c, "vip-list", names)
}

func SetFormerNameCookie(c *gin.Context, names []string) {
	setCookie(c, "former-names", names)
}

func setCookie(c *gin.Context, cookieName string, names []string) {
	if len(names) == 0 {
		c.SetCookie(cookieName, "", -1, "/", os.Getenv("DOMAIN"), false, true)
		return
	}
	c.SetCookie(cookieName, strings.Join(names, ","), 365*24*60*60, "/", os.Getenv("DOMAIN"), false, true)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db := DB{}
	db.InitializeGorm()

	formerNameService := &FormerNameService{db: db.db}
	vipListService := &VipListService{db: db.db}

	go CheckWorlds(db.db)
	go CheckFormerNames(db.db)

	r := gin.Default()

	r.LoadHTMLGlob("./templates/**/*")
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	r.GET("/vip-list", func(c *gin.Context) {
		c.HTML(200, "VipListTable.html", db.GetVipList(GetVipList(c)))
	})

	r.POST("/vip-list", func(c *gin.Context) {
		name := c.PostForm("name")
		vipListService.CreateVipListFriend(name)
		c.HTML(200, "VipListTable.html", db.GetVipList(AddVipList(c, name)))
	})

	r.DELETE("/vip-list/:name", func(c *gin.Context) {
		RemoveVipList(c, c.Params.ByName("name"))
		c.HTML(200, "VipListTable.html", db.GetVipList(RemoveVipList(c, c.Params.ByName("name"))))
	})

	r.GET("/former-names", func(c *gin.Context) {
		c.HTML(200, "FormerNamesTable.html", db.GetFormerNames(defaultNames))
	})

	r.POST("/former-names", func(c *gin.Context) {
		formerNameService.CreateFormerName(c.PostForm("name"))
		c.HTML(200, "FormerNamesTable.html", db.GetFormerNames(defaultNames))
	})

	r.DELETE("/former-names/:formerName", func(c *gin.Context) {
		formerNameService.DeleteFormerName(c.Params.ByName("formerName"))
		c.HTML(200, "FormerNamesTable.html", db.GetFormerNames(defaultNames))
	})

	r.Run(os.Getenv("SERVER_URL"))
}
