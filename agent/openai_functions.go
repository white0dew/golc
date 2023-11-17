package agent

import (
	"context"
	"errors"
	"fmt"
	"github.com/hupe1980/golc"
	"github.com/hupe1980/golc/model"
	"github.com/hupe1980/golc/prompt"
	"github.com/hupe1980/golc/rag"
	"github.com/hupe1980/golc/schema"
	"github.com/hupe1980/golc/tool"
)

// Compile time check to ensure OpenAIFunctions satisfies the agent interface.
var _ schema.Agent = (*OpenAIFunctions)(nil)

// OpenAIFunctionsOptions represents the configuration options for the OpenAIFunctions agent.
type OpenAIFunctionsOptions struct {
	*schema.CallbackOptions
	// OutputKey is the key to store the output of the agent in the ChainValues.
	OutputKey                 string
	SystemMessage             *prompt.SystemMessageTemplate
	ExtraMessages             []prompt.MessageTemplate
	MaxIterations             int
	RetrievalQAOptionsOptions *rag.RetrievalQAOptions
	Retriever                 schema.Retriever
	MaxTokenLimit             uint
	ChatMessageHistory        schema.ChatMessageHistory
}

// OpenAIFunctions is an agent that uses OpenAI chatModels and schema.Tools to perform actions.
type OpenAIFunctions struct {
	model     schema.ChatModel
	functions []schema.FunctionDefinition // 插件系统
	opts      OpenAIFunctionsOptions
	retriever schema.Retriever // 知识库查询
}

// NewOpenAIFunctions creates a new instance of the OpenAIFunctions agent with the given model and tools.
// It returns an error if the model is not an OpenAI chatModel or fails to convert tools to function definitions.
func NewOpenAIFunctions(model schema.ChatModel, tools []schema.Tool, optFns ...func(o *OpenAIFunctionsOptions)) (*Executor, error) {
	opts := OpenAIFunctionsOptions{
		CallbackOptions: &schema.CallbackOptions{
			Verbose: golc.Verbose,
		},
		OutputKey:     "output",
		SystemMessage: prompt.NewSystemMessageTemplate("You are a helpful AI assistant."),
		ExtraMessages: []prompt.MessageTemplate{},
		MaxIterations: DefaultMaxIterations,
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	if model.Type() != "chatmodel.OpenAI" {
		return nil, errors.New("agent only supports OpenAI chatModels")
	}

	//// 获取相关文档
	//docs, err := a.getDocuments(context.Background(), question, opts)
	//if err != nil {
	//	return nil, err
	//}

	// 插件系统
	functions := make([]schema.FunctionDefinition, len(tools))
	for i, t := range tools {
		f, err := tool.ToFunction(t)
		if err != nil {
			return nil, err
		}

		functions[i] = *f
	}
	agent := &OpenAIFunctions{
		model:     model,
		functions: functions,
		opts:      opts,
	}

	return NewExecutor(agent, tools, func(o *ExecutorOptions) {
		o.MaxIterations = opts.MaxIterations
	})
}

// Plan executes the agent with the given context, intermediate steps, and inputs.
// It returns the agent actions, agent finish, or an error, if any.
func (a *OpenAIFunctions) Plan(ctx context.Context, intermediateSteps []schema.AgentStep, inputs schema.ChainValues) ([]*schema.AgentAction, *schema.AgentFinish, error) {
	var promptMessages schema.ChatMessages
	var err error
	stepMessage := a.constructScratchPad(intermediateSteps)
	if a.opts.ChatMessageHistory == nil { // 模板化
		inputs["agentScratchpad"] = stepMessage
		templates := []prompt.MessageTemplate{a.opts.SystemMessage}
		templates = append(templates, a.opts.ExtraMessages...)
		templates = append(templates, prompt.NewHumanMessageTemplate("{{.input}}"))

		chatTemplate := prompt.NewChatTemplate(templates)

		placeholder := prompt.NewMessagesPlaceholder("agentScratchpad")

		wrapper := prompt.NewChatTemplateWrapper(chatTemplate, placeholder)
		prompt, err := wrapper.FormatPrompt(inputs)
		if err != nil {
			return nil, nil, err
		}
		promptMessages = prompt.Messages()
	} else { // 非模板化
		// 历史消息
		promptMessages, err = a.opts.ChatMessageHistory.Messages(ctx)
		if err != nil {
			return nil, nil, err
		}

		// 当前用户问题
		promptMessages = append(promptMessages, schema.NewHumanChatMessage(inputs["input"].(string)))

		// 当前步骤
		promptMessages = append(promptMessages, stepMessage...)
	}

	result, err := model.ChatModelGenerate(ctx, a.model, promptMessages, func(o *model.Options) {
		o.Functions = a.functions
	})
	if err != nil {
		return nil, nil, err
	}

	msg := result.Generations[0].Message

	aiMsg, ok := msg.(*schema.AIChatMessage)
	if !ok {
		return nil, nil, fmt.Errorf("unexpected chatMessage type. Expected ai, but got %s", msg.Type())
	}

	ext := aiMsg.Extension()

	if ext.FunctionCall != nil {
		toolInput := schema.NewToolInputFromArguments(ext.FunctionCall.Arguments)

		msgContent := ""
		if aiMsg.Content() != "" {
			msgContent = fmt.Sprintf("responded: %s", aiMsg.Content())
		}

		log := fmt.Sprintf("\nInvoking `%s` with `%s`\n%s\n", ext.FunctionCall.Name, toolInput, msgContent)

		return []*schema.AgentAction{
			{Tool: ext.FunctionCall.Name, ToolInput: toolInput, Log: log, MessageLog: schema.ChatMessages{aiMsg}},
		}, nil, nil
	}

	return nil, &schema.AgentFinish{
		ReturnValues: map[string]any{
			a.opts.OutputKey: aiMsg.Content(),
		},
		Log: aiMsg.Content(),
	}, nil
}

// InputKeys returns the expected input keys for the agent.
func (a *OpenAIFunctions) InputKeys() []string {
	return []string{"input"}
}

// OutputKeys returns the output keys that the agent will return.
func (a *OpenAIFunctions) OutputKeys() []string {
	return []string{a.opts.OutputKey}
}

// constructScratchPad constructs the scratch pad from the given intermediate steps.
func (a *OpenAIFunctions) constructScratchPad(steps []schema.AgentStep) schema.ChatMessages {
	messages := schema.ChatMessages{}

	for _, step := range steps {
		if step.Action.MessageLog != nil {
			messages = append(messages, step.Action.MessageLog...)
			messages = append(messages, schema.NewFunctionChatMessage(step.Action.Tool, step.Observation))
		} else {
			messages = append(messages, schema.NewAIChatMessage(step.Action.Log))
		}
	}

	return messages
}

//func (a *OpenAIFunctions) getDocuments(ctx context.Context, query string, opts schema.CallOptions) ([]schema.Document, error) {
//	docs, err := retriever.Run(ctx, a.retriever, query, func(o *retriever.Options) {
//		o.Callbacks = opts.CallbackManger.GetInheritableCallbacks()
//		o.ParentRunID = opts.CallbackManger.RunID()
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	numDocs := len(docs)
//
//	if a.opts.MaxTokenLimit > 0 {
//		tokens := make([]uint, len(docs))
//
//		for i, doc := range docs {
//			t, err := a.llmChain.GetNumTokens(doc.PageContent)
//			if err != nil {
//				return nil, err
//			}
//
//			tokens[i] = t
//		}
//
//		tokenCount := util.SumInt(tokens[:numDocs])
//		for tokenCount > a.opts.MaxTokenLimit {
//			numDocs--
//			tokenCount -= tokens[numDocs]
//		}
//	}
//
//	return docs[:numDocs], nil
//}
