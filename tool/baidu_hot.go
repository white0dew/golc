package tool

import (
	"context"
	"encoding/json"
	"github.com/hupe1980/golc/schema"
	"io"
	"net/http"
	"reflect"
)

// Compile time check to ensure Human satisfies the Tool interface.
var _ schema.Tool = (*BaiduHot)(nil)

// BaiduHotOptions contains options for configuring the Human tool.
type BaiduHotOptions struct {
	// Function for displaying prompts.
	PromptFunc PromptFunc

	// Function for retrieving user input.
	InputFunc InputFunc
}

// BaiduHot is a tool that allows interaction with a human user.
type BaiduHot struct {
	opts BaiduHotOptions
}

// NewBaiduHot creates a new instance of the Human tool with the provided options.
func NewBaiduHot(optFns ...func(o *BaiduHotOptions)) *BaiduHot {
	opts := BaiduHotOptions{}

	for _, fn := range optFns {
		fn(&opts)
	}

	return &BaiduHot{
		opts: opts,
	}
}

// important 注意,函数名称不能够有空格!

// Name returns the name of the tool.
func (t *BaiduHot) Name() string {
	return "BaiduHot"
}

// Description returns the description of the tool.
func (t *BaiduHot) Description() string {
	return `返回当前百度网站的新闻热点热榜`
}

// ArgsType returns the type of the input argument expected by the tool.
func (t *BaiduHot) ArgsType() reflect.Type {
	return reflect.TypeOf("") // string
}

// Run executes the tool with the given input and returns the output.
func (t *BaiduHot) Run(ctx context.Context, input any) (string, error) {
	//query, ok := input.(string)
	//if !ok {
	//	return "", errors.New("illegal input type")
	//}

	baiduList, err := GetBaiduHot()
	if err != nil {
		return "", err
	}

	res, err := json.Marshal(baiduList)
	if err != nil {
		return "", err
	}

	return string(res), err
}

// Verbose returns the verbosity setting of the tool.
func (t *BaiduHot) Verbose() bool {
	return true
}

// Callbacks returns the registered callbacks of the tool.
func (t *BaiduHot) Callbacks() []schema.Callback {
	return nil
}

type HotItem struct {
	Rank int    `json:"rank"`
	Name string `json:"name"`
	Link string `json:"link"`
	Heat string `json:"heat"`
}

type ResponseData struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data []HotItem `json:"data"`
	Time int64     `json:"time"`
}

func GetBaiduHot() ([]HotItem, error) {
	resp, err := http.Get("https://api.vore.top/api/baiduHot")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var responseData ResponseData
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return nil, err
	}

	return responseData.Data, nil
}
