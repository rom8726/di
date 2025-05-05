package di

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

const (
	DefaultStartTimeout = 30 * time.Second
	DefaultStopTimeout  = 30 * time.Second
)

type Servicer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type App struct {
	container *Container

	logger       *slog.Logger
	startTimeout time.Duration
	stopTimeout  time.Duration
}

type AppOpt func(*App)

func WithLogger(logger *slog.Logger) AppOpt {
	return func(app *App) {
		app.logger = logger
	}
}

func WithStartTimeout(timeout time.Duration) AppOpt {
	return func(app *App) {
		app.startTimeout = timeout
	}
}

func WithStopTimeout(timeout time.Duration) AppOpt {
	return func(app *App) {
		app.stopTimeout = timeout
	}
}

func NewApp(container *Container, opts ...AppOpt) *App {
	app := &App{
		container:    container,
		startTimeout: DefaultStartTimeout,
		stopTimeout:  DefaultStopTimeout,
	}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

func (app *App) Run(ctx context.Context) error {
	if err := app.runStart(ctx); err != nil {
		_ = app.runStop(ctx)

		return err
	}

	<-ctx.Done()

	return app.runStop(context.Background())
}

func (app *App) Start(ctx context.Context) error {
	app.logInfo("Starting...")

	var services []Servicer
	for _, instance := range app.container.instancesList {
		service, ok := instance.(Servicer)
		if ok {
			services = append(services, service)
		}
	}

	var err error
	for _, service := range services {
		if err = withTimeout(ctx, service.Start); err != nil {
			break
		}
	}

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		app.logError("Start timed out.")

		return err

	case err != nil:
		app.logError("Failed to start: %v", err)

		return err
	}

	app.logInfo("Started.")

	return nil
}

func (app *App) Stop(ctx context.Context) error {
	app.logInfo("Stopping...")

	var services []Servicer
	for i := len(app.container.instancesList) - 1; i >= 0; i-- {
		instance := app.container.instancesList[i]
		service, ok := instance.(Servicer)
		if ok {
			services = append(services, service)
		}
	}

	var err error
	for _, service := range services {
		if stopErr := withTimeout(ctx, service.Stop); stopErr != nil {
			if err == nil {
				err = stopErr
			}
		}
	}

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		app.logError("Stop timed out.")

		return nil
	case err != nil:
		app.logError("Failed to stop cleanly: %v", err)

		return err
	}

	app.logInfo("Stopped.")

	return nil
}

func (app *App) runStart(ctx context.Context) error {
	if app.startTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, app.startTimeout)
		defer cancel()
	}

	return app.Start(ctx)
}

func (app *App) runStop(ctx context.Context) error {
	if app.stopTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, app.stopTimeout)
		defer cancel()
	}

	return app.Stop(ctx)
}

func (app *App) logInfo(msg string, args ...any) {
	if app.logger == nil {
		return
	}

	app.logger.Info(fmt.Sprintf(msg, args...))
}

func (app *App) logError(msg string, args ...any) {
	if app.logger == nil {
		return
	}

	app.logger.Error(fmt.Sprintf(msg, args...))
}

func withTimeout(ctx context.Context, fn func(context.Context) error) error {
	ch := make(chan error, 1)
	go func() {
		ch <- fn(ctx)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-ch:
		return err
	}
}
