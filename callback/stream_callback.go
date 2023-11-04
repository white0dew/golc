package callback

import (
	"context"
	"fmt"
	"github.com/hupe1980/golc/schema"
	"github.com/olahol/melody"
	"io"
	"os"
)

// Compile time check to ensure StreamWriterHandler satisfies the Callback interface.
var _ schema.Callback = (*OwnStreamWriterHandler)(nil)

type OwnStreamWriterHandlerOptions struct {
	Writer        io.Writer
	SocketSession melody.Session
}

type OwnStreamWriterHandler struct {
	NoopHandler
	writer io.Writer
	opts   OwnStreamWriterHandlerOptions
}

func NewOwnStreamWriterHandlerOptions(optFns ...func(o *OwnStreamWriterHandlerOptions)) *OwnStreamWriterHandler {
	opts := OwnStreamWriterHandlerOptions{
		Writer: os.Stdout,
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	return &OwnStreamWriterHandler{
		writer: opts.Writer,
		opts:   opts,
	}
}

func (cb *OwnStreamWriterHandler) AlwaysVerbose() bool {
	return true
}

func (cb *OwnStreamWriterHandler) OnModelNewToken(ctx context.Context, input *schema.ModelNewTokenInput) error {
	//fmt.Fprint(cb.writer, input.Token)
	fmt.Println("here new token")
	return nil
}

func (cb *OwnStreamWriterHandler) OnToolStart(ctx context.Context, input *schema.ToolStartInput) error {
	//fmt.Println(input)
	fmt.Println("here OnToolStart")
	return nil
}

func (cb *OwnStreamWriterHandler) OnToolEnd(ctx context.Context, input *schema.ToolEndInput) error {
	//fmt.Println(input)
	fmt.Println("here OnToolEnd")
	return nil
}

func (cb *OwnStreamWriterHandler) OnToolError(ctx context.Context, input *schema.ToolErrorInput) error {
	//fmt.Println(input)
	fmt.Println("here OnToolError")
	return nil
}
