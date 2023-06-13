package tokens

import (
	"encoding/json"
	"os"
	"sync"
)

type AccessToken struct {
	tokens []string
	lock   sync.Mutex
}

type GPT4AccessToken struct {
	gpt4Tokens []string
	gpt4Lock   sync.Mutex
}

func NewAccessToken(tokens []string) AccessToken {
	// Save the tokens to a file
	if _, err := os.Stat("access_tokens.json"); os.IsNotExist(err) {
		// Create the file
		file, err := os.Create("access_tokens.json")
		if err != nil {
			return AccessToken{}
		}
		defer file.Close()
	}
	file, err := os.OpenFile("access_tokens.json", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return AccessToken{}
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(tokens)
	if err != nil {
		return AccessToken{}
	}
	return AccessToken{
		tokens: tokens,
	}
}

func NewGPT4AccessToken(gpt4Tokens []string) GPT4AccessToken {
	// Save the tokens to a file
	if _, err := os.Stat("gpt4_tokens.json"); os.IsNotExist(err) {
		// Create the file
		file, err := os.Create("gpt4_tokens.json")
		if err != nil {
			return GPT4AccessToken{}
		}
		defer file.Close()
	}
	file, err := os.OpenFile("gpt4_tokens.json", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return GPT4AccessToken{}
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(gpt4Tokens)
	if err != nil {
		return GPT4AccessToken{}
	}
	return GPT4AccessToken{
		gpt4Tokens: gpt4Tokens,
	}
}

func (a *AccessToken) GetToken() string {
	a.lock.Lock()
	defer a.lock.Unlock()

	if len(a.tokens) == 0 {
		return ""
	}

	token := a.tokens[0]
	a.tokens = append(a.tokens[1:], token)
	return token
}

func (a *GPT4AccessToken) GetGpt4Token() string {
	a.gpt4Lock.Lock()
	defer a.gpt4Lock.Unlock()

	if len(a.gpt4Tokens) == 0 {
		return ""
	}

	gpt4Token := a.gpt4Tokens[0]
	a.gpt4Tokens = append(a.gpt4Tokens[1:], gpt4Token)
	return gpt4Token
}
