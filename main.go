package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"gitlab.meitu.com/xmqibu/internal/config"
	"gitlab.meitu.com/xmqibu/internal/site"
	"gitlab.meitu.com/xmqibu/internal/store"
	"gitlab.meitu.com/xmqibu/internal/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	repo := store.NewMySQLStore(db)
	if err := repo.EnsureSchema(ctx); err != nil {
		log.Fatal(err)
	}

	templates, err := web.ParseTemplates()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(cfg.MediaDir, 0o755); err != nil {
		log.Fatalf("create media dir: %v", err)
	}

	service := site.NewService(repo)
	app := web.NewApp(templates, cfg.IngestToken, cfg.MediaDir, service, service)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           app.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	waitForShutdown(server)
}

func waitForShutdown(server *http.Server) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	<-signals

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
