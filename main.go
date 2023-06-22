package main

import (
	"ChatGPT-to-API/internal/tokens"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/acheong08/OpenAIAuth/auth"
	"github.com/acheong08/endless"
	"github.com/gin-gonic/gin"
)

var HOST string
var PORT string
var ACCESS_TOKENS tokens.AccessToken
var GPT4_TOKENS tokens.GPT4AccessToken

var authorizations struct {
	OpenAI_Email    string `json:"openai_email"`
	OpenAI_Password string `json:"openai_password"`
}

func init() {
	authorizations.OpenAI_Email = os.Getenv("OPENAI_EMAIL")
	authorizations.OpenAI_Password = os.Getenv("OPENAI_PASSWORD")
	if authorizations.OpenAI_Email != "" && authorizations.OpenAI_Password != "" {
		go func() {
			for {
				authenticator := auth.NewAuthenticator(authorizations.OpenAI_Email, authorizations.OpenAI_Password, os.Getenv("http_proxy"))
				err := authenticator.Begin()
				if err != nil {
					log.Println(err)
					break
				}
				puid, err := authenticator.GetPUID()
				if err != nil {
					break
				}
				os.Setenv("PUID", puid)
				println(puid)
				time.Sleep(24 * time.Hour * 7)
			}
		}()
	}
	HOST = os.Getenv("SERVER_HOST")
	PORT = os.Getenv("SERVER_PORT")
	if HOST == "" {
		HOST = "127.0.0.1"
	}
	if PORT == "" {
		PORT = "4242"
	}
	accessToken := os.Getenv("ACCESS_TOKENS")
	if accessToken != "" {
		accessTokens := strings.Split(accessToken, ",")
		ACCESS_TOKENS = tokens.NewAccessToken(accessTokens)
	}
	// Check if access_tokens.json exists
	if _, err := os.Stat("access_tokens.json"); os.IsNotExist(err) {
		// Create the file
		file, err := os.Create("access_tokens.json")
		if err != nil {
			panic(err)
		}
		defer file.Close()
	} else {
		// Load the tokens
		file, err := os.Open("access_tokens.json")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		decoder := json.NewDecoder(file)
		var token_list []string
		err = decoder.Decode(&token_list)
		if err != nil {
			return
		}
		ACCESS_TOKENS = tokens.NewAccessToken(token_list)
	}
}
func main() {
	router := gin.Default()

	router.Use(cors)

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	admin_routes := router.Group("/admin")
	admin_routes.Use(adminCheck)

	/// Admin routes
	admin_routes.PATCH("/password", passwordHandler)
	admin_routes.GET("/password", getPasswordHandler)
	admin_routes.PATCH("/tokens", tokensHandler)
	admin_routes.PATCH("/gpt4tokens", gpt4TokensHandler)
	admin_routes.GET("/gettokens", getTokensHandler)
	admin_routes.PATCH("/puid", puidHandler)
	admin_routes.PATCH("/openai", openaiHandler)
	/// Public routes
	router.OPTIONS("/v1/chat/completions", optionsHandler)
	router.POST("/v1/chat/completions", Authorization, nightmare)
	endless.ListenAndServe(HOST+":"+PORT, router)
}
