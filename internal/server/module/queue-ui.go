package module

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"go.uber.org/fx"
	"riverqueue.com/riverui"
)

// depends on queue & http
func QueueUIModule() fx.Option {
	return fx.Module(
		"queue-ui",
		fx.Invoke(registerRiverUI),
	)
}

func registerRiverUI(
	lc fx.Lifecycle,
	runCtx runtimeContext,
	mux *http.ServeMux,
	client *river.Client[pgx.Tx],
) error {
	endpoints := riverui.NewEndpoints(client, nil)

	prefix := "/admin/queue/"

	opts := &riverui.HandlerOpts{
		Endpoints: endpoints,
		Logger:    slog.Default(),
		Prefix:    prefix,
	}

	handler, err := riverui.NewHandler(opts)

	if err != nil {
		return err
	}

	mux.Handle(prefix, handler)

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			return handler.Start(runCtx)
		},
	})

	return nil
}
