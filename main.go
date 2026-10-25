package main

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	InitDB()
	InitTranslations()

	var router *gin.Engine

	if os.Getenv("USE_DEBUG") == "1" {
		router = gin.Default()
	} else {
		router = gin.New()
		router.Use(gin.Recovery())
	}

	_ = router.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	router.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})

	router.Static("/static", "./static")
	router.Static("/css", "./static/css")
	router.Static("/img", "./static/img")
	router.Static("/js", "./static/js")

	t := template.New("")
	err := filepath.Walk("views", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".html") {
			rel, err := filepath.Rel("views", path)
			if err != nil {
				return err
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			_, err = t.New(filepath.ToSlash(rel)).Parse(string(content))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error while loading templates: %v", err)
	}
	router.SetHTMLTemplate(t)

	router.Use(DirectoryMiddleware())
	router.Use(LangMiddleware())
	router.Use(PjaxMiddleware())
	router.Use(AuthMiddleware(DB))

	router.Use(func(c *gin.Context) {
		path := c.Request.URL.Path

		if path == "/welcome" || path == "/health" || strings.HasPrefix(path, "/static") || strings.HasPrefix(path, "/css") || strings.HasPrefix(path, "/img") || strings.HasPrefix(path, "/js") {
			c.Next()
			return
		}

		account, exists := c.Get("account")
		if !exists || account == nil {
			c.Redirect(302, "/welcome")
			c.Abort()
			return
		}

		c.Next()
	})

	router.GET("/welcome", func(c *gin.Context) {
		dir := c.GetString("directory")
		if dir == "" {
			dir = "portal"
		}
		lang, _ := c.Get("lang")
		account, _ := c.Get("account")

		c.HTML(200, dir+"/welcome.html", gin.H{
			"lang":    lang,
			"account": account,
		})
	})

	commRoutes := router.Group("/communities")
	{
		commRoutes.GET("", func(c *gin.Context) {
			dir := c.GetString("directory")
			c.HTML(200, dir+"/communities.html", gin.H{})
		})
	}

	router.GET("/messages", func(c *gin.Context) {
		dir := c.GetString("directory")
		c.HTML(200, dir+"/messages.html", gin.H{})
	})

	newsRoutes := router.Group("/news")
	{
		newsRoutes.GET("/my_news", func(c *gin.Context) {
			dir := c.GetString("directory")
			c.HTML(200, dir+"/news.html", gin.H{})
		})
		newsRoutes.GET("/friend_requests", func(c *gin.Context) {
			dir := c.GetString("directory")
			c.HTML(200, dir+"/friend_requests.html", gin.H{})
		})
	}

	router.GET("/show", func(c *gin.Context) {
		c.Redirect(302, "/titles")
	})

	router.GET("/users/:user_id", func(c *gin.Context) {
		dir := c.GetString("directory")
		c.HTML(200, dir+"/users.html", gin.H{})
	})

	activityRoutes := router.Group("/activity")
	{
		activityRoutes.GET("", func(c *gin.Context) {
			dir := c.GetString("directory")
			lang, _ := c.Get("lang")
			account, _ := c.Get("account")
			c.HTML(200, dir+"/activity.html", gin.H{
				"lang":    lang,
				"account": account,
			})
		})
		activityRoutes.GET("/welcome", func(c *gin.Context) {
			dir := c.GetString("directory")
			lang, _ := c.Get("lang")
			account, _ := c.Get("account")
			c.HTML(200, dir+"/welcome/welcome.html", gin.H{
				"lang":    lang,
				"account": account,
			})
		})
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "3001"
	}

	LogSuccess("Global Miiverse started on port " + port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
