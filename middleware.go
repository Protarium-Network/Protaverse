package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Translations = make(map[string]map[string]interface{})
var DefaultLang = "en"

type contextKey string

const (
	ContextAccount      contextKey = "account"
	ContextPlatform     contextKey = "platform"
	ContextParamPack    contextKey = "param_pack"
	ContextServiceToken contextKey = "service_token"
)

type Account struct {
	ID               int
	ServiceToken3DS  sql.NullString
	ServiceTokenWiiU sql.NullString
}

func AuthMiddleware(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		paramPackHeader := c.GetHeader("X-Nintendo-Parampack")
		serviceToken := c.GetHeader("X-Nintendo-Servicetoken")

		if serviceToken == "" || paramPackHeader == "" {
			c.String(http.StatusUnauthorized, "service_token and param_pack are missing.")
			c.Abort()
			return
		}

		serviceToken = strings.NewReplacer("\r", "", "\n", "").Replace(strings.TrimSpace(serviceToken))
		paramPackHeader = strings.NewReplacer("\r", "", "\n", "", " ", "", "\t", "").Replace(paramPackHeader)

		decodedBytes, err := base64.StdEncoding.DecodeString(paramPackHeader)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid ParamPack")
			c.Abort()
			return
		}

		decodedStr := string(decodedBytes)
		if len(decodedStr) < 2 {
			c.String(http.StatusBadRequest, "Invalid ParamPack")
			c.Abort()
			return
		}

		trimmed := decodedStr[1 : len(decodedStr)-1]
		base64Result := strings.Split(trimmed, "\\")

		paramPack := make(map[string]string)
		for i := 0; i < len(base64Result); i += 2 {
			if i+1 < len(base64Result) {
				paramPack[strings.TrimSpace(base64Result[i])] = strings.TrimSpace(base64Result[i+1])
			}
		}

		platformIDStr, exists := paramPack["platform_id"]
		if !exists {
			c.String(http.StatusBadRequest, "Invalid ParamPack")
			c.Abort()
			return
		}

		platform, err := strconv.Atoi(platformIDStr)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid ParamPack")
			c.Abort()
			return
		}

		var query string
		var platformName string

		if platform == 0 {
			platformName = "3ds"
			query = "SELECT * FROM accounts WHERE service_token_3ds = $1 LIMIT 1"
		} else {
			platformName = "wiiu"
			query = "SELECT * FROM accounts WHERE service_token_wiiu = $1 LIMIT 1"
		}

		rows, err := db.Query(c.Request.Context(), query, serviceToken)
		if err != nil {
			LogError("Error database (Query): %v", err)
			c.Status(http.StatusInternalServerError)
			c.Abort()
			return
		}

		account, err := pgx.CollectOneRow(rows, pgx.RowToMap)
		if err != nil {
			if err == pgx.ErrNoRows {
				if c.Request.URL.Path == "/welcome" {
					c.Next()
					return
				}
				c.Redirect(http.StatusFound, "/welcome")
				c.Abort()
				return
			}
			LogError("Error database (Scan): %v", err)
			c.Status(http.StatusInternalServerError)
			c.Abort()
			return
		}

		c.Set("account", account)
		c.Set("platform", platformName)
		c.Set("param_pack", paramPack)
		c.Set("service_token", serviceToken)

		c.Next()
	}
}

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
		for k := range Translations {
			DefaultLang = k
			break
		}
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
