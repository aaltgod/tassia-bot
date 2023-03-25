package chat_gpt

import (
	"context"
	"errors"
	"strings"

	chatgpt "github.com/sashabaranov/go-openai"
)

type Client struct {
	token           string
	VerifiedUserIDs []int

	chatGPT *chatgpt.Client
}

func New(token string, verifiedUserIDs []int) *Client {
	client := &Client{
		token:           token,
		VerifiedUserIDs: verifiedUserIDs,
	}

	client.chatGPT = chatgpt.NewClient(client.token)

	return client
}

func (c *Client) SendMessage(message string) (string, error) {
	resp, err := c.chatGPT.CreateChatCompletion(context.Background(), chatgpt.ChatCompletionRequest{
		Model: chatgpt.GPT3Dot5Turbo,
		Messages: []chatgpt.ChatCompletionMessage{
			{
				Role:    chatgpt.ChatMessageRoleUser,
				Content: message,
			},
		},
	})
	if err != nil {
		return "", errors.New("CreateChatCompletion " + err.Error())
	}

	var result strings.Builder

	for _, v := range resp.Choices {
		result.WriteString(v.Message.Content)
	}

	return result.String(), nil
}
