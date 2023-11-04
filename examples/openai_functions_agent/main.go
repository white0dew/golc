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
	"github.com/hupe1980/golc/toolkit"
	"github.com/playwright-community/playwright-go"
	"github.com/sashabaranov/go-openai"
	"log"
)

func main() {
	golc.Verbose = false

	//if err := playwright.Install(); err != nil {
	//	log.Fatal(err)
	//}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatal(err)
	}

	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatal(err)
	}

	openai, err := chatmodel.NewOpenAI("sk-peisUrRs7gPLZKPk3c758475E6604f87B427Df3f4f34Cd45", func(o *chatmodel.OpenAIOptions) {
		o.ModelName = openai.GPT3Dot5Turbo16K0613
		o.BaseURL = "https://35.nekoapi.com/v1"
		o.Stream = true
		o.MaxTokens = 200
		o.CallbackOptions.Callbacks = []schema.Callback{
			callback.NewStreamWriterHandler(),
		}
	})
	if err != nil {
		log.Fatal(err)
	}

	browserKit, err := toolkit.NewBrowser(browser)
	if err != nil {
		log.Fatal(err)
	}

	// 历史的回答记录
	extraMessages := []prompt.MessageTemplate{
		prompt.NewSystemMessageTemplate("You are a world class algorithm for extracting information in structured formats."),
		prompt.NewHumanMessageTemplate("请用中文回答最后的答案"),
	}

	agent, err := agent.NewOpenAIFunctions(openai, browserKit.Tools(), func(o *agent.OpenAIFunctionsOptions) {
		o.MaxIterations = 10
		o.CallbackOptions.Callbacks = []schema.Callback{callback.NewStreamWriterHandler()}
		o.ExtraMessages = extraMessages
	})
	if err != nil {
		log.Fatal(err)
	}

	result, err := golc.SimpleCall(context.Background(), agent, "Navigate to https://yqsas.com/2019/03/21/how-to-develop-forked-go-project/ and summarize the text",
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
