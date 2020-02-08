package main

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/valve"
	"github.com/rs/zerolog/log"
	"media-roller/src/media"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	// Setup routes
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", media.Index)
		r.Get("/fetch", media.FetchMedia)
		r.Get("/download", media.ServeMedia)
	})

	// Print out all routes
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Info().Msgf("%s %s", method, route)
		return nil
	}
	// Panic if there is an error
	if err := chi.Walk(r, walkFunc); err != nil {
		log.Panic().Msgf("%s\n", err.Error())
	}

	valv := valve.New()
	baseCtx := valv.Context()
	srv := http.Server{Addr: ":3000", Handler: chi.ServerBaseContext(baseCtx, r)}

	// Create a shutdown hook for graceful shutdowns
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a ^C, handle it
			log.Info().Msgf("Shutting down...")

			// first valv
			_ = valv.Shutdown(20 * time.Second)

			// create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			// start http shutdown
			_ = srv.Shutdown(ctx)

			// verify, in worst case call cancel via defer
			select {
			case <-time.After(21 * time.Second):
				log.Error().Msgf("Not all connections done")
			case <-ctx.Done():
			}
		}
	}()

	// Start the listener
	err := srv.ListenAndServe()
	if err != nil {
		log.Info().Msg(err.Error())
	}
	log.Info().Msgf("Shutdown complete")
}
