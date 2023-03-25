package main

import (
	"context"
	"documentize"
	"documentize/database"
	"github.com/caarlos0/env/v6"
	//"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/zeebo/errs"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	onSigInt(func() {
		// starting graceful exit on context cancellation.
		cancel()
	})

	config := new(documentize.Config)
	err := env.Parse(config)
	if err != nil {
		log.Println("could not parse env to config:", err)
		return
	}

	db, err := database.New(config.DatabaseURL)
	if err != nil {
		log.Println("Error starting master database", err)
		return
	}
	defer func() {
		err = errs.Combine(err, db.Close())
		log.Println("close db error", err)
	}()

	err = db.CreateSchema(ctx)
	if err != nil {
		log.Println("Error creating schema", err)
		return
	}

	app, err := documentize.New(config, db)
	if err != nil {
		log.Println("could not initialize app:", err)
		return
	}

	err = errs.Combine(app.Run(ctx), app.Close())
	log.Println(err)
}

// OnSigInt fires in SIGINT or SIGTERM event (usually CTRL+C).
func onSigInt(onSigInt func()) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-done
		onSigInt()
	}()
}
