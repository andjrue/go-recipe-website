package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"recipe-website/internal/application"
	"syscall"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := application.New()
	if err != nil {
		panic(err)
	}

	if err := app.Start(ctx); err != nil {
		panic(err)
	}

	fmt.Println("Application shut down gracefully")
}
