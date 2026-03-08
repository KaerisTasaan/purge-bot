package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/glebarez/sqlite"
	"github.com/keshon/purge-bot/internal/bot"
	"github.com/keshon/purge-bot/internal/config"
	"github.com/keshon/purge-bot/internal/logutil"
	"github.com/keshon/purge-bot/internal/version"
	"gorm.io/gorm"
)

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

	dg.AddHandler(b.Ready)
	dg.AddHandler(b.MessageCreate)

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
	if err := dg.Close(); err != nil {
		slog.Error("closing Discord session", "error", err)
	}
}
