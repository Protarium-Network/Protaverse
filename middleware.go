package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var Translations = make(map[string]map[string]interface{})

const DefaultLang = "en"

func InitTranslations() {
	langDir := "./lang"
	files, err := os.ReadDir(langDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}
		code := strings.TrimSuffix(file.Name(), ".json")
		content, err := os.ReadFile(filepath.Join(langDir, file.Name()))
		if err != nil {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal(content, &data); err == nil {
			Translations[code] = data
		}
	}

	if _, exists := Translations[DefaultLang]; !exists && len(Translations) > 0 {

	}
}

func LangMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptLang := c.GetHeader("Accept-Language")
		lang := DefaultLang

		if len(acceptLang) >= 2 {
			prefix := strings.ToLower(acceptLang[:2])
			if _, exists := Translations[prefix]; exists {
				lang = prefix
			}
		}

		c.Set("lang", Translations[lang])
		c.Set("local", lang)
		c.Next()
	}
}

func DirectoryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("directory", "portal")
		c.Next()
	}
}

func PjaxMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isPjax := c.Query("_pjax") != "" && c.GetHeader("X-PJAX") != ""
		c.Set("isPjax", isPjax)

		if isPjax {
			cleanUrl := strings.Split(c.Request.RequestURI, "?")[0]
			c.Header("X-PJAX-PATH", cleanUrl)
			c.Header("X-PJAX-OK", "1")
		}
		c.Next()
	}
}
