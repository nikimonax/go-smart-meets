package module

import (
	"context"
	"errors"
	"net"
	"net/http"

	"go.uber.org/fx"

	"github.com/nikimonax/go-smart-meets/internal/server/config"
)

// depends on *config.Config
func HTTPModule() fx.Option {
	return fx.Module(
		"http",
		fx.Provide(
			provideHTTPMux,
			provideHTTPServer,
		),
		fx.Invoke(
			registerHTTPServerLifecycle,
		),
	)
}

func provideHTTPMux() *http.ServeMux {
	return http.NewServeMux()
}

func provideHTTPServer(mux *http.ServeMux) *http.Server {
	return &http.Server{Handler: mux}
}

func registerHTTPServerLifecycle(
	lc fx.Lifecycle,
	cfg *config.Config,
	srv *http.Server,
) {
	errs := make(chan error, 1)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lc := net.ListenConfig{}

			ln, err := lc.Listen(ctx, "tcp", cfg.HTTP.Listen.String())
			if err != nil {
				return err
			}

			go func() {
				errs <- srv.Serve(ln)
				close(errs)
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownErr := srv.Shutdown(ctx)

			if errors.Is(shutdownErr, http.ErrServerClosed) {
				shutdownErr = nil
			}

			select {
			case <-ctx.Done():
				return shutdownErr
			case serveErr := <-errs:
				return errors.Join(shutdownErr, serveErr)
			}
		},
	})
}
