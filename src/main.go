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
	"path"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// Setup routes
	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		router.Get("/", media.Index)
		router.Get("/fetch", media.FetchMedia)
		router.Get("/download", media.ServeMedia)
	})
	fileServer(router, "/static", "static/")

	// Print out all routes
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Info().Msgf("%s %s", method, route)
		return nil
	}
	// Panic if there is an error
	if err := chi.Walk(router, walkFunc); err != nil {
		log.Panic().Msgf("%s\n", err.Error())
	}

	valv := valve.New()
	baseCtx := valv.Context()
	srv := http.Server{Addr: ":3000", Handler: chi.ServerBaseContext(baseCtx, router)}

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

	err := srv.ListenAndServe()
	if err != nil {
		log.Info().Msg(err.Error())
	}
	log.Info().Msgf("Shutdown complete")
}

func fileServer(r chi.Router, public string, static string) {
	if strings.ContainsAny(public, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	root, _ := filepath.Abs(static)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		panic("Static Documents Directory Not Found")
	}

	fs := http.StripPrefix(public, http.FileServer(http.Dir(root)))

	if public != "/" && public[len(public)-1] != '/' {
		r.Get(public, http.RedirectHandler(public+"/", 301).ServeHTTP)
		public += "/"
	}

	r.Get(public+"*", func(w http.ResponseWriter, r *http.Request) {
		file := strings.Replace(r.RequestURI, public, "/", 1)
		if _, err := os.Stat(root + file); os.IsNotExist(err) {
			http.ServeFile(w, r, path.Join(root, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	})
}
