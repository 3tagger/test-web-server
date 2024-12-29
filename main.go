package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type httpResponse struct {
	Message string `json:"message"`
}

type flags struct {
	Port string
	ID   string
}

const (
	logTagAddr  = "addr"
	logTagID    = "id"
	logTagError = "error"
)

func parseFlags() *flags {
	id := flag.String("id", "id", "the ID to display when calling the index route")
	port := flag.String("port", "80", "the port used for the HTTP server")
	flag.Parse()

	return &flags{
		ID:   *id,
		Port: *port,
	}
}

func getAddr(flgs *flags) string {
	return fmt.Sprintf(":%s", flgs.Port)
}

func createServer(flgs *flags) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(httpResponse{
			Message: fmt.Sprintf("hello world - %s", flgs.ID),
		})
	})

	srv := &http.Server{
		Addr:    getAddr(flgs),
		Handler: mux,
	}

	return srv
}

func main() {
	flgs := parseFlags()

	srv := createServer(flgs)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	go func(flgs *flags) {
		slog.Info(fmt.Sprintf("server running on %s", getAddr(flgs)), logTagID, flgs.ID)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("error .ListenAndServe", logTagError, err.Error(), logTagID, flgs.ID)
		}
	}(flgs)

	<-ch

	graceCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := srv.Shutdown(graceCtx); err != nil {
		slog.Error("error .Shutdown", logTagError, err.Error())
	}

	slog.Info("server stopping", logTagID, flgs.ID)
}
