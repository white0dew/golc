package tool

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/hupe1980/golc/schema"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// Compile time check to ensure Human satisfies the Tool interface.
var _ schema.Tool = (*BingWebSearch)(nil)

// BingSearchOptions contains options for configuring the Human tool.
type BingSearchOptions struct {
	Token string `json:"token"`
}

type BingWebSearch struct {
	opts BingSearchOptions
}

// NewBingWebSearch creates a new instance of the Human tool with the provided options.
func NewBingWebSearch(optFns ...func(o *BingSearchOptions)) *BingWebSearch {
	opts := BingSearchOptions{}

	for _, fn := range optFns {
		fn(&opts)
	}

	return &BingWebSearch{
		opts: opts,
	}
}

// important 注意,函数名称不能够有空格!

// Name returns the name of the tool.
func (t *BingWebSearch) Name() string {
	return "BingWebSearch"
}

// Description returns the description of the tool.
func (t *BingWebSearch) Description() string {
	return `BingWebSearch analyzes the keywords inputted by users and ranks web pages based on relevance and weight, in order to present the most relevant and useful search results.Input should be a search query.`
}

// ArgsType returns the type of the input argument expected by the tool.
func (t *BingWebSearch) ArgsType() reflect.Type {
	return reflect.TypeOf("") // string
}

// Run executes the tool with the given input and returns the output.
func (t *BingWebSearch) Run(ctx context.Context, input any) (string, error) {
	query, ok := input.(string)
	if !ok {
		return "", errors.New("illegal input type")
	}

	searchList, err := t.getBingSearchAPIResult(query, 10)
	if err != nil {
		return "", err
	}

	return strings.Join(searchList, ""), err
}

// Verbose returns the verbosity setting of the tool.
func (t *BingWebSearch) Verbose() bool {
	return true
}

// Callbacks returns the registered callbacks of the tool.
func (t *BingWebSearch) Callbacks() []schema.Callback {
	return nil
}

// The is the struct for the data returned by Bing.
type BingAnswer struct {
	Type         string `json:"_type"`
	QueryContext struct {
		OriginalQuery string `json:"originalQuery"`
	} `json:"queryContext"`
	WebPages struct {
		WebSearchURL          string `json:"webSearchUrl"`
		TotalEstimatedMatches int    `json:"totalEstimatedMatches"`
		Value                 []struct {
			ID               string    `json:"id"`
			Name             string    `json:"name"`
			URL              string    `json:"url"`
			IsFamilyFriendly bool      `json:"isFamilyFriendly"`
			DisplayURL       string    `json:"displayUrl"`
			Snippet          string    `json:"snippet"`
			DateLastCrawled  time.Time `json:"dateLastCrawled"`
			SearchTags       []struct {
				Name    string `json:"name"`
				Content string `json:"content"`
			} `json:"searchTags,omitempty"`
			About []struct {
				Name string `json:"name"`
			} `json:"about,omitempty"`
		} `json:"value"`
	} `json:"webPages"`
	RelatedSearches struct {
		ID    string `json:"id"`
		Value []struct {
			Text         string `json:"text"`
			DisplayText  string `json:"displayText"`
			WebSearchURL string `json:"webSearchUrl"`
		} `json:"value"`
	} `json:"relatedSearches"`
	RankingResponse struct {
		Mainline struct {
			Items []struct {
				AnswerType  string `json:"answerType"`
				ResultIndex int    `json:"resultIndex"`
				Value       struct {
					ID string `json:"id"`
				} `json:"value"`
			} `json:"items"`
		} `json:"mainline"`
		Sidebar struct {
			Items []struct {
				AnswerType string `json:"answerType"`
				Value      struct {
					ID string `json:"id"`
				} `json:"value"`
			} `json:"items"`
		} `json:"sidebar"`
	} `json:"rankingResponse"`
}

// getBingSearchAPIResult
func (t *BingWebSearch) getBingSearchAPIResult(searchTerm string, answerCount int32) ([]string, error) {
	const endpoint = "https://api.bing.microsoft.com/v7.0/search"
	// Declare a new GET request.
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Add the payload to the request.
	param := req.URL.Query()
	param.Add("q", searchTerm)
	param.Add("answerCount", string(answerCount))
	req.URL.RawQuery = param.Encode()

	// Insert the request header.
	req.Header.Add("Ocp-Apim-Subscription-Key", t.opts.Token)

	// Create a new client.
	client := new(http.Client)

	// Send the request to Bing.
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Close the response.
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Create a new answer.
	ans := new(BingAnswer)
	err = json.Unmarshal(body, &ans)
	if err != nil {
		return nil, err
	}
	var res []string
	for _, result := range ans.WebPages.Value {
		res = append(res, result.Snippet)
	}
	return res, nil
}
