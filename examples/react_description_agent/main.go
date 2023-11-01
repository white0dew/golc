package main

import (
	"context"
	"fmt"
	"github.com/hupe1980/golc"
	"github.com/hupe1980/golc/agent"
	"github.com/hupe1980/golc/callback"
	"github.com/hupe1980/golc/model/llm"
	"github.com/hupe1980/golc/schema"
	"github.com/hupe1980/golc/toolkit"
	"github.com/playwright-community/playwright-go"
	"log"
)

func main() {
	golc.Verbose = true

	if err := playwright.Install(); err != nil {
		log.Fatal(err)
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatal(err)
	}

	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatal(err)
	}

	openai, err := llm.NewOpenAI("sk-peisUrRs7gPLZKPk3c758475E6604f87B427Df3f4f34Cd45",
		func(o *llm.OpenAIOptions) {
			o.BaseURL = "https://35.nekoapi.com/v1"
			o.Stream = true
			o.CallbackOptions = &schema.CallbackOptions{
				Callbacks: []schema.Callback{callback.NewStreamWriterHandler()},
			}
		})
	if err != nil {
		log.Fatal(err)
	}

	browserKit, err := toolkit.NewBrowser(browser)
	if err != nil {
		log.Fatal(err)
	}

	agent, err := agent.NewReactDescription(openai, browserKit.Tools())
	if err != nil {
		log.Fatal(err)
	}

	result, err := golc.SimpleCall(context.Background(), agent, "Navigate to https://news.ycombinator.com and summarize the text")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}
