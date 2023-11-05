package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hupe1980/golc/schema"
	"io"
	"net/http"
	"reflect"
)

var _ schema.Tool = (*ChengYuDict)(nil)

// ChengYuDictOptions contains options for configuring the Human tool.
type ChengYuDictOptions struct {
	RankNum int `json:"rank_num"`
}

// ChengYuDict is a tool that allows interaction with a human user.
type ChengYuDict struct {
	opts ChengYuDictOptions
}

func NewChengYuDict(optFns ...func(o *ChengYuDictOptions)) *ChengYuDict {
	opts := ChengYuDictOptions{
		RankNum: 3,
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	return &ChengYuDict{
		opts: opts,
	}
}

func (c ChengYuDict) Name() string {
	return `Chinese_Idioms_Dict`
}

func (c ChengYuDict) Description() string {
	return `Search for the pinyin, meaning, origin, and example sentences of Chinese idioms in the ChengYu Dictionary. The input should be a valid Chinese idiom.`
}

func (c ChengYuDict) Run(ctx context.Context, input any) (string, error) {
	query, ok := input.(string)
	if !ok {
		return "", errors.New("illegal input type")
	}
	chengYu, err := getIdiomInfo(query)
	if err != nil {
		return "", err
	}
	if chengYu.Data.Idiom == "" {
		return "", errors.New("未查询到相关成语")
	}
	// TODO 可以考虑对所有tools进行chongshi
	res, _ := json.Marshal(chengYu.Data)
	return string(res), nil
}

func (c ChengYuDict) ArgsType() reflect.Type {
	return reflect.TypeOf("") // string
}

func (c ChengYuDict) Verbose() bool {
	return true
}

func (c ChengYuDict) Callbacks() []schema.Callback {
	return nil
}

type IdiomResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Idiom   string `json:"成语"`
		Pinyin  string `json:"拼音"`
		Shiyi   string `json:"释义"`
		Chuchu  string `json:"出处"`
		Liju    string `json:"例句"`
		Message string `json:"info"`
	} `json:"data"`
	Time int64 `json:"time"`
}

func getIdiomInfo(idiom string) (*IdiomResponse, error) {
	url := fmt.Sprintf("https://api.vore.top/api/idiom?q=%s", idiom)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data IdiomResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	if data.Code != 200 {
		return nil, errors.New("成语查询失败,可尝试重试")
	}
	return &data, nil
}
