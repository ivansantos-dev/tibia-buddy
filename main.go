package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

func filterEmpty(slice []string) []string {
	var result []string
	for _, str := range slice {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}
func GetFormerNameList(c *gin.Context) []string {
	listStr, err := c.Cookie("former-names")
	if err != nil {
		log.Error("Failed to get cookie", err)
	}
	return filterEmpty(strings.Split(listStr, ","))
}

func AddFormerNameList(c *gin.Context, name string) []string {
	list := GetFormerNameList(c)
	list = addName(list, name)
	SetFormerNameCookie(c, list)
	return list
}

func RemoveFormerNameList(c *gin.Context, name string) []string {
	list := GetFormerNameList(c)
	list = removeName(list, name)
	SetFormerNameCookie(c, list)
	return list
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
	list = addName(list, name)
	SetVipListCookie(c, list)
	return list
}

func addName(list []string, name string) []string {
	for _, n := range list {
		if n == name {
			return list
		}
	}
	return append(list, name)
}

func RemoveVipList(c *gin.Context, name string) []string {
	list := GetVipList(c)
	list = removeName(list, name)
	SetVipListCookie(c, list)
	return list
}

func removeName(list []string, name string) []string {
	for i, n := range list {
		if n == name {
			if len(list) == 1 {
				list = []string{}
				break
			}
			list = append(list[:i], list[i+1:]...)
		}
	}
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
		c.HTML(200, "VipListTable.html", gin.H{
			"VipList": db.GetVipList(GetVipList(c)),
		})
	})

	r.PUT("/vip-list/:name", func(c *gin.Context) {
		name := c.Params.ByName("name")
		err := vipListService.CreateVipListFriend(name)
		if err != nil {
		}

		c.HTML(200, "VipListTable.html", gin.H{
			"VipList": db.GetVipList(AddVipList(c, c.Params.ByName("name"))),
		})
	})

	r.POST("/search-vip-list", func(c *gin.Context) {
		name := c.PostForm("name")
		apiChar, err := GetCharacter(name)
		if err != nil {
			log.Error(err)
		}

		c.HTML(200, "SearchVipListModal.html", gin.H{"Char": apiChar.CharacterInfo, "SafeCharName": apiChar.CharacterInfo.Name})
	})

	r.DELETE("/vip-list/:name", func(c *gin.Context) {
		c.HTML(200, "VipListTable.html", gin.H{
			"VipList": db.GetVipList(RemoveVipList(c, c.Params.ByName("name"))),
		})
	})

	r.GET("/former-names", func(c *gin.Context) {
		c.HTML(200, "FormerNamesTable.html", gin.H{"FormerNames": db.GetFormerNames(GetFormerNameList(c))})
	})

	r.POST("/former-names", func(c *gin.Context) {
		name := c.PostForm("name")
		err := formerNameService.CreateFormerName(name)
		var errorText string
		if err != nil {
			errorText = err.Error()
		}

		c.HTML(200, "FormerNamesTable.html", gin.H{"Error": errorText, "FormerNames": db.GetFormerNames(AddFormerNameList(c, name))})
	})

	r.DELETE("/former-names/:formerName", func(c *gin.Context) {
		name := c.Params.ByName("formerName")
		c.HTML(200, "FormerNamesTable.html", gin.H{
			"FormerNames": db.GetFormerNames(RemoveFormerNameList(c, name)),
		})
	})

	r.Run(os.Getenv("SERVER_URL"))
}
