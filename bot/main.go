package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/easy-tactics/bot/handlers"
	"github.com/easy-tactics/bot/middleware"
	"github.com/easy-tactics/bot/router"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Config struct {
	BotToken        string
	APIAddr         string
	OwnerTelegramID string
}

func Load() Config {
	return Config{
		BotToken:        getEnv("TELEGRAM_BOT_TOKEN", ""),
		APIAddr:         getEnv("API_ADDR", "localhost:50051"),
		OwnerTelegramID: getEnv("OWNER_TELEGRAM_ID", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	cfg := Load()
	if cfg.BotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Printf("Bot authorized as %s", bot.Self.UserName)

	apiClient, err := handlers.NewAPIClient(cfg.APIAddr)
	if err != nil {
		log.Printf("Warning: failed to connect to API: %v", err)
	}

	authMw := middleware.NewAuthMiddleware(apiClient, cfg.OwnerTelegramID)
	cmdRouter := router.New(bot, authMw, apiClient)

	u := bot.GetUpdatesChan(tgbotapi.NewUpdate(0))

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutting down bot...")
		if apiClient != nil {
			apiClient.Close()
		}
		os.Exit(0)
	}()

	log.Println("Bot is running...")

	for update := range u {
		if update.Message == nil {
			continue
		}

		ctx := context.Background()
		cmdRouter.Handle(ctx, update.Message)
	}
}
