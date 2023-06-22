package main

import (
	"ChatGPT-to-API/internal/chatgpt"
	"ChatGPT-to-API/internal/tokens"
	typings "ChatGPT-to-API/internal/typings"
	"ChatGPT-to-API/internal/typings/responses"
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func openaiHandler(c *gin.Context) {
	err := c.BindJSON(&authorizations)
	if err != nil {
		c.JSON(400, gin.H{"error": "JSON invalid"})
	}
	os.Setenv("OPENAI_EMAIL", authorizations.OpenAI_Email)
	os.Setenv("OPENAI_PASSWORD", authorizations.OpenAI_Password)
	c.String(200, "OpenAI credentials updated")
}

func passwordHandler(c *gin.Context) {
	// Get the password from the request (json) and update the password
	type password_struct struct {
		Password string `json:"password"`
	}
	var password password_struct
	err := c.BindJSON(&password)
	if err != nil {
		c.String(400, "password not provided")
		return
	}
	ADMIN_PASSWORD = password.Password
	// Set environment variable
	os.Setenv("ADMIN_PASSWORD", ADMIN_PASSWORD)
	c.String(200, "password updated")
}

func getPasswordHandler(c *gin.Context) {
	if ADMIN_PASSWORD == "" {
		c.String(200, "password is empty")
	}
	c.String(200, "password is: "+ADMIN_PASSWORD)
}

func puidHandler(c *gin.Context) {
	// Get the password from the request (json) and update the password
	type puid_struct struct {
		PUID string `json:"puid"`
	}
	var puid puid_struct
	err := c.BindJSON(&puid)
	if err != nil {
		c.String(400, "puid not provided")
		return
	}
	// Set environment variable
	os.Setenv("PUID", puid.PUID)
	c.String(200, "puid updated")
}

func tokensHandler(c *gin.Context) {
	// Get the request_tokens from the request (json) and update the request_tokens
	var requestTokens []string
	err := c.BindJSON(&requestTokens)
	if err != nil {
		c.String(400, "gpt3.5 tokens not provided")
		return
	}
	ACCESS_TOKENS = tokens.NewAccessToken(requestTokens)
	c.String(200, "gpt3.5 tokens updated")
}

func gpt4TokensHandler(c *gin.Context) {
	// Get the request_tokens from the request (json) and update the request_tokens
	var requestTokens []string
	err := c.BindJSON(&requestTokens)
	if err != nil {
		c.String(400, "gpt4 tokens not provided")
		return
	}
	GPT4_TOKENS = tokens.NewGPT4AccessToken(requestTokens)
	c.String(200, "gpt4 tokens updated")
}

func getTokensHandler(c *gin.Context) {
	// Get the request_tokens from the request (json) and update the request_tokens
	type tokenStruct struct {
		Model string `json:"model"`
	}
	var modelParam tokenStruct
	var res []string
	err := c.BindJSON(&modelParam)
	if err != nil {
		c.String(400, "request param is invalid")
		return
	}
	if modelParam.Model == "gpt-3.5-turbo" {
		res = ACCESS_TOKENS.GetPublicGpt3Tokens()
		c.JSON(200, gin.H{
			"data": res,
		})
	} else if modelParam.Model == "gpt-4" {
		res = GPT4_TOKENS.GetPublicGpt4Tokens()
		c.JSON(200, gin.H{
			"data": res,
		})
	} else {
		c.String(400, "no valid model tokens")
	}
}

func optionsHandler(c *gin.Context) {
	// Set headers for CORS
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST")
	c.Header("Access-Control-Allow-Headers", "*")
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
func nightmare(c *gin.Context) {
	var original_request typings.APIRequest
	err := c.BindJSON(&original_request)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{
			"message": "Request must be proper JSON",
			"type":    "invalid_request_error",
			"param":   nil,
			"code":    err.Error(),
		}})
	}
	// Convert the chat request to a ChatGPT request
	translated_request := chatgpt.ConvertAPIRequest(original_request)

	authHeader := c.GetHeader("Authorization")

	var token string
	if strings.HasPrefix(original_request.Model, "gpt-4") {
		token = GPT4_TOKENS.GetGpt4Token()
		log.Println("now token is gpt-4: ", token)
	} else {
		token = ACCESS_TOKENS.GetToken()
		log.Println("now token is gpt-3.5: ", token)
	}
	log.Printf("origin request: %+v \n", translated_request)
	if authHeader != "" {
		customAccessToken := strings.Replace(authHeader, "Bearer ", "", 1)
		// Check if customAccessToken starts with sk-
		if strings.HasPrefix(customAccessToken, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik1UaEVOVUpHTkVNMVFURTRNMEZCTWpkQ05UZzVNRFUxUlRVd1FVSkRNRU13UmtGRVFrRXpSZyJ9") {
			token = customAccessToken
		}
	}

	response, err := chatgpt.SendRequest(translated_request, token)
	if err != nil {
		c.JSON(response.StatusCode, gin.H{
			"error":   "error sending request",
			"message": response.Status,
		})
		return
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		// Try read response body as JSON
		var error_response map[string]interface{}
		err = json.NewDecoder(response.Body).Decode(&error_response)
		if err != nil {
			// Read response body
			body, _ := io.ReadAll(response.Body)
			c.JSON(500, gin.H{"error": gin.H{
				"message": "Unknown error",
				"type":    "internal_server_error",
				"param":   nil,
				"code":    "500",
				"details": string(body),
			}})
			return
		}
		c.JSON(response.StatusCode, gin.H{"error": gin.H{
			"message": error_response["detail"],
			"type":    response.Status,
			"param":   nil,
			"code":    "error",
		}})
		return
	}
	// Create a bufio.Reader from the response body
	reader := bufio.NewReader(response.Body)

	var fulltext string

	var conversationID string

	// Read the response byte by byte until a newline character is encountered
	if original_request.Stream {
		// Response content type is text/event-stream
		c.Header("Content-Type", "text/event-stream")
	} else {
		// Response content type is application/json
		c.Header("Content-Type", "application/json")
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		if len(line) < 6 {
			continue
		}
		// Remove "data: " from the beginning of the line
		line = line[6:]
		// Check if line starts with [DONE]
		if !strings.HasPrefix(line, "[DONE]") {
			// Parse the line as JSON
			var original_response responses.Data
			err = json.Unmarshal([]byte(line), &original_response)

			conversationID = original_response.ConversationID

			if err != nil {
				continue
			}
			if original_response.Error != nil {
				return
			}
			if original_response.Message.Content.Parts == nil {
				continue
			}
			if original_response.Message.Content.Parts[0] == "" || original_response.Message.Author.Role != "assistant" {
				continue
			}
			if original_response.Message.Metadata.Timestamp == "absolute" {
				continue
			}
			tmp_fulltext := original_response.Message.Content.Parts[0]
			original_response.Message.Content.Parts[0] = strings.ReplaceAll(original_response.Message.Content.Parts[0], fulltext, "")
			translated_response := responses.NewChatCompletionChunk(original_response.Message.Content.Parts[0], original_request.Model)

			// Stream the response to the client
			response_string := translated_response.String()
			if original_request.Stream {
				_, err = c.Writer.WriteString("data: " + string(response_string) + "\n\n")
				if err != nil {
					return
				}
			}

			// Flush the response writer buffer to ensure that the client receives each line as it's written
			c.Writer.Flush()
			fulltext = tmp_fulltext
		} else {
			// 将chatgpt账户中的对话内容删除
			url := "http://sj1.nonezero.top:4141/chatgpt/conversation/" + conversationID

			log.Println("message url : ", url)

			payload := strings.NewReader(`{
    				"is_visible": false
			}`)

			client := &http.Client{}
			req, err := http.NewRequest("PATCH", url, payload)

			if err != nil {
				log.Println(err)
			}
			req.Header.Add("Authorization", token)
			req.Header.Add("Content-Type", "application/json")

			res, err := client.Do(req)
			if err != nil {
				log.Println(err)
			}
			defer res.Body.Close()

			if !original_request.Stream {
				full_response := responses.NewChatCompletion(fulltext, original_request.Model)
				if err != nil {
					return
				}
				c.JSON(200, full_response)
				return
			}
			final_line := responses.StopChunk(original_request.Model)
			c.Writer.WriteString("data: " + final_line.String() + "\n\n")

			c.String(200, "data: [DONE]\n\n")
			return

		}
	}

}
