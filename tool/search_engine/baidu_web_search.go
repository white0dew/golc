package main

import (
	"encoding/json"
	"fmt"
	"github.com/hupe1980/golc/tool/search_engine/baidu"
	"github.com/hupe1980/golc/tool/search_engine/core"
	"time"
)

func main() {
	query := core.Query{
		Text:  "哈哈",
		Limit: 10,
	}
	var engine core.SearchEngine

	opts := core.BrowserOpts{
		IsHeadless:    true, // Disable headless if browser head mode is set
		IsLeakless:    false,
		Timeout:       time.Second * time.Duration(15),
		LeavePageOpen: false,
	}

	browser, err := core.NewBrowser(opts)
	if err != nil {
		fmt.Errorf("err:%v", err)
		return
	}
	engine = baidu.New(*browser, core.SearchEngineOptions{
		RateRequests: 4,
		RateBurst:    2,
	})

	res, err := engine.Search(query)
	resR, _ := json.Marshal(res)

	fmt.Printf("res:%v,%v", string(resR), err)
	fmt.Println("123")
	return
}
