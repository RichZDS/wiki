package callback

import (
	"context"
	"sync"

	"github.com/cloudwego/eino/callbacks"

	"wiki/pkg/logger"
)

var registerEinoGlobalHandlersOnce sync.Once

// RegisterEinoGlobalHandlers registers process-wide Eino trace callbacks.
func RegisterEinoGlobalHandlers() {
	registerEinoGlobalHandlersOnce.Do(func() {
		log := logger.GetLogger()
		handler := callbacks.NewHandlerBuilder().
			OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
				log.Printf("[trace] %s/%s start", info.Component, info.Name)
				return ctx
			}).
			OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
				log.Printf("[trace] %s/%s end", info.Component, info.Name)
				return ctx
			}).
			OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
				log.Printf("[trace] %s/%s error: %v", info.Component, info.Name, err)
				return ctx
			}).
			Build()

		callbacks.AppendGlobalHandlers(handler)
	})
}
