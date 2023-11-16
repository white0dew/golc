package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hupe1980/golc/model/chatmodel"
	"github.com/hupe1980/golc/schema"
	"github.com/sashabaranov/go-openai"
	"log"
)

type Item struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

type Data struct {
	Content []Item `json:"content"`
}

func main() {
	openai, err := chatmodel.NewOpenAI("sk-6TQ9GKHl4pmGW024E9Fc582aF449407183B3Cb0aB2360e2c", func(o *chatmodel.OpenAIOptions) {
		o.Stream = true
		o.ModelName = openai.GPT4VisionPreview
		o.BaseURL = "https://api.rcouyi.com/v1"
	})
	if err != nil {
		log.Fatal(err)
	}

	content, err := json.Marshal([]Item{
		{
			Type:     "text",
			Text:     "这个图片内容是什么",
			ImageURL: "",
		},
		{
			Type:     "image_url",
			ImageURL: "https://lmg.jj20.com/up/allimg/tp05/1Z9291T012CB-0-lp.jpg",
		},
	})
	res, _ := openai.Generate(context.Background(), schema.ChatMessages{
		schema.NewHumanChatMessage(
			string(content)),
	}, func(o *schema.GenerateOptions) {
	})

	//res, mErr := model.GeneratePrompt(context.Background(), openai, prompt.StringPromptValue("give me a pic of beauty"), func(o *model.Options) {
	//	o.Callbacks = []schema.Callback{callback.NewStreamWriterHandler()}
	//})
	//if mErr != nil {
	//	log.Fatal(mErr)
	//}
	fmt.Printf("res:%v", res)
}
