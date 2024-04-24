// Package tool provides tools that agents can use to perform various operations.
package tool

import (
	"context"
	"github.com/gofiber/fiber/v2/log"
	"reflect"

	"github.com/hupe1980/golc/callback"
	"github.com/hupe1980/golc/integration/jsonschema"
	"github.com/hupe1980/golc/schema"
)

type Options struct {
	Callbacks   []schema.Callback
	ParentRunID string
}

func Run(ctx context.Context, t schema.Tool, input *schema.ToolInput, optFns ...func(o *Options)) (string, error) {
	opts := Options{}

	for _, fn := range optFns {
		fn(&opts)
	}

	cm := callback.NewManager(opts.Callbacks, t.Callbacks(), t.Verbose())

	rm, err := cm.OnToolStart(ctx, &schema.ToolStartManagerInput{
		ToolName: t.Name(),
		Input:    input,
	})
	if err != nil {
		return "", err
	}

	var inputValue any

	if input.Structured() {
		value := reflect.New(t.ArgsType())
		ptr := value.Interface()

		if unErr := input.Unmarshal(ptr); unErr != nil {
			return "", unErr
		}

		inputValue = reflect.ValueOf(ptr).Elem().Interface()
	} else {
		inputValue, _ = input.GetString()
	}

	output, err := t.Run(ctx, inputValue)
	if err != nil {
		if cbErr := rm.OnToolError(ctx, &schema.ToolErrorManagerInput{
			Error:    err,
			ToolName: t.Name(),
		}); cbErr != nil {
			return "", cbErr
		}

		return "", err
	}

	if err := rm.OnToolEnd(ctx, &schema.ToolEndManagerInput{
		Output:   output,
		ToolName: t.Name(),
	}); err != nil {
		return "", err
	}

	return output, nil
}

// ToFunction formats a tool into a function API
func ToFunction(t schema.Tool) (*schema.FunctionDefinition, error) {
	function := &schema.FunctionDefinition{
		Name:        t.Name(),
		Description: t.Description(),
	}

	argsType := t.ArgsType()

	// TODO fixed
	if argsType.Kind() == reflect.String {
		function.Parameters = schema.FunctionDefinitionParameters{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"__arg1": {
					Type:        "string",
					Description: "__arg1",
				},
			},
			Required: []string{"__arg1"},
		}

		return function, nil
	}

	//log.Infof("[ToFunction] argsType:%+v", argsType)
	jsonSchema, err := jsonschema.Generate(argsType)
	if err != nil {
		log.Warnf("[jsonschema.Generate] failed to generate json schema for %s: %s", t.Name(), err)
		return nil, err
	}
	//log.Infof("[ToFunction] jsonSchema:%+v", util.ToString(jsonSchema))

	function.Parameters = schema.FunctionDefinitionParameters{
		Type:       "object",
		Properties: jsonSchema.Properties,
		Required:   jsonSchema.Required,
	}

	return function, nil
}
