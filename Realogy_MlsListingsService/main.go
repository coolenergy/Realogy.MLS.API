package main

import (
	"context"
	"fmt"
	"mlslisting/internal/config"
	"mlslisting/internal/gateway"
	"mlslisting/internal/server"
	"os"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var conf *config.Config
	if len(os.Args) > 1 && os.Args[1] != "" {
		conf = config.Load(os.Args[1])
	} else {
		conf = config.Load("configs")
	}

	grpc := server.Server{Config: conf}
	defer grpc.Cleanup(ctx)

	ch := make(chan error, 2)

	// start grpc server
	go func() {
		if err := grpc.Start(ctx); err != nil {
			ch <- fmt.Errorf("error while starting grpc server: %v", err)
		}
	}()

	// start gateway server
	go func() {
		if err := gateway.Start(ctx, *grpc.Config); err != nil {
			ch <- fmt.Errorf("error while starting gateway server: %v", err)
		}
	}()

	select {
	case err := <-ch:
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
