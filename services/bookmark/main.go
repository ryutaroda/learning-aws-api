package main

import (
    "log"

    "bookmark/config"
    "bookmark/handler"
    "bookmark/pkg/database"
    "bookmark/repository"
    "bookmark/service"
)

func main() {
    // 設定読み込み
    cfg := config.Load()

    // DB接続
    db := database.Connect(cfg.DatabaseURL)

    // Repository初期化
    bookmarkRepo := repository.NewBookmarkRepository(db)

    // Service初期化
    bookmarkService := service.NewBookmarkService(bookmarkRepo)

    // Router初期化
    router := handler.NewRouter(bookmarkService)

    // サーバー起動
    log.Println("Starting server on :8080")
    if err := router.Run(":8080"); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}