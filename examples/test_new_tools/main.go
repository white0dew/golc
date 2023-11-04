package main

import (
	"context"
	"fmt"
	"github.com/hupe1980/golc"
	"github.com/hupe1980/golc/agent"
	"github.com/hupe1980/golc/callback"
	"github.com/hupe1980/golc/model/chatmodel"
	"github.com/hupe1980/golc/prompt"
	"github.com/hupe1980/golc/schema"
	"github.com/hupe1980/golc/tool"
	"github.com/sashabaranov/go-openai"
	"log"
	"os"
)

func main() {
	golc.Verbose = true

	openai, err := chatmodel.NewOpenAI(os.Getenv("openai_key"), func(o *chatmodel.OpenAIOptions) {
		o.ModelName = openai.GPT3Dot5Turbo16K0613
		o.BaseURL = "https://35.pixelmoe.com/v1"
		o.Stream = true
		o.MaxTokens = 2000
		o.CallbackOptions.Callbacks = []schema.Callback{
			callback.NewStreamWriterHandler(),
		}
	})
	if err != nil {
		log.Fatal(err)
	}

	// 历史的回答记录
	extraMessages := []prompt.MessageTemplate{
		prompt.NewHumanMessageTemplate("请用中文回答最后的答案"),
	}

	agent, err := agent.NewOpenAIFunctions(openai, []schema.Tool{
		tool.NewBaiduHot(),
	}, func(o *agent.OpenAIFunctionsOptions) {
		o.CallbackOptions.Callbacks = []schema.Callback{callback.NewStreamWriterHandler()}
		o.ExtraMessages = extraMessages
	})
	if err != nil {
		log.Fatal(err)
	}

	result, err := golc.SimpleCall(context.Background(), agent, "现在的百度热榜是什么?简单总结一下",
		func(options *golc.SimpleCallOptions) {
			options.Callbacks = []schema.Callback{
				callback.NewStreamWriterHandler(),
			}
		})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}

//Navigate to https://yqsas.com/2019/03/21/how-to-develop-forked-go-project/ and summarize the text
