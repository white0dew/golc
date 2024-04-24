package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hupe1980/golc/schema"
	"github.com/sashabaranov/go-openai"
	"reflect"
)

var _ schema.Tool = (*DalleDraw)(nil)

// DalleOptions contains options for configuring the Human tool.
type DalleOptions struct {
	RankNum int `json:"rank_num"`
}

// DalleDraw is a tool that allows interaction with a human user.
type DalleDraw struct {
	opts DalleOptions
}

func NewDalleDraw(optFns ...func(o *DalleOptions)) *DalleDraw {
	opts := DalleOptions{
		RankNum: 3,
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	return &DalleDraw{
		opts: opts,
	}
}

func (c DalleDraw) Name() string {
	return `Dalle`
}

func (c DalleDraw) Description() string {
	return `
// Whenever a description of an image is given, use dalle to create the images and then summarize the prompts used to generate the images in plain text. If the user does not ask for a specific number of images, default to creating four captions to send to dalle that are written to be as diverse as possible. All captions sent to dalle must abide by the following policies:
// 1. If the description is not in English, then translate it.
// 2. Do not create more than 4 images, even if the user requests more.
// 3. Don't create images of politicians or other public figures. Recommend other ideas instead.
// All descriptions sent to dalle should be a paragraph of text that is extremely descriptive and detailed. Each should be more than 3 sentences long.
`
}

func (c DalleDraw) Run(ctx context.Context, input any) (string, error) {
	query, ok := input.(*DalleDrawRequest)
	if !ok {
		return "", errors.New("DalleDraw illegal input type")
	}
	resp, err := drawImage(query)
	if err != nil {
		return "", err
	}

	res, _ := json.Marshal(resp)
	return string(res), nil
}

func (c DalleDraw) ArgsType() reflect.Type {
	a := DalleDrawRequest{}
	return reflect.TypeOf(a) // struct
}

func (c DalleDraw) Verbose() bool {
	return true
}

func (c DalleDraw) Callbacks() []schema.Callback {
	return nil
}

// 给json加上desc以及enum
type DalleDrawRequest struct {
	Size    string   `json:"size" default:"512*512" description:"图片尺寸" enum:"256*256,512*512,1024*1024"`
	Prompts []string `json:"prompts" description:"每一张图片的描述" `
	Seeds   []int64  `json:"seeds,omitempty" description:"随机种子"`
	PicNum  int      `json:"pic_num" default:"1" description:"生成图片数量" enum:"1,2,3,4"`
	Style   string   `json:"style" default:"vivid" description:"绘图风格" enum:"vivid, natural"`
	Quality string   `json:"quality" default:"hd" description:"图像质量" enum:"standard, hd"`
}

type DalleDrawResponse struct {
	ImageUrl []string `json:"image_url,omitempty" description:"图片地址"`
	Prompt   []string `json:"prompt,omitempty" description:"Prompt描述"`
	Seeds    []string `json:"seeds,omitempty" description:"随机种子"`
}

// 调用openai的dalle3接口绘图
func drawImage(r *DalleDrawRequest) (*DalleDrawResponse, error) {
	fmt.Printf("[drawImage] r:%+v", r)
	//// 获取key
	//randomKey := key_service.SelectAPIKey()
	//// 获取客户端
	//cli := openai_service.GetNewOpenaiClient(randomKey)
	//// Sample image by link
	//request := openai.ImageRequest{
	//	Prompt:         r.Prompts[0],
	//	Size:           openai.CreateImageSize512x512,
	//	ResponseFormat: openai.CreateImageResponseFormatURL,
	//	N:              1,
	//	Model:          openai.CreateImageModelDallE3,
	//	Style:          openai.CreateImageStyleVivid,
	//	Quality:        openai.CreateImageQualityStandard,
	//}
	//// 参数设置
	//request.Size = getImageSize(r.Size)
	//
	//respUrl, err := cli.CreateImage(context.Background(), request)
	//if err != nil {
	//	golog.Errorf("Image creation error: %v\n", err)
	//	return nil, nil
	//}
	//golog.Infof("Image URL: %v\n", respUrl)
	//
	//if len(respUrl.Data) == 0 {
	//	return nil, errors.New("error")
	//}
	//resp := &DalleDrawResponse{}
	//resp.ImageUrl = append(resp.ImageUrl, respUrl.Data[0].URL)
	return nil, nil
}

// 根据req的图片尺寸输出openai的图片尺寸
func getImageSize(size string) string {
	switch size {
	case "256*256":
		return openai.CreateImageSize256x256
	case "512*512":
		return openai.CreateImageSize512x512
	case "1024*1024":
		return openai.CreateImageSize1024x1024
	case "1024*1792":
		return openai.CreateImageSize1792x1024
	case "1792*1024":
		return openai.CreateImageSize1024x1792
	default:
		return openai.CreateImageSize512x512
	}
}
