package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"chi-deutschland.com/ecommerce-one-record-converter/cmd/converter/config"
	"chi-deutschland.com/ecommerce-one-record-converter/pkg/neone"
	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/ratelimiter"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
	"github.com/failsafe-go/failsafe-go/timeout"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	setUpLogging()

	cfg, err := config.Load()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to load configuration, using defaults")

		cfg = config.Default()
	}

	neoneServer := setUpNeOneServerConn(cfg.NEONE)
	mux := setUpMux(cfg.HTTP, neoneServer)

	const (
		readTimeout       = 10 * time.Second
		readHeaderTimeout = 5 * time.Second
		writeTimeout      = 15 * time.Second
		maxHeaderBytes    = 1 << 20 // 1 MB
	)

	server := http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           mux,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
		BaseContext:       func(net.Listener) context.Context { return context.Background() },
	}

	log.Debug().Str("addr", server.Addr).Msg("Starting HTTP server")

	err = server.ListenAndServe()
	if err != nil {
		log.WithLevel(zerolog.FatalLevel).Err(err).Msg("HTTP Listen and Serve failed")

		return
	}
}

func setUpLogging() {
	const timeFormat = "2006-01-02T15:04:05.000Z07:00"

	zerolog.TimeFieldFormat = time.RFC3339Nano

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: timeFormat,
	})
}

func setUpNeOneServerConn(cfg config.NEONE) *neone.Client {
	timeoutPolicy := timeout.NewBuilder[*http.Response](cfg.RequestTimeout).Build()

	rateLimitPolicy := ratelimiter.NewSmoothBuilder[*http.Response](
		cfg.RateLimiterPolicy.MaxExecutionsPerMinute, time.Minute).
		OnRateLimitExceeded(func(failsafe.ExecutionEvent[*http.Response]) {
			log.Warn().Msg("rate limit exceeded for NE-ONE Server requests, waiting to retry...")
		}).
		WithMaxWaitTime(cfg.RateLimiterPolicy.MaxWaitTime).
		Build()

	retryPolicy := retrypolicy.NewBuilder[*http.Response]().
		HandleIf(func(response *http.Response, err error) bool {
			return err != nil || response.StatusCode >= http.StatusInternalServerError
		}).
		OnRetry(func(failsafe.ExecutionEvent[*http.Response]) {
			log.Warn().Msg("about to retry request to NE-ONE Server.")
		}).
		WithMaxAttempts(cfg.RetryPolicy.MaxAttempts).
		WithBackoff(cfg.RetryPolicy.Delay, cfg.RetryPolicy.MaxDelay).
		Build()

	return neone.NewServer(http.DefaultClient, timeoutPolicy, rateLimitPolicy, retryPolicy)
}

func setUpMux(cfg config.HTTP, server *neone.Client) *http.ServeMux {
	neoneForwarder := NewNeoneDataForwarder(server)

	mux := http.NewServeMux()

	mux.Handle("POST /upload", logMiddleware(neoneForwarder))

	fileServer := http.FileServer(http.Dir(cfg.StaticFilesDir))

	mux.Handle("/", http.StripPrefix("/", fileServer)) // not logging static file requests, too many entries

	logStaticFileInfo(cfg.StaticFilesDir)

	return mux
}

// logStaticFileInfo logs the names of the root static HTML files in the provided
// directory. This is useful for troubleshooting issues with static file serving.
func logStaticFileInfo(staticFilesDir string) {
	dirEntries, err := os.ReadDir(staticFilesDir)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to read static files directory, static file serving may not work")

		return
	}

	for _, entry := range dirEntries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}

		log.Debug().Str("name", entry.Name()).Msg("Root static HTML file entry found")
	}
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Str("method", r.Method).Str("url", r.URL.String()).Msg("HTTP request")
		next.ServeHTTP(w, r)
	})
}
