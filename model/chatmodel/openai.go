package chatmodel

import (
	"context"
	"errors"
	"github.com/avast/retry-go"
	"github.com/hupe1980/golc"
	"github.com/hupe1980/golc/callback"
	"github.com/hupe1980/golc/integration"
	"github.com/hupe1980/golc/schema"
	"github.com/hupe1980/golc/tokenizer"
	"github.com/hupe1980/golc/util"
	"github.com/sashabaranov/go-openai"
	"io"
)

// Compile time check to ensure OpenAI satisfies the ChatModel interface.
var _ schema.ChatModel = (*OpenAI)(nil)

// OpenAIClient is an interface for the OpenAI chat model client.
type OpenAIClient interface {
	CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (response openai.ChatCompletionResponse, err error)
	CreateChatCompletionStream(ctx context.Context, request openai.ChatCompletionRequest) (stream *openai.ChatCompletionStream, err error)
}

// OpenAIOptions contains the options for the OpenAI chat model.
type OpenAIOptions struct {
	*schema.CallbackOptions `map:"-"`
	schema.Tokenizer        `map:"-"`
	// Model name to use.
	ModelName string `map:"model_name,omitempty"`
	// Sampling temperature to use.
	Temperature float32 `map:"temperature,omitempty"`
	// The maximum number of tokens to generate in the completion.
	// -1 returns as many tokens as possible given the prompt and
	//the models maximal context size.
	MaxTokens int `map:"max_tokens,omitempty"`
	// Total probability mass of tokens to consider at each step.
	TopP float32 `map:"top_p,omitempty"`
	// Penalizes repeated tokens.
	PresencePenalty float32 `map:"presence_penalty,omitempty"`
	// Penalizes repeated tokens according to frequency.
	FrequencyPenalty float32 `map:"frequency_penalty,omitempty"`
	// How many completions to generate for each prompt.
	N int `map:"n,omitempty"`
	// BaseURL is the base URL of the OpenAI service.
	BaseURL string `map:"base_url,omitempty"`
	// OrgID is the organization ID for accessing the OpenAI service.
	OrgID string `map:"org_id,omitempty"`
	// MaxRetries represents the maximum number of retries to make when generating.
	MaxRetries uint `map:"max_retries,omitempty"`
	// Stream indicates whether to stream the results or not.
	Stream bool `map:"stream,omitempty"`
}

var DefaultOpenAIOptions = OpenAIOptions{
	CallbackOptions: &schema.CallbackOptions{
		Verbose: golc.Verbose,
	},
	ModelName:        openai.GPT3Dot5Turbo,
	Temperature:      1,
	TopP:             1,
	PresencePenalty:  0,
	FrequencyPenalty: 0,
	MaxRetries:       3,
}

// OpenAI represents the OpenAI chat model.
type OpenAI struct {
	schema.Tokenizer
	client OpenAIClient
	opts   OpenAIOptions
}

// NewOpenAI creates a new instance of the OpenAI chat model.
func NewOpenAI(apiKey string, optFns ...func(o *OpenAIOptions)) (*OpenAI, error) {
	opts := OpenAIOptions{
		CallbackOptions: &schema.CallbackOptions{
			Verbose: golc.Verbose,
		},
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	config := openai.DefaultConfig(apiKey)

	if opts.BaseURL != "" {
		config.BaseURL = opts.BaseURL
	}

	if opts.OrgID != "" {
		config.OrgID = opts.OrgID
	}

	client := openai.NewClientWithConfig(config)

	return NewOpenAIFromClient(client, optFns...)
}

// NewOpenAIFromClient creates a new instance of the OpenAI chat model with the provided client and options.
func NewOpenAIFromClient(client OpenAIClient, optFns ...func(o *OpenAIOptions)) (*OpenAI, error) {
	opts := DefaultOpenAIOptions

	for _, fn := range optFns {
		fn(&opts)
	}

	if opts.Tokenizer == nil {
		opts.Tokenizer = tokenizer.NewOpenAI(opts.ModelName)
	}

	return &OpenAI{
		Tokenizer: opts.Tokenizer,
		client:    client,
		opts:      opts,
	}, nil
}

// Generate generates text based on the provided chat messages and options.
func (cm *OpenAI) Generate(ctx context.Context, messages schema.ChatMessages, optFns ...func(o *schema.GenerateOptions)) (*schema.ModelResult, error) {
	opts := schema.GenerateOptions{
		CallbackManger: &callback.NoopManager{},
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	openAIMessages, err := integration.ToOpenAIChatCompletionMessages(messages)
	if err != nil {
		return nil, err
	}

	var functions []openai.FunctionDefinition
	if opts.Functions != nil {
		functions = util.Map(opts.Functions, func(fd schema.FunctionDefinition, i int) openai.FunctionDefinition {
			return openai.FunctionDefinition{
				Name:        fd.Name,
				Description: fd.Description,
				Parameters:  fd.Parameters,
			}
		})
	}
	completionRequest := openai.ChatCompletionRequest{
		Model:            cm.opts.ModelName,
		Temperature:      cm.opts.Temperature,
		MaxTokens:        cm.opts.MaxTokens,
		TopP:             cm.opts.TopP,
		N:                cm.opts.N,
		PresencePenalty:  cm.opts.PresencePenalty,
		FrequencyPenalty: cm.opts.PresencePenalty,
		Messages:         openAIMessages,
		Functions:        functions,
		Stop:             opts.Stop,
	}

	chatResponse := openai.ChatCompletionResponse{
		ID:      "",
		Object:  "",
		Created: 0,
		Model:   "",
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionMessage{
					Role:         "",
					Content:      "",
					Name:         "",
					FunctionCall: nil,
				},
				FinishReason: "",
			},
		},
		Usage: openai.Usage{},
	}

	tokenUsage := make(map[string]int)
	if cm.opts.Stream {
		completionRequest.Stream = true
		//havaFunctionCall := false // import 是否出现函数调用

		stream, err := cm.client.CreateChatCompletionStream(ctx, completionRequest)
		if err != nil {
			return nil, err
		}

		defer stream.Close()
		tokens := []string{}

	streamProcessing:
		for {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				res, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					break streamProcessing
				}

				if err != nil {
					return nil, err
				}

				// 防止有空的
				if len(res.Choices) == 0 {
					continue
				}

				chatResponse.Choices[0].Message.Content += res.Choices[0].Delta.Content
				if res.Choices[0].Delta.Role != "" {
					chatResponse.Choices[0].Message.Role = res.Choices[0].Delta.Role
				}

				chatResponse.Choices[0].Index = res.Choices[0].Index
				chatResponse.Choices[0].FinishReason = res.Choices[0].FinishReason

				if res.Choices[0].Delta.FunctionCall != nil {
					if chatResponse.Choices[0].Message.FunctionCall == nil {
						chatResponse.Choices[0].Message.FunctionCall = res.Choices[0].Delta.FunctionCall
					} else {
						chatResponse.Choices[0].Message.FunctionCall.Arguments += res.Choices[0].Delta.FunctionCall.Arguments
					}
				}

				if err := opts.CallbackManger.OnModelNewToken(ctx, &schema.ModelNewTokenManagerInput{
					Token:  res.Choices[0].Delta.Content,
					Choice: chatResponse.Choices,
				}); err != nil {
					return nil, err
				}
				tokens = append(tokens, res.Choices[0].Delta.Content)
			}
		}
	} else {
		res, err := cm.createChatCompletionWithRetry(ctx, completionRequest)
		if err != nil {
			return nil, err
		}
		chatResponse = res
		// stream-需要重新计算
		tokenUsage["CompletionTokens"] += chatResponse.Usage.CompletionTokens
		tokenUsage["PromptTokens"] += chatResponse.Usage.PromptTokens
		tokenUsage["TotalTokens"] += chatResponse.Usage.TotalTokens
	}

	return &schema.ModelResult{
		Generations: []schema.Generation{{
			Text:    chatResponse.Choices[0].Message.Content,
			Message: openAIResponseToChatMessage(chatResponse.Choices[0].Message),
		}},
		LLMOutput: map[string]any{
			"ModelName":  cm.opts.ModelName,
			"TokenUsage": tokenUsage,
		},
	}, nil
}

func (cm *OpenAI) createChatCompletionWithRetry(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	retryOpts := []retry.Option{
		retry.Attempts(cm.opts.MaxRetries),
		retry.DelayType(retry.FixedDelay),
		retry.RetryIf(func(err error) bool {
			e := &openai.APIError{}
			if errors.As(err, &e) {
				switch e.HTTPStatusCode {
				case 429, 500:
					return true
				default:
					return false
				}
			}

			return false
		}),
	}

	var res openai.ChatCompletionResponse

	err := retry.Do(
		func() error {
			r, cErr := cm.client.CreateChatCompletion(ctx, request)
			if cErr != nil {
				return cErr
			}
			res = r
			return nil
		},
		retryOpts...,
	)

	return res, err
}

// Type returns the type of the model.
func (cm *OpenAI) Type() string {
	return "chatmodel.OpenAI"
}

// Verbose returns the verbosity setting of the model.
func (cm *OpenAI) Verbose() bool {
	return cm.opts.CallbackOptions.Verbose
}

// Callbacks returns the registered callbacks of the model.
func (cm *OpenAI) Callbacks() []schema.Callback {
	return cm.opts.CallbackOptions.Callbacks
}

// InvocationParams returns the parameters used in the model invocation.
func (cm *OpenAI) InvocationParams() map[string]any {
	return util.StructToMap(cm.opts)
}

// openAIResponseToChatMessage converts an OpenAI ChatCompletionMessage to a schema.ChatMessage.
func openAIResponseToChatMessage(msg openai.ChatCompletionMessage) schema.ChatMessage {
	switch msg.Role {
	case "user":
		return schema.NewHumanChatMessage(msg.Content)
	case "assistant":
		if msg.FunctionCall != nil {
			return schema.NewAIChatMessage(msg.Content, func(o *schema.ChatMessageExtension) {
				o.FunctionCall = &schema.FunctionCall{
					Name:      msg.FunctionCall.Name,
					Arguments: msg.FunctionCall.Arguments,
				}
			})
		}

		return schema.NewAIChatMessage(msg.Content)
	case "system":
		return schema.NewSystemChatMessage(msg.Content)
	case "function":
		return schema.NewFunctionChatMessage(msg.Content, msg.Name)
	}

	return schema.NewGenericChatMessage(msg.Content, "unknown")
}
