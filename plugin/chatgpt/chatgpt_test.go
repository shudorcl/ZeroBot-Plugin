package chatgpt

import (
	"testing"
)

func TestGPT(t *testing.T) {
	messages := []chatMessage{}
	messages = append(messages, chatMessage{
		Role:    "user",
		Content: "你好！",
	})
	resp, err := completions(messages, "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")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp.Choices[0].Message)
}

//  curl --location 'https://open.aiproxy.xyz/dashboard/billing/credit_grants' --header 'Authorization: Bearer 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'
