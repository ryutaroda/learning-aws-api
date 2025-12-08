package main

import (
    "flag"
    "fmt"
    "log"
    "os"

    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
    // コマンドライン引数の定義
    var (
        migrationsPath = flag.String("path", "db/mydb/migrations", "マイグレーションファイルのパス")
        databaseURL    = flag.String("database", "", "データベース接続URL")
        command        = flag.String("cmd", "up", "実行するコマンド (up/down/version)")
        steps          = flag.Int("steps", -1, "マイグレーションのステップ数（-1で全て）")
    )
    flag.Parse()

    // 環境変数からデータベースURLを取得（フラグが指定されていない場合）
    if *databaseURL == "" {
        *databaseURL = os.Getenv("DATABASE_URL")
        if *databaseURL == "" {
            *databaseURL = "postgresql://postgres:postgres@localhost:5432/bookmark_dev?sslmode=disable"
        }
    }

    // マイグレーションインスタンスの作成
    m, err := migrate.New(
        fmt.Sprintf("file://%s", *migrationsPath),
        *databaseURL,
    )
    if err != nil {
        log.Fatalf("マイグレーションの初期化に失敗: %v", err)
    }
    defer m.Close()

    // コマンドの実行
    switch *command {
    case "up":
        if *steps < 0 {
            err = m.Up()
        } else {
            err = m.Steps(*steps)
        }
    case "down":
        if *steps < 0 {
            err = m.Down()
        } else {
            err = m.Steps(-*steps)
        }
    case "version":
        version, dirty, verr := m.Version()
        if verr != nil {
            log.Printf("バージョン取得エラー: %v", verr)
            return
        }
        log.Printf("現在のバージョン: %d (dirty: %v)", version, dirty)
        return
    default:
        log.Fatalf("不明なコマンド: %s (up/down/version のいずれかを指定)", *command)
    }

    // エラーハンドリング
    if err != nil {
        if err == migrate.ErrNoChange {
            log.Println("マイグレーションの変更はありません")
        } else {
            log.Fatalf("マイグレーション実行エラー: %v", err)
        }
    } else {
        log.Println("マイグレーションが正常に完了しました")
    }
}