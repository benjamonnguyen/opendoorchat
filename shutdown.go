package opendoorchat

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type GracefulShutdownManager struct {
	handlers []func()
}

func (m *GracefulShutdownManager) ShutdownOnInterrupt(timeout time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, os.Interrupt)
	go func() {
		<-interruptSignal
		cancel()
	}()
	<-ctx.Done()

	start := time.Now()
	log.Info().Msg("starting graceful shutdown")
	errCh := make(chan error)
	go func() {
		for err := range errCh {
			if err != nil {
				log.Error().Err(err).Msg("failed graceful shutdown")
			}
		}
	}()

	var wg sync.WaitGroup
	done := make(chan struct{})
	add := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}
	for _, fn := range m.handlers {
		add(fn)
	}

	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		log.Info().Msgf("completed graceful shutdown after %s", time.Since(start))
	case <-time.After(timeout):
		log.Error().Msgf("graceful shutdown timed out after %s", timeout)
	}
}

func (m *GracefulShutdownManager) AddHandler(fn func()) {
	m.handlers = append(m.handlers, fn)
}
