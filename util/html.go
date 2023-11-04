package util

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParseHTMLAndGetStrippedStrings parses the HTML content and returns the stripped strings.
// It uses the goquery package to extract the text from HTML elements.
func ParseHTMLAndGetStrippedStrings(htmlContent string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	var strippedStrings []string

	doc.Find("body > *").Each(func(_ int, s *goquery.Selection) {
		strippedString := strings.TrimSpace(s.Text())
		if strippedString != "" {
			strippedStrings = append(strippedStrings, strippedString)
		}
	})
	//fmt.Println("ParseHTMLAndGetStrippedStrings" + compressStr(strings.Join(strippedStrings, " ")))
	return compressStr(strings.Join(strippedStrings, " ")), nil
}

// 利用正则表达式压缩字符串，去除空格或制表符
func compressStr(str string) string {
	if str == "" {
		return ""
	}
	//匹配一个或多个空白符的正则表达式
	reg := regexp.MustCompile("\\s+")
	return reg.ReplaceAllString(str, "")
}
