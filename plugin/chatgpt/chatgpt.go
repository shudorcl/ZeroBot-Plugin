package chatgpt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	// baseURL  = "https://api.openai.com/v1/"
	proxyURL          = "https://openai.geekr.cool/v1/"
	GPT3Dot5Turbo0301 = "gpt-3.5-turbo-0301"
	GPT3Dot5Turbo     = "gpt-3.5-turbo"
)

// chatGPTResponseBody 响应体
type chatGPTResponseBody struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int          `json:"created"`
	Model   string       `json:"model"`
	Choices []chatChoice `json:"choices"`
	Usage   chatUsage    `json:"usage"`
}

// chatGPTRequestBody 请求体
type chatGPTRequestBody struct {
	Model            string        `json:"model"`
	Messages         []chatMessage `json:"messages"`
	MaxTokens        int           `json:"max_tokens"`
	Temperature      float32       `json:"temperature"`
	TopP             int           `json:"top_p"`
	FrequencyPenalty int           `json:"frequency_penalty"`
	PresencePenalty  int           `json:"presence_penalty"`
}

// chatMessage 消息
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatChoice struct {
	Index        int `json:"index"`
	Message      chatMessage
	FinishReason string `json:"finish_reason"`
}

type chatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

var client = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	},
	Timeout: 5 * time.Minute,
}

// completions gtp3.5文本模型回复
// curl https://api.openai.com/v1/chat/completions
// -H "Content-Type: application/json"
// -H "Authorization: Bearer YOUR_API_KEY"
// -d '{ "model": "gpt-3.5-turbo",  "messages": [{"role": "user", "content": "Hello!"}]}'
func completions(messages []chatMessage, apiKey string) (*chatGPTResponseBody, error) {
	com := chatGPTRequestBody{
		Messages: messages,
	}
	// default model
	if com.Model == "" {
		com.Model = GPT3Dot5Turbo
	}

	body, err := json.Marshal(com)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, proxyURL+"chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		// TODO: introduce typed error
		return nil, errors.New("response error")
	}

	v := new(chatGPTResponseBody)
	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}
