package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
)

func buildPayload(msg string) openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: msg,
			},
		},
		Stream: true,
	}
}

func buildReq(ctx context.Context, key string, req openai.ChatCompletionRequest) (chan *openai.ChatCompletionStreamResponse, error) {
	client := openai.NewClient(key)
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}
	respCh := make(chan *openai.ChatCompletionStreamResponse, 1)
	go func() {
		defer close(respCh)
		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				fmt.Printf("chatgpt 流式异常中断, err: %v\n", err)
				return
			}
			respCh <- &resp
		}
	}()
	return respCh, nil
}

func main() {
	ctx := context.Background()
	cgReq := buildPayload(msg)
	respCh, err := buildReq(ctx, key, cgReq)
	if err != nil {
		fmt.Printf("请求openai 错误 error: %v", err)
	}
	for resp := range respCh {
		if len(resp.Choices) > 0 {
			fmt.Print(resp.Choices[0].Delta.Content)
		}
	}
}

var (
	msg string
	key string
)

func init() {
	flag.StringVar(&msg, "t", "讲个故事吧", "msg")
	flag.StringVar(&key, "k", "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", "key")
}
