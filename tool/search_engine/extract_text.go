package main

import (
	"fmt"
	"github.com/hupe1980/golc/tool/search_engine/core"
	"github.com/hupe1980/golc/util"
	"time"
)

func main() {
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

	page := browser.Navigate("https://blog.csdn.net/daily886/article/details/96422463")
	res, _ := page.HTML()
	util.ParseHTMLAndGetStrippedStrings(res)
}
