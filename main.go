package main

import (
	"github.com/gorilla/sessions"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"html/template"
	"io"
	"net/http"
	"os"
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

	e.HideBanner = true
	e.Logger.SetLevel(log.INFO)

	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	store.MaxAge(60 * 60 * 24 * 30)
	store.Options.Path = "/"
	store.Options.HttpOnly = true // HttpOnly should always be enabled
	store.Options.Secure = false  // TODO set in env
	gothic.Store = store

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(store))

	clientId := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
	clientCallback := os.Getenv("GOOGLE_OAUTH_CALLBACK_URL")
	e.Logger.Info(clientId, clientSecret, clientCallback)
	goth.UseProviders(google.New(
		clientId,
		clientSecret,
		clientCallback,
	))

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

	e.GET("/auth/:provider", func(c echo.Context) error {
		providerName := c.Param("provider")
		provider, err := goth.GetProvider(providerName)
		req := c.Request()
		res := c.Response()
		if err != nil {
			e.Logger.Fatal(err)
		}
		sess, err := provider.BeginAuth(gothic.SetState(req))
		if err != nil {
			e.Logger.Fatal(err)
		}

		authUrl, err := sess.GetAuthURL()
		if err != nil {
			e.Logger.Fatal(err)
		}

		err = gothic.StoreInSession(providerName, sess.Marshal(), req, res)
		if err != nil {
			e.Logger.Fatal(err)
		}

		return c.Redirect(http.StatusTemporaryRedirect, authUrl)
	})
	e.GET("/auth/:provider/callback", func(c echo.Context) error {
		user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
		if err != nil {
			e.Logger.Fatal(err)
		}

		e.Logger.Info("logged in as", user)
		return c.Redirect(http.StatusFound, "/")
	})

	e.GET("/logout/:provider", func(c echo.Context) error {
		err := gothic.Logout(c.Response(), c.Request())
		if err != nil {
			e.Logger.Fatal(err)
		}

		return c.Redirect(http.StatusFound, "/")
	})

	e.Logger.Fatal(e.Start("127.0.0.1:8090"))
}
