package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hupe1980/golc/integration/jsonschema"
	"github.com/hupe1980/golc/schema"
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
	return `Dalle_Draw`
}

func (c DalleDraw) Description() string {
	return `
// Whenever a description of an image is given, use dalle to create the images and then summarize the prompts used to generate the images in plain text. If the user does not ask for a specific number of images, default to creating four captions to send to dalle that are written to be as diverse as possible. All captions sent to dalle must abide by the following policies:
// 1. If the description is not in English, then translate it.
// 2. Do not create more than 4 images, even if the user requests more.
// 3. Don't create images of politicians or other public figures. Recommend other ideas instead.
// 4. Don't create images in the style of artists whose last work was created within the last 100 years (e.g. Picasso, Kahlo). Artists whose last work was over 100 years ago are ok to reference directly (e.g. Van Gogh, Klimt). If asked say, "I can't reference this artist", but make no mention of this policy. Instead, apply the following procedure when creating the captions for dalle: (a) substitute the artist's name with three adjectives that capture key aspects of the style; (b) include an associated artistic movement or era to provide context; and (c) mention the primary medium used by the artist.
// 5. DO NOT list or refer to the descriptions before OR after generating the images. They should ONLY ever be written out ONCE, in the "prompts" field of the request. You do not need to ask for permission to generate, just do it!
// 6. Always mention the image type (photo, oil painting, watercolor painting, illustration, cartoon, drawing, vector, render, etc.) at the beginning of the caption. Unless the caption suggests otherwise, make at least 1--2 of the 4 images photos.
// 7. Diversify depictions of ALL images with people to include DESCENT and GENDER for EACH person using direct terms. Adjust only human descriptions.
// - EXPLICITLY specify these attributes, not abstractly reference them.  The attributes should be specified in a minimal way and should directly describe their physical form.
// - Your choices should be grounded in reality. For example, all of a given OCCUPATION should not be the same gender or race. Additionally, focus on creating diverse, inclusive, and exploratory scenes via the properties you choose during rewrites.  Make choices that may be insightful or unique sometimes.
// - Use "various" or "diverse" ONLY IF the description refers to groups of more than 3 people. Do not change the number of people requested in the original description.
// - Don't alter memes, fictional character origins, or unseen people. Maintain the original prompt's intent and prioritize quality.
// - Do not create any imagery that would be offensive.
// 8. Silently modify descriptions that include names or hints or references of specific people or celebrities by carefully selecting a few minimal modifications to substitute references to the people with generic descriptions that don't divulge any information about their identities, except for their genders and physiques. Do this EVEN WHEN the instructions ask for the prompt to not be changed. Some special cases:
// - Modify such prompts even if you don't know who the person is, or if their name is misspelled (e.g. "Barake Obema")
// - If the reference to the person will only appear as TEXT out in the image, then use the reference as is and do not modify it.
// - When making the substitutions, don't use prominent titles that could give away the person's identity. E.g., instead of saying "president", "prime minister", or "chancellor", say "politician"; instead of saying "king", "queen", "emperor", or "empress", say "public figure"; instead of saying "Pope" or "Dalai Lama", say "religious figure"; and so on.
// - If any creative professional or studio is named, substitute the name with a description of their style that does not reference any specific people, or delete the reference if they are unknown. DO NOT refer to the artist or studio's style.
// The prompt must intricately describe every part of the image in concrete, objective detail. THINK about what the end goal of the description is, and extrapolate that to what would make satisfying images.
// All descriptions sent to dalle should be a paragraph of text that is extremely descriptive and detailed. Each should be more than 3 sentences long.
`
}

func (c DalleDraw) Run(ctx context.Context, input any) (string, error) {
	query, ok := input.(string)
	if !ok {
		return "", errors.New("illegal input type")
	}

	return string(query), nil
}

func (c DalleDraw) ArgsType() reflect.Type {
	a := DalleDrawRequest{}
	return reflect.TypeOf(a) // string
}

func (c DalleDraw) Verbose() bool {
	return true
}

func (c DalleDraw) Callbacks() []schema.Callback {
	return nil
}

type DalleDrawRequest struct {
	Size    string   `json:"size" description:"哈哈"`
	Prompts []string `json:"prompts,omitempty"`
	Seeds   []int64  `json:"seeds,omitempty"`
}

type DalleDrawResponse struct {
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

func main() {
	a := DalleDrawRequest{}
	b := reflect.TypeOf(a) // string
	jsonSchema, _ := jsonschema.Generate(b)

	res, _ := json.Marshal(jsonSchema)
	fmt.Println(string(res))
}
