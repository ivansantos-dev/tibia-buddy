package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"html/template"
	"io"
	"net/http"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type IndexData struct {
	VipList []VipCharacter
}

type VipCharacter struct {
	Name   string
	World  string
	Status CharacterStatus
}

type CharacterStatus string

const (
	Online  CharacterStatus = "online"
	Offline CharacterStatus = "offline"
)

type SearchData struct {
	Name        string
	World       string
	Status      CharacterStatus
	Level       int
	FormerNames []string
	Error       string
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)
	if l, ok := e.Logger.(*log.Logger); ok {
		l.SetHeader("${time_rfc3339} ${level}")
	}

	t := &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Renderer = t

	e.Static("/static", "static")

	e.GET("/", func(c echo.Context) error {
		data := IndexData{
			VipList: []VipCharacter{
				{Name: "Luffy", World: "East Blue", Status: Online},
				{Name: "Law", World: "West Blue", Status: Offline},
			},
		}

		return c.Render(http.StatusOK, "index", data)
	})

	e.POST("/search", func(c echo.Context) error {
		name := c.FormValue("search_name")
		character, err := GetCharacter(name)

		data := &SearchData{}
		if err != nil {
			data.Error = err.Error()
			return c.Render(http.StatusOK, "search-table", data)
		}

		data.Name = character.CharacterInfo.Name
		data.World = character.CharacterInfo.World
		data.Level = character.CharacterInfo.Level
		data.FormerNames = character.CharacterInfo.FormerNames

		return c.Render(http.StatusOK, "search-table", data)
	})

	e.PUT("/vip-list/:name", func(c echo.Context) error {
		name := c.Param("name")
		e.Logger.Info(name)

		return c.HTML(http.StatusOK, "<button class=\"secondary\" disabled>Added</button>")
	})

	e.Logger.Fatal(e.Start("127.0.0.1:1323"))
}
