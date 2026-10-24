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

	router := gin.Default()

	_ = router.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	router.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})

	router.Static("/static", "./static")
	router.Static("/css", "./static/css")

	// Chargement récursif et exact de tous les templates .html sous views/
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
		log.Fatalf("Erreur lors du chargement des templates: %v", err)
	}
	router.SetHTMLTemplate(t)

	// Middlewares globaux
	router.Use(DirectoryMiddleware())
	router.Use(LangMiddleware())
	router.Use(PjaxMiddleware())

	// Middleware de redirection si non connecté
	router.Use(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Pages publiques autorisées sans connexion
		if path == "/welcome" || path == "/health" || strings.HasPrefix(path, "/static") || strings.HasPrefix(path, "/css") {
			c.Next()
			return
		}

		// Vérifie si le compte existe dans le contexte
		account, exists := c.Get("account")
		if !exists || account == nil {
			c.Redirect(302, "/welcome")
			c.Abort()
			return
		}

		c.Next()
	})

	// Route spécifique pour /welcome
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
			communities, err := GetAllCommunities()
			if err != nil {
				c.HTML(500, dir+"/error.html", gin.H{"code": 500, "message": "Error: " + err.Error()})
				return
			}
			c.HTML(200, dir+"/communities.html", gin.H{"communities": communities})
		})

		commRoutes.GET("/favorites", func(c *gin.Context) {
			dir := c.GetString("directory")
			c.HTML(200, dir+"/fav_communities.html", nil)
		})

		commRoutes.GET("/:olive_title_id/:olive_community_id", func(c *gin.Context) {
			dir := c.GetString("directory")
			oliveCommunityId := c.Param("olive_community_id")

			community, err := GetCommunityById(oliveCommunityId)
			if err != nil || community == nil {
				c.HTML(404, dir+"/error.html", gin.H{"code": 404, "message": "Community not found"})
				return
			}
			c.HTML(200, dir+"/page_community.html", gin.H{"community": community})
		})
	}

	router.GET("/messages", func(c *gin.Context) {
		dir := c.GetString("directory")
		c.HTML(200, dir+"/messages.html", nil)
	})

	newsRoutes := router.Group("/news")
	{
		newsRoutes.GET("/my_news", func(c *gin.Context) {
			dir := c.GetString("directory")
			c.HTML(200, dir+"/news.html", nil)
		})
		newsRoutes.GET("/friend_requests", func(c *gin.Context) {
			dir := c.GetString("directory")
			c.HTML(200, dir+"/friend_requests.html", nil)
		})
	}

	router.GET("/show", func(c *gin.Context) {
		c.Redirect(302, "/titles")
	})

	router.GET("/users/:user_id", func(c *gin.Context) {
		dir := c.GetString("directory")
		userId := c.Param("user_id")

		user, err := GetUserById(userId)
		if err != nil || user == nil {
			c.HTML(404, dir+"/error.html", gin.H{"code": 404, "message": "User not found"})
			return
		}
		c.HTML(200, dir+"/users.html", gin.H{"user": user, "own_profile": false})
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
		port = "3000"
	}

	LogSuccess("Global Miiverse started on port " + port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
