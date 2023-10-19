package main

import (
	"github.com/labstack/echo/v4"
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

func main() {
	t := &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}

	e := echo.New()
	e.Renderer = t
	e.Static("/static", "static")

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", "World!")
	})

	e.POST("/search", func(c echo.Context) error {
		name := c.FormValue("search_name")

		return c.Render(http.StatusOK, "search-table", name)
	})

	e.PUT("/vip-list/:name", func(c echo.Context) error {
		name := c.Param("name")
		e.Logger.Info(name)

		return c.Render(http.StatusOK, "index", nil)
	})

	e.Logger.Fatal(e.Start("127.0.0.1:1323"))
}
