package bot

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/ziyodbekabdunasimov/chiptatop-bot/internal/config"
)

type Bot struct {
    api *tgbotapi.BotAPI
    cfg config.Config
}

func New(cfg config.Config) (*Bot, error) {
    api, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
    if err != nil {
        return nil, err
    }
    return &Bot{api: api, cfg: cfg}, nil
}

func (b *Bot) Run(ctx context.Context) error {
    u := tgbotapi.NewUpdate(0)
    u.Timeout = 30

    updates := b.api.GetUpdatesChan(u)

    // Graceful shutdown handling
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-stop:
            log.Println("Shutting down bot...")
            return nil
        case update := <-updates:
            if update.Message == nil {
                continue
            }
            if update.Message.IsCommand() {
                b.handleCommand(update)
                continue
            }
        }
    }
}

func (b *Bot) handleCommand(update tgbotapi.Update) {
    switch update.Message.Command() {
    case "start":
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome! I will notify you instantly when tickets become available for trains, buses, and flights. Use /help to see what I can do.")
        msg.DisableWebPagePreview = true
        b.safeSend(msg)
    case "help":
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Available commands:\n/start - Start the bot\n/help - Show help")
        b.safeSend(msg)
    default:
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command. Try /help")
        b.safeSend(msg)
    }
}

func (b *Bot) safeSend(msg tgbotapi.MessageConfig) {
    if _, err := b.api.Send(msg); err != nil {
        log.Printf("send error: %v", err)
        time.Sleep(200 * time.Millisecond)
    }
}


