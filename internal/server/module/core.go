package module

import (
	"context"

	"go.uber.org/fx"
)

type runtimeContext context.Context

func CoreModule() fx.Option {
	return fx.Module(
		"core",
		fx.Provide(provideRuntimeContext),
	)
}

func provideRuntimeContext(lc fx.Lifecycle) runtimeContext {
	ctx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.StopHook(cancel))

	return ctx
}
