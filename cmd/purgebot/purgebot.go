package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/glebarez/sqlite"
	"github.com/keshon/purge-bot/internal/bot"
	"github.com/keshon/purge-bot/internal/config"
	"github.com/keshon/purge-bot/internal/health"
	"github.com/keshon/purge-bot/internal/logutil"
	"github.com/keshon/purge-bot/internal/version"
	"gorm.io/gorm"
)

func healthSocketPath() string {
	if p := os.Getenv("HEALTH_SOCKET"); p != "" {
		return p
	}
	return health.DefaultSocketPath
}

// discordgoAdapter adapts discordgo.Session to bot.MessageFetcherDeleter interface.
type discordgoAdapter struct {
	session *discordgo.Session
}

func (a *discordgoAdapter) ChannelMessages(channelID string, limit int, beforeID, afterID, aroundID string) ([]*discordgo.Message, error) {
	return a.session.ChannelMessages(channelID, limit, beforeID, afterID, aroundID)
}

func (a *discordgoAdapter) ChannelMessageDelete(channelID, msgID string) error {
	return a.session.ChannelMessageDelete(channelID, msgID)
}

func main() {
	ver := flag.Bool("version", false, "print version")
	envPath := flag.String("env", "", "Path to .env file (empty = load from current working directory)")
	dbPath := flag.String("db", "database.db", "Path to database file")
	logLevel := flag.String("log-level", "info", "Log level: debug, info, warn, error")
	logFormat := flag.String("log-format", "text", "Log format: text or json")
	flag.Parse()

	if *ver {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	// Healthcheck subcommand: ping the running purgebot via Unix socket and exit 0/1.
	if len(flag.Args()) > 0 && flag.Arg(0) == "healthcheck" {
		if health.Ping(healthSocketPath()) {
			os.Exit(0)
		}
		os.Exit(1)
	}

	level := logutil.ParseLogLevel(*logLevel)
	format := strings.ToLower(strings.TrimSpace(*logFormat))
	if format != "json" {
		format = "text"
	}

	// Normal logs to stdout; errors to stderr (Docker/container-friendly).
	handler := logutil.NewSplitHandler(os.Stdout, os.Stderr, level, format)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	cfg, err := config.Load(*envPath)
	if err != nil {
		slog.Error("loading config", "error", err)
		os.Exit(1)
	}

	// CLI flag overrides env/config value
	cfg.DBPath = *dbPath

	slog.Info("starting", "config", cfg.LogSafe(), "log_level", *logLevel, "log_format", *logFormat)

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		slog.Error("opening database", "error", err)
		os.Exit(1)
	}

	err = db.AutoMigrate(&bot.Task{}, &bot.ThreadCleanupTask{}, &bot.UserPermission{}, &bot.RolePermission{})
	if err != nil {
		slog.Error("migrating database", "error", err)
		os.Exit(1)
	}

	dg, err := discordgo.New("Bot " + cfg.DiscordKey)
	if err != nil {
		slog.Error("creating Discord session", "error", err)
		os.Exit(1)
	}

	adapter := &discordgoAdapter{session: dg}
	b := bot.NewBot(db, adapter)
	b.SetSession(dg)
	b.SetLogger(logger)

	const (
		heartbeatInterval = 60 * time.Second
		maxIdle           = 150 * time.Second // 2.5 min; must be > heartbeatInterval
	)
	var ready atomic.Bool
	var lastActivityAt atomic.Int64
	b.SetOnReady(func() {
		ready.Store(true)
		lastActivityAt.Store(time.Now().UnixNano())
	})
	isReady := func() bool {
		return ready.Load() && time.Since(time.Unix(0, lastActivityAt.Load())) < maxIdle
	}
	socketPath := healthSocketPath()
	healthSock := health.NewSocketServer(socketPath, isReady)
	go healthSock.Run()
	slog.Info("health check socket listening", "path", socketPath)

	dg.AddHandler(b.Ready)
	dg.AddHandler(b.MessageCreate)
	dg.AddHandler(func(_ *discordgo.Session, _ *discordgo.Resumed) {
		ready.Store(true)
		lastActivityAt.Store(time.Now().UnixNano())
	})

	healthCtx, healthCancel := context.WithCancel(context.Background())
	defer healthCancel()
	go func() {
		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-healthCtx.Done():
				return
			case <-ticker.C:
				if _, err := dg.User("@me"); err == nil {
					lastActivityAt.Store(time.Now().UnixNano())
				}
			}
		}
	}()

	if err := dg.Open(); err != nil {
		slog.Error("opening Discord session", "error", err)
		os.Exit(1)
	}

	slog.Info("bot running", "purge_interval", "33s", "min_duration", "30s", "max_duration", "3333d")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	slog.Info("shutting down gracefully")
	b.Stop()
	healthCancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	_ = healthSock.Shutdown(shutdownCtx)
	shutdownCancel()
	if err := dg.Close(); err != nil {
		slog.Error("closing Discord session", "error", err)
	}
}
