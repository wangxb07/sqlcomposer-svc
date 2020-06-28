package main

//go:generate sqlboiler mysql

import (
	"context"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jessevdk/go-flags"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"github.com/user/sqlcomposer-svc/restapi"
	"github.com/user/sqlcomposer-svc/restapi/v1"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ServerConfig struct {
	Host string `long:"host" description:"the IP to listen on" default:"localhost" env:"HOST"`
	Port int    `long:"port" description:"the port to listen on for insecure connections" default:"8080" env:"PORT"`
	DB   string `long:"db" description:"the database connection dns string" env:"DB"`
}

func main() {
	cfg := new(ServerConfig)

	log.SetLevel(log.DebugLevel)
	// flags parser config
	parser := flags.NewParser(cfg, flags.Default)

	parser.ShortDescription = "SQL Composer API"
	parser.LongDescription = "SQL Composer Management API"

	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}

	db := sqlx.MustConnect("mysql", cfg.DB)
	defer db.Close()

	v1.Setup(&v1.Config{
		DB: db,
	})

	restapi.Setup(&restapi.Config{
		DB: db,
	})

	defer v1.Destroy()

	router := restapi.InitRoutes()

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: router,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Info("Server exiting")
}
