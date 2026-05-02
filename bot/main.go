package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

	u := bot.GetUpdatesChan(tgbotapi.NewUpdate(0))

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutting down bot...")
		os.Exit(0)
	}()

	for update := range u {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ParseMode = "Markdown"

		switch update.Message.Command() {
		case "start":
			msg.Text = "Привет! Я бот для анализа фехтовальщиков.\n\nИспользуй /help для списка команд."
		case "help":
			msg.Text = `📋 *Доступные команды:*

/start - Начать работу
/help - Показать это сообщение
/search [имя] - Поиск фехтовальщика
/profile [никнейм] - Профиль фехтовальщика

🔧 *Команды админа:*
/users - Список пользователей
/adduser - Добавить пользователя`
		case "search":
			msg.Text = "Поиск временно недоступен"
		case "users":
			msg.Text = "Список пользователей временно недоступен"
		default:
			msg.Text = "Неизвестная команда. Используй /help"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}

func fmtUser(u *tgbotapi.User) string {
	name := u.FirstName
	if u.LastName != "" {
		name += " " + u.LastName
	}
	if u.UserName != "" {
		name += fmt.Sprintf(" (@%s)", u.UserName)
	}
	return name
}
