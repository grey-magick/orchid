package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/go-logr/logr"

	"github.com/isutton/orchid/pkg/server"

	"k8s.io/klog/klogr"
)

// main is server entrypoint.
func main() {
	// TODO: add cobra support
	ctx := context.TODO()
	logger := klogr.New().WithName("orchid")

	options := server.Options{
		Address: ":8080",
	}
	srv := server.NewServer(logger, options)

	logger.Info("Starting server")
	if err := srv.Start(ctx); err != nil {
		logger.Error(err, "An error happened while starting the server")
		os.Exit(1)
	}
	logger.Info("Server started")

	ShutdownOnInterrupt(logger, srv)
}

// ShutdownOnInterrupt waits for an interrupt signal to shutdown the server.
func ShutdownOnInterrupt(logger logr.Logger, srv *server.Server) {
	logger = logger.WithName("shutdownOnInterrupt")
	interruptChan := make(chan os.Signal, 1)
	doneChan := make(chan error, 1)

	// the pattern here is:
	// - register the interrupt channel to receive INT notifications
	// - spawn a go func to monitor the interrupt channel and shutdown the server
	// - block until the server has been finalized
	signal.Notify(interruptChan, os.Interrupt)
	go func() {
		<-interruptChan
		if err := srv.Shutdown(context.TODO()); err != nil {
			logger.Error(err, "An error happened while stopping the server")
		} else {
			logger.Info("Server stopped")
		}
		doneChan <- nil
	}()
	select {
	case <-doneChan:
		break
	}
}
