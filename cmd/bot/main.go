package main

import (
    "context"
    "log"
    "time"

    "github.com/ziyodbekabdunasimov/chiptatop-bot/internal/bot"
    "github.com/ziyodbekabdunasimov/chiptatop-bot/internal/config"
)

func main() {
    cfg := config.Load()

    b, err := bot.New(cfg)
    if err != nil {
        log.Fatalf("failed to create bot: %v", err)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    if err := b.Run(ctx); err != nil {
        // Avoid noisy error on normal shutdown
        if err.Error() != context.Canceled.Error() {
            log.Printf("bot stopped: %v", err)
        }
    }

    // Give some time for pending operations
    time.Sleep(200 * time.Millisecond)
}


