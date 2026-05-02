package router

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/easy-tactics/bot/handlers"
	"github.com/easy-tactics/bot/middleware"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Router struct {
	bot    *tgbotapi.BotAPI
	authMw *middleware.AuthMiddleware
	api    *handlers.APIClient
}

func New(bot *tgbotapi.BotAPI, authMw *middleware.AuthMiddleware, api *handlers.APIClient) *Router {
	return &Router{
		bot:    bot,
		authMw: authMw,
		api:    api,
	}
}

func (r *Router) Handle(ctx context.Context, msg *tgbotapi.Message) {
	log.Printf("[%s] %s", msg.From.UserName, msg.Text)

	if msg.IsCommand() {
		r.handleCommand(ctx, msg)
		return
	}

	r.handleMessage(ctx, msg)
}

func (r *Router) handleCommand(ctx context.Context, msg *tgbotapi.Message) {
	role := r.authMw.GetUserRole(msg.From.ID)
	args := msg.CommandArguments()

	switch msg.Command() {
	case "start":
		r.handleStart(msg)
	case "help":
		r.handleHelp(msg, role)
	case "search":
		r.handleSearch(ctx, msg, args)
	case "profile":
		r.handleProfile(ctx, msg, args)
	case "tournaments":
		r.handleTournaments(ctx, msg, args)
	case "fights":
		r.handleFights(ctx, msg, args)
	case "users":
		r.handleUsers(ctx, msg, role)
	case "adduser":
		r.handleAddUser(ctx, msg, role, args)
	default:
		r.sendMessage(msg.Chat.ID, "Неизвестная команда. Используй /help")
	}
}

func (r *Router) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	r.sendMessage(msg.Chat.ID, "Я понимаю только команды. Используй /help")
}

func (r *Router) handleStart(msg *tgbotapi.Message) {
	text := "Привет! 🎯\n\nЯ бот для анализа HEMA фехтовальщиков.\n\nИспользуй /help для списка команд."
	r.sendMessage(msg.Chat.ID, text)
}

func (r *Router) handleHelp(msg *tgbotapi.Message, role string) {
	text := `📋 *Доступные команды:*

/start - Начать работу
/help - Показать это сообщение
/search [имя] - Поиск фехтовальщика
/profile [никнейм] - Профиль фехтовальщика
/tournaments [никнейм] - Турниры
/fights [никнейм] - Бои`

	if role == "owner" || role == "admin" {
		text += `

🔧 *Команды админа:*
/users - Список пользователей
/adduser - Добавить пользователя`
	}

	r.sendMarkdown(msg.Chat.ID, text)
}

func (r *Router) handleSearch(ctx context.Context, msg *tgbotapi.Message, args string) {
	if args == "" {
		r.sendMessage(msg.Chat.ID, "Укажи имя для поиска: /search [имя]")
		return
	}

	if r.api == nil {
		r.sendMessage(msg.Chat.ID, "API недоступен")
		return
	}

	result, err := r.api.SearchFighter(ctx, args)
	if err != nil {
		r.sendMessage(msg.Chat.ID, "Ошибка поиска")
		return
	}

	if len(result.Matches) == 0 {
		r.sendMessage(msg.Chat.ID, "Фехтовальщик не найден")
		return
	}

	text := "🔍 *Результаты поиска:*\n\n"
	for _, m := range result.Matches {
		text += fmt.Sprintf("• %s (%s)\n", m.FullName, m.Slug)
	}
	r.sendMarkdown(msg.Chat.ID, text)
}

func (r *Router) handleProfile(ctx context.Context, msg *tgbotapi.Message, args string) {
	slug := args
	if slug == "" {
		slug = msg.From.UserName
	}

	if r.api == nil {
		r.sendMessage(msg.Chat.ID, "API недоступен")
		return
	}

	fighter, err := r.api.GetFighterBySlug(ctx, slug)
	if err != nil || fighter == nil {
		r.sendMessage(msg.Chat.ID, "Фехтовальщик не найден")
		return
	}

	text := fmt.Sprintf("🥷 *%s*\n\n", fighter.FullName)
	if fighter.City != "" {
		text += fmt.Sprintf("🏙 %s\n", fighter.City)
	}
	if fighter.Club != "" {
		text += fmt.Sprintf("🏅 %s\n", fighter.Club)
	}
	if fighter.HemagonURL != "" {
		text += fmt.Sprintf("🔗 %s\n", fighter.HemagonURL)
	}

	r.sendMarkdown(msg.Chat.ID, text)
}

func (r *Router) handleTournaments(ctx context.Context, msg *tgbotapi.Message, args string) {
	slug := args
	if slug == "" {
		slug = msg.From.UserName
	}

	r.sendMessage(msg.Chat.ID, "Турниры временно недоступны")
}

func (r *Router) handleFights(ctx context.Context, msg *tgbotapi.Message, args string) {
	slug := args
	if slug == "" {
		slug = msg.From.UserName
	}

	r.sendMessage(msg.Chat.ID, "Бои временно недоступны")
}

func (r *Router) handleUsers(ctx context.Context, msg *tgbotapi.Message, role string) {
	if role != "owner" && role != "admin" {
		r.sendMessage(msg.Chat.ID, "Нет доступа")
		return
	}

	if r.api == nil {
		r.sendMessage(msg.Chat.ID, "API недоступен")
		return
	}

	users, err := r.api.ListUsers(ctx)
	if err != nil || len(users) == 0 {
		r.sendMessage(msg.Chat.ID, "Нет пользователей")
		return
	}

	text := "👥 *Пользователи:*\n\n"
	for _, u := range users {
		text += fmt.Sprintf("• %s (%s) - %s\n", u.FullName, u.Username, u.Role)
	}

	r.sendMarkdown(msg.Chat.ID, text)
}

func (r *Router) handleAddUser(ctx context.Context, msg *tgbotapi.Message, role string, args string) {
	if role != "owner" && role != "admin" {
		r.sendMessage(msg.Chat.ID, "Нет доступа")
		return
	}

	if args == "" {
		r.sendMessage(msg.Chat.ID, "Формат: /adduser [telegram_id] [username] [full_name] [role]\nПример: /adduser 12345678 john John Doe fighter")
		return
	}

	parts := strings.Fields(args)
	if len(parts) < 4 {
		r.sendMessage(msg.Chat.ID, "Недостаточно аргументов.\nФормат: /adduser [telegram_id] [username] [full_name] [role]")
		return
	}

	telegramID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		r.sendMessage(msg.Chat.ID, "Неверный telegram_id")
		return
	}

	username := parts[1]
	fullName := parts[2]
	userRole := parts[3]

	validRoles := map[string]bool{"owner": true, "admin": true, "coach": true, "fighter": true}
	if !validRoles[userRole] {
		r.sendMessage(msg.Chat.ID, "Неверная роль. Используй: owner, admin, coach, fighter")
		return
	}

	if r.api == nil {
		r.sendMessage(msg.Chat.ID, "API недоступен")
		return
	}

	err = r.api.AddUser(ctx, telegramID, username, fullName, userRole)
	if err != nil {
		r.sendMessage(msg.Chat.ID, fmt.Sprintf("Ошибка: %v", err))
		return
	}

	r.sendMessage(msg.Chat.ID, fmt.Sprintf("Пользователь %s добавлен с ролью %s", username, userRole))
}

func (r *Router) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	r.bot.Send(msg)
}

func (r *Router) sendMarkdown(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	r.bot.Send(msg)
}
