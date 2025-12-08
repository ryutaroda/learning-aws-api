# Go三層アーキテクチャ プロンプト（ブックマーク管理サービス）

このドキュメントは、Go + Gin + SQS + ECS を使用した三層アーキテクチャのバックエンドアプリケーションを新規作成するためのプロンプトです。

**学習目的**: 三層アーキテクチャで「痛み」を経験しながら、将来的なDDD移行の動機を理解する。

---

## 概要

**ブックマーク管理サービス**を構築します：

- **APIサーバー**: Gin フレームワーク（URL登録・一覧・検索）
- **ワーカー**: SQSからURL受信 → OGP情報を非同期取得
- **バッチ**: 古いOGP情報を定期更新・リンク切れチェック
- **DBマイグレーション**: golang-migrate

同一イメージを環境変数（MODE）で切り替えて運用します。

### 機能概要

```
┌─────────────────────────────────────────────────────────────┐
│  POST /api/bookmarks        → URLを登録（即座に202返却）    │
│  GET  /api/bookmarks        → 一覧取得                      │
│  GET  /api/bookmarks/:id    → 詳細取得                      │
│  DELETE /api/bookmarks/:id  → 削除                          │
│  GET  /api/bookmarks/search → タイトル・タグで検索          │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  【Worker】 SQSからURL受信 → OGP取得 → DB更新               │
│  - title, description, image, favicon を取得               │
│  - 取得失敗 → リトライ or DLQ                               │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  【Batch】 7日以上古いOGP情報を再取得                       │
│  - リンク切れチェック                                       │
│  - 404 → ステータス更新                                     │
└─────────────────────────────────────────────────────────────┘
```

---

## 三層アーキテクチャとは

```
┌─────────────────────────────────────────────┐
│         Presentation Layer (handler/)       │  ← HTTP/CLI/SQS受信
├─────────────────────────────────────────────┤
│         Business Logic Layer (service/)     │  ← ビジネスロジック
├─────────────────────────────────────────────┤
│         Data Access Layer (repository/)     │  ← DB/外部API/SQS送信
└─────────────────────────────────────────────┘
```

| 層 | 責務 | 依存先 |
|----|------|--------|
| Presentation | リクエスト受信・レスポンス返却 | Service |
| Business Logic | ビジネスルール・ユースケース | Repository |
| Data Access | データ永続化・外部連携 | なし |

**特徴**: クリーンアーキテクチャより**シンプル**で、小〜中規模プロジェクトに最適。

---

## リポジトリ構成: モノリポ + マイクロサービス

本プロジェクトは**モノリポ（Monorepo）+ マイクロサービス**構成を採用します。

```
┌─────────────────────────────────────────────────────────────┐
│  learning-aws-api（1つのリポジトリ）                        │
│                                                             │
│  services/                                                  │
│  ├── bookmark/    → bookmarkコンテナ（ECS Service）        │
│  ├── user/        → userコンテナ（ECS Service）将来追加    │
│  └── notification/→ notificationコンテナ 将来追加          │
│                                                             │
│  各サービスは独立したコンテナとしてデプロイ                 │
└─────────────────────────────────────────────────────────────┘
```

### メリット

| 項目 | 説明 |
|------|------|
| コード共有 | 共通ライブラリを`shared/`で管理可能 |
| リファクタリング | 横断的な変更が1つのPRで完結 |
| CI/CD | 統一されたパイプラインで管理 |
| 依存関係 | go.workで複数モジュールを統合可能 |

### デプロイ戦略

| パターン | 説明 |
|---------|------|
| 全サービス同時 | `make deploy-all` |
| 変更サービスのみ | CI/CDで変更検知 → 該当サービスのみデプロイ |

---

## ディレクトリ構成

```
learning-aws-api/                   # モノリポ
├── services/
│   ├── bookmark/                   # ブックマーク管理サービス（マイクロサービス1）
│       ├── main.go                 # エントリーポイント（MODE切り替え）
│       ├── go.mod
│       ├── Dockerfile
│       ├── Makefile
│       │
│       ├── config/
│       │   ├── config.go           # 設定読み込み
│       │   └── config.toml.tmpl    # 設定テンプレート
│       │
│       ├── model/                  # エンティティ・DTO
│       │   ├── bookmark.go        # Bookmarkモデル
│       │   └── errors.go           # エラー定義
│       │
│       ├── handler/                # Presentation Layer
│       │   ├── router.go           # Gin初期化・ルーティング
│       │   ├── bookmark.go        # ブックマークAPIハンドラー
│       │   ├── health.go           # ヘルスチェック
│       │   ├── worker.go           # SQSワーカー（OGP取得）
│       │   └── batch.go            # バッチ処理（OGP更新）
│       │
│       ├── service/                # Business Logic Layer
│       │   ├── bookmark.go        # ブックマークサービス
│       │   └── ogp.go              # OGP取得サービス（痛みポイント1）
│       │
│       ├── repository/             # Data Access Layer
│       │   ├── bookmark.go        # ブックマークリポジトリ（DB）
│       │   └── queue.go            # キューリポジトリ（SQS）
│       │
│       └── pkg/                    # サービス内共通ユーティリティ
│           ├── database/
│           │   └── postgres.go     # DB接続
│           ├── sqs/
│           │   └── client.go       # SQSクライアント
│           └── http/
│               └── client.go       # HTTPクライアント（OGP取得用）
│
│   ├── user/                       # ユーザー管理サービス（将来追加の例）
│   │   ├── main.go
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   ├── model/
│   │   ├── handler/
│   │   ├── service/
│   │   └── repository/
│   │
│   └── notification/               # 通知サービス（将来追加の例）
│       ├── main.go
│       └── ...
│
├── shared/                         # 複数サービス間で共有するコード（将来）
│   └── pkg/
│       ├── logger/                 # 共通ロガー
│       └── middleware/             # 共通ミドルウェア
│
├── ops/
│   ├── db-migrator/                # DBマイグレーション
│   │   ├── main.go
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   ├── Makefile
│   │   └── db/
│   │       └── mydb/
│   │           └── migrations/
│   │               ├── 000001_create_bookmarks.up.sql
│   │               └── 000001_create_bookmarks.down.sql
│   │
│   └── ecspresso/                  # ECSデプロイ設定（サービスごとに分離）
│       ├── bookmark-api/           # bookmark APIサービス用
│       │   └── stg/
│       │       ├── ecspresso.yml
│       │       ├── ecs-task-def.json
│       │       └── ecs-service-def.json
│       ├── bookmark-worker/        # bookmark Workerサービス用
│       │   └── stg/
│       │       ├── ecspresso.yml
│       │       ├── ecs-task-def.json
│       │       └── ecs-service-def.json
│       ├── bookmark-batch/         # bookmark Batchサービス用
│       │   └── stg/
│       │       ├── ecspresso.yml
│       │       ├── ecs-task-def.json
│       │       └── ecs-task-def.overrides.json
│       ├── user-api/               # userサービス用（将来追加）
│       │   └── stg/
│       │       └── ...
│       └── db-migrator/
│           └── stg/
│               ├── ecspresso.yml
│               └── ecs-task-def.json
│
├── docs/                           # ドキュメント
│   ├── architecture/
│   │   └── go-three-layer-architecture.md
│   └── implementation-guide.md
│
├── .github/
│   └── workflows/
│       └── deploy.yaml             # CI/CD（サービスごとにデプロイ制御）
│
├── go.work                         # 複数モジュール統合（オプション）
├── Makefile                        # ルートMakefile（全サービス管理）
└── README.md
```

### go.work（複数モジュール統合）

`go.work`を使うと、複数の`go.mod`を持つサービスを統合して開発できます：

```go
// go.work
go 1.24

use (
    ./services/bookmark
    ./services/user
    ./shared
)
```

### ルートMakefile例

```makefile
# Makefile
.PHONY: build-all deploy-bookmark deploy-all

# 全サービスビルド
build-all:
	cd services/bookmark && go build -o bin/server .
	cd services/user && go build -o bin/server .

# bookmarkサービスのみデプロイ
deploy-bookmark:
	cd ops/ecspresso/bookmark-api/stg && ecspresso deploy

# 全サービスデプロイ
deploy-all:
	$(MAKE) deploy-bookmark
	# $(MAKE) deploy-user  # 将来追加
```

---

## クリーンアーキテクチャとの比較

| 項目 | 三層アーキテクチャ | クリーンアーキテクチャ |
|------|-------------------|----------------------|
| 層の数 | 3層 | 4層（domain/application/infrastructure/interface） |
| インターフェース | なし（直接依存） | リポジトリインターフェース必須 |
| 複雑さ | 低い | 高い |
| テスト容易性 | 中程度 | 高い |
| 適用規模 | 小〜中規模 | 中〜大規模 |
| 学習コスト | 低い | 高い |

---

## 技術スタック

| カテゴリ | 技術 |
|----------|------|
| 言語 | Go 1.24+ |
| Webフレームワーク | Gin |
| ORM | GORM |
| DB | PostgreSQL（配列型・GINインデックス使用） |
| キュー | Amazon SQS |
| HTTPクライアント | resty（OGP取得用） |
| HTMLパース | goquery（OGP抽出用） |
| マイグレーション | golang-migrate |
| コンテナ | Docker + ECS Fargate |
| デプロイ | ecspresso |
| CI/CD | GitHub Actions |

## 学習過程で経験する「痛みポイント」

### 痛み1: ロジックの置き場所問題
```
Q: OGP取得ロジックはどこに書く？
  - service/bookmark.go？
  - 別ファイル service/ogp.go？
  - Worker専用のロジック？

→ DDDなら「OgpFetcher」をドメインサービスとして定義
```

### 痛み2: モデルの肥大化
```
Q: Bookmark構造体が大きくなりすぎる
  - リクエストDTO
  - レスポンスDTO
  - DBモデル
  - SQSメッセージ
  全部同じ構造体で良い？

→ DDDならEntity/ValueObject/DTOを明確に分離
```

### 痛み3: 外部依存の変更
```
Q: OGP取得のHTTPクライアントを差し替えたい
  - テスト時はモック
  - 本番はresty
  
→ DDDならRepository Interfaceで抽象化
```

### 痛み4: バリデーションの置き場所
```
Q: URLの形式チェックはどこ？
  - handler?
  - service?
  
→ DDDならValueObject（URL型）を作ってそこで検証
```

---

## 実行モード

| MODE | 用途 | 実行方法 |
|------|------|----------|
| (default) | APIサーバー | ECS Service（`bookmark-api-stg`） |
| sqs | SQSワーカー | ECS Service（`bookmark-worker-stg`） |
| batch | バッチ処理 | ECS Task（`bookmark-batch-stg`） / EventBridge Scheduler |

**重要**: 各モードは**独立したECSサービス**としてデプロイします（サイドカーパターンではありません）。

### ECS構成

```
┌─────────────────────────────────────────────────────────────┐
│  ECSクラスター: learning-cluster-stg                        │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ ECS Service: bookmark-api-stg                     │   │
│  │   └── Task Definition: bookmark-api-task-stg      │   │
│  │       └── Container: api (MODE=api)               │   │
│  │           └── ALB → Target Group → Port 8080     │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ ECS Service: bookmark-worker-stg                  │   │
│  │   └── Task Definition: bookmark-worker-task-stg  │   │
│  │       └── Container: worker (MODE=sqs)          │   │
│  │           └── SQS Queue → メッセージ処理           │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ ECS Service: bookmark-batch-stg                   │   │
│  │   └── Task Definition: bookmark-batch-task-stg    │   │
│  │       └── Container: batch (MODE=batch)           │   │
│  │           └── EventBridge → 定期実行               │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### メリット

| 項目 | 説明 |
|------|------|
| **独立スケーリング** | API/Worker/Batchを個別にスケール可能 |
| **コスト最適化** | 用途別にリソースを最適化（API: 0.5 vCPU, Worker: 0.25 vCPU） |
| **障害分離** | WorkerがクラッシュしてもAPIは影響なし |
| **デプロイ柔軟性** | 各サービスを個別にデプロイ可能 |
| **同じイメージ** | 1つのDockerイメージを`MODE`環境変数で切り替え |

---

## 実装の流れ（Phase別）

### Phase 1: 基本CRUD（1週間）
```
- POST/GET/DELETE 実装
- PostgreSQL接続
- ECSデプロイ
```

### Phase 2: Worker追加（1週間）
```
- SQS連携
- OGP取得ロジック
- ここで「痛み1」を経験
```

### Phase 3: Batch + 検索（1週間）
```
- 定期更新バッチ
- タグ検索
- ここで「痛み2」を経験
```

### Phase 4: リファクタリング（任意）
```
- テスト追加
- ここで「痛み3」を経験
- DDD化の動機が生まれる
```

---

## 作成してほしいファイル一覧

### 1. services/bookmark/main.go

```go
package main

import (
    "log"
    "os"

    "bookmark/config"
    "bookmark/handler"
    "bookmark/pkg/database"
    "bookmark/pkg/http"
    "bookmark/pkg/sqs"
    "bookmark/repository"
    "bookmark/service"
)

func main() {
    // 設定読み込み
    cfg := config.Load()

    // DB接続
    db := database.Connect(cfg.DatabaseURL)

    // SQSクライアント
    sqsClient := sqs.NewClient(cfg.SQSQueueURL)

    // HTTPクライアント（OGP取得用）
    httpClient := http.NewClient()

    // Repository初期化
    bookmarkRepo := repository.NewBookmarkRepository(db)
    queueRepo := repository.NewQueueRepository(sqsClient, cfg.SQSQueueURL)

    // Service初期化
    ogpService := service.NewOgpService(httpClient)
    bookmarkService := service.NewBookmarkService(bookmarkRepo, queueRepo, ogpService)

    // MODE切り替え
    mode := os.Getenv("MODE")
    switch mode {
    case "sqs":
        worker := handler.NewWorker(sqsClient, cfg.SQSQueueURL, bookmarkService)
        worker.Run()
    case "batch":
        batch := handler.NewBatch(bookmarkService)
        batch.Run()
    default:
        router := handler.NewRouter(bookmarkService)
        log.Fatal(router.Run(":8080"))
    }
}
```

### 2. services/bookmark/config/config.go

```go
package config

import "os"

type Config struct {
    DatabaseURL string
    SQSQueueURL string
    AppEnv      string
}

func Load() *Config {
    return &Config{
        DatabaseURL: os.Getenv("DATABASE_URL"),
        SQSQueueURL: os.Getenv("SQS_QUEUE_URL"),
        AppEnv:      os.Getenv("APP_ENV"),
    }
}
```

### 3. services/bookmark/model/bookmark.go

```go
package model

import (
    "time"
    "github.com/lib/pq"
)

// Bookmark エンティティ
type Bookmark struct {
    ID          uint           `json:"id" gorm:"primaryKey"`
    URL         string         `json:"url" gorm:"not null;uniqueIndex"`
    Title       string         `json:"title"`
    Description string         `json:"description" gorm:"type:text"`
    ImageURL    string         `json:"image_url"`
    FaviconURL  string         `json:"favicon_url"`
    Status      string         `json:"status" gorm:"default:pending"` // pending/fetched/error/dead
    Tags        pq.StringArray `json:"tags" gorm:"type:text[]"`
    FetchedAt   *time.Time     `json:"fetched_at"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
}

// CreateBookmarkRequest リクエストDTO
type CreateBookmarkRequest struct {
    URL  string   `json:"url" binding:"required,url"`
    Tags []string `json:"tags"`
}

// BookmarkResponse レスポンスDTO
type BookmarkResponse struct {
    ID          uint      `json:"id"`
    URL         string    `json:"url"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    ImageURL    string    `json:"image_url"`
    FaviconURL  string    `json:"favicon_url"`
    Status      string    `json:"status"`
    Tags        []string  `json:"tags"`
    FetchedAt   *time.Time `json:"fetched_at"`
    CreatedAt   time.Time `json:"created_at"`
}

// SearchRequest 検索リクエスト
type SearchRequest struct {
    Query string   `form:"q"`
    Tags  []string `form:"tags"`
}

// SQSメッセージ
type BookmarkCreatedMessage struct {
    BookmarkID uint   `json:"bookmark_id"`
    URL        string `json:"url"`
}

// OGP情報
type OgpInfo struct {
    Title       string
    Description string
    ImageURL    string
    FaviconURL  string
}
```

### 4. services/bookmark/model/errors.go

```go
package model

import "errors"

var (
    ErrNotFound     = errors.New("not found")
    ErrInvalidInput = errors.New("invalid input")
    ErrInternal     = errors.New("internal error")
)
```

### 5. services/bookmark/handler/router.go

```go
package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "bookmark/model"
    "bookmark/service"
)

func NewRouter(bookmarkService *service.BookmarkService) *gin.Engine {
    r := gin.Default()

    // ヘルスチェック
    r.GET("/up", HealthCheck)

    // APIルート
    api := r.Group("/api")
    {
        bookmarkHandler := NewBookmarkHandler(bookmarkService)
        api.POST("/bookmarks", bookmarkHandler.Create)
        api.GET("/bookmarks/:id", bookmarkHandler.Get)
        api.GET("/bookmarks", bookmarkHandler.List)
        api.DELETE("/bookmarks/:id", bookmarkHandler.Delete)
        api.GET("/bookmarks/search", bookmarkHandler.Search)
    }

    return r
}

// handleError 共通エラーハンドラー
func handleError(c *gin.Context, err error) {
    switch err {
    case model.ErrNotFound:
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
    case model.ErrInvalidInput:
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
    }
}
```

### 6. services/bookmark/handler/bookmark.go

```go
package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "bookmark/model"
    "bookmark/service"
)

type BookmarkHandler struct {
    bookmarkService *service.BookmarkService
}

func NewBookmarkHandler(bookmarkService *service.BookmarkService) *BookmarkHandler {
    return &BookmarkHandler{bookmarkService: bookmarkService}
}

// Create POST /api/bookmarks
func (h *BookmarkHandler) Create(c *gin.Context) {
    var req model.CreateBookmarkRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    bookmark, err := h.bookmarkService.Create(c.Request.Context(), req.URL, req.Tags)
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusAccepted, model.BookmarkResponse{
        ID:          bookmark.ID,
        URL:         bookmark.URL,
        Title:       bookmark.Title,
        Description: bookmark.Description,
        ImageURL:    bookmark.ImageURL,
        FaviconURL:  bookmark.FaviconURL,
        Status:      bookmark.Status,
        Tags:        bookmark.Tags,
        FetchedAt:   bookmark.FetchedAt,
        CreatedAt:   bookmark.CreatedAt,
    })
}

// Get GET /api/bookmarks/:id
func (h *BookmarkHandler) Get(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }

    bookmark, err := h.bookmarkService.GetByID(c.Request.Context(), uint(id))
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, model.BookmarkResponse{
        ID:          bookmark.ID,
        URL:         bookmark.URL,
        Title:       bookmark.Title,
        Description: bookmark.Description,
        ImageURL:    bookmark.ImageURL,
        FaviconURL:  bookmark.FaviconURL,
        Status:      bookmark.Status,
        Tags:        bookmark.Tags,
        FetchedAt:   bookmark.FetchedAt,
        CreatedAt:   bookmark.CreatedAt,
    })
}

// List GET /api/bookmarks
func (h *BookmarkHandler) List(c *gin.Context) {
    bookmarks, err := h.bookmarkService.GetAll(c.Request.Context())
    if err != nil {
        handleError(c, err)
        return
    }

    var response []model.BookmarkResponse
    for _, b := range bookmarks {
        response = append(response, model.BookmarkResponse{
            ID:          b.ID,
            URL:         b.URL,
            Title:       b.Title,
            Description: b.Description,
            ImageURL:    b.ImageURL,
            FaviconURL:  b.FaviconURL,
            Status:      b.Status,
            Tags:        b.Tags,
            FetchedAt:   b.FetchedAt,
            CreatedAt:   b.CreatedAt,
        })
    }

    c.JSON(http.StatusOK, response)
}

// Delete DELETE /api/bookmarks/:id
func (h *BookmarkHandler) Delete(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }

    if err := h.bookmarkService.Delete(c.Request.Context(), uint(id)); err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusNoContent, nil)
}

// Search GET /api/bookmarks/search?q=keyword&tags=tag1,tag2
func (h *BookmarkHandler) Search(c *gin.Context) {
    var req model.SearchRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    bookmarks, err := h.bookmarkService.Search(c.Request.Context(), req.Query, req.Tags)
    if err != nil {
        handleError(c, err)
        return
    }

    var response []model.BookmarkResponse
    for _, b := range bookmarks {
        response = append(response, model.BookmarkResponse{
            ID:          b.ID,
            URL:         b.URL,
            Title:       b.Title,
            Description: b.Description,
            ImageURL:    b.ImageURL,
            FaviconURL:  b.FaviconURL,
            Status:      b.Status,
            Tags:        b.Tags,
            FetchedAt:   b.FetchedAt,
            CreatedAt:   b.CreatedAt,
        })
    }

    c.JSON(http.StatusOK, response)
}
```

### 7. services/bookmark/handler/health.go

```go
package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
```

### 8. services/bookmark/handler/worker.go

```go
package handler

import (
    "context"
    "encoding/json"
    "log"
    "sync"

    "github.com/aws/aws-sdk-go-v2/service/sqs"
    "bookmark/model"
    "bookmark/service"
)

type Worker struct {
    sqsClient        *sqs.Client
    queueURL         string
    bookmarkService  *service.BookmarkService
    concurrency      int
}

func NewWorker(sqsClient *sqs.Client, queueURL string, bookmarkService *service.BookmarkService) *Worker {
    return &Worker{
        sqsClient:       sqsClient,
        queueURL:        queueURL,
        bookmarkService: bookmarkService,
        concurrency:     10,
    }
}

func (w *Worker) Run() {
    log.Println("Starting SQS worker for OGP fetching...")

    for {
        output, err := w.sqsClient.ReceiveMessage(context.Background(), &sqs.ReceiveMessageInput{
            QueueUrl:            &w.queueURL,
            MaxNumberOfMessages: 10,
            WaitTimeSeconds:     20,
        })
        if err != nil {
            log.Printf("Error receiving messages: %v", err)
            continue
        }

        if len(output.Messages) == 0 {
            continue
        }

        var wg sync.WaitGroup
        sem := make(chan struct{}, w.concurrency)

        for _, msg := range output.Messages {
            wg.Add(1)
            sem <- struct{}{}

            go func(m sqs.Message) {
                defer wg.Done()
                defer func() { <-sem }()

                if err := w.processMessage(m); err != nil {
                    log.Printf("Error processing message: %v", err)
                    return // メッセージを削除しない → SQSリトライ
                }

                // 成功時のみ削除
                w.sqsClient.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
                    QueueUrl:      &w.queueURL,
                    ReceiptHandle: m.ReceiptHandle,
                })
            }(msg)
        }

        wg.Wait()
    }
}

func (w *Worker) processMessage(msg sqs.Message) error {
    var event model.BookmarkCreatedMessage
    if err := json.Unmarshal([]byte(*msg.Body), &event); err != nil {
        return err
    }

    // OGP情報を取得してDB更新
    return w.bookmarkService.FetchOgp(context.Background(), event.BookmarkID, event.URL)
}
```

### 9. services/bookmark/handler/batch.go

```go
package handler

import (
    "context"
    "log"
    "os"

    "bookmark/service"
)

type Batch struct {
    bookmarkService *service.BookmarkService
}

func NewBatch(bookmarkService *service.BookmarkService) *Batch {
    return &Batch{bookmarkService: bookmarkService}
}

func (b *Batch) Run() {
    batchType := os.Getenv("TYPE")

    log.Printf("Starting batch: %s", batchType)

    var err error
    switch batchType {
    case "refresh-ogp":
        // 7日以上古いOGP情報を再取得
        err = b.bookmarkService.RefreshOldOgp(context.Background(), 7)
    case "check-dead-links":
        // リンク切れチェック
        err = b.bookmarkService.CheckDeadLinks(context.Background())
    default:
        log.Fatalf("Unknown batch type: %s", batchType)
    }

    if err != nil {
        log.Fatalf("Batch failed: %v", err)
    }

    log.Println("Batch completed successfully")
}
```

### 10. services/bookmark/service/bookmark.go

```go
package service

import (
    "context"
    "log"
    "time"

    "bookmark/model"
    "bookmark/repository"
)

type BookmarkService struct {
    bookmarkRepo *repository.BookmarkRepository
    queueRepo    *repository.QueueRepository
    ogpService   *OgpService
}

func NewBookmarkService(
    bookmarkRepo *repository.BookmarkRepository,
    queueRepo *repository.QueueRepository,
    ogpService *OgpService,
) *BookmarkService {
    return &BookmarkService{
        bookmarkRepo: bookmarkRepo,
        queueRepo:    queueRepo,
        ogpService:   ogpService,
    }
}

// Create ブックマーク作成 → SQS送信
func (s *BookmarkService) Create(ctx context.Context, url string, tags []string) (*model.Bookmark, error) {
    bookmark := &model.Bookmark{
        URL:    url,
        Status: "pending",
        Tags:   tags,
    }

    if err := s.bookmarkRepo.Save(ctx, bookmark); err != nil {
        return nil, err
    }

    // SQSに送信（OGP取得を非同期で実行）
    if err := s.queueRepo.SendBookmarkCreated(ctx, bookmark.ID, bookmark.URL); err != nil {
        log.Printf("Failed to send SQS message: %v", err)
        // SQS送信失敗してもブックマークは作成済み
    }

    return bookmark, nil
}

// GetByID ブックマーク取得
func (s *BookmarkService) GetByID(ctx context.Context, id uint) (*model.Bookmark, error) {
    return s.bookmarkRepo.FindByID(ctx, id)
}

// GetAll 全ブックマーク取得
func (s *BookmarkService) GetAll(ctx context.Context) ([]model.Bookmark, error) {
    return s.bookmarkRepo.FindAll(ctx)
}

// Delete ブックマーク削除
func (s *BookmarkService) Delete(ctx context.Context, id uint) error {
    return s.bookmarkRepo.Delete(ctx, id)
}

// Search 検索（タイトル・タグ）
func (s *BookmarkService) Search(ctx context.Context, query string, tags []string) ([]model.Bookmark, error) {
    return s.bookmarkRepo.Search(ctx, query, tags)
}

// FetchOgp OGP情報取得（ワーカー用）
func (s *BookmarkService) FetchOgp(ctx context.Context, bookmarkID uint, url string) error {
    bookmark, err := s.bookmarkRepo.FindByID(ctx, bookmarkID)
    if err != nil {
        return err
    }

    // OGP情報を取得（痛みポイント1: このロジックはどこに書く？）
    ogpInfo, err := s.ogpService.Fetch(ctx, url)
    if err != nil {
        bookmark.Status = "error"
        s.bookmarkRepo.Save(ctx, bookmark)
        return err
    }

    // DB更新
    bookmark.Title = ogpInfo.Title
    bookmark.Description = ogpInfo.Description
    bookmark.ImageURL = ogpInfo.ImageURL
    bookmark.FaviconURL = ogpInfo.FaviconURL
    bookmark.Status = "fetched"
    now := time.Now()
    bookmark.FetchedAt = &now

    return s.bookmarkRepo.Save(ctx, bookmark)
}

// RefreshOldOgp 古いOGP情報を再取得（バッチ用）
func (s *BookmarkService) RefreshOldOgp(ctx context.Context, days int) error {
    cutoffDate := time.Now().AddDate(0, 0, -days)
    bookmarks, err := s.bookmarkRepo.FindOldFetched(ctx, cutoffDate)
    if err != nil {
        return err
    }

    for _, bookmark := range bookmarks {
        if err := s.FetchOgp(ctx, bookmark.ID, bookmark.URL); err != nil {
            log.Printf("Failed to refresh OGP for bookmark %d: %v", bookmark.ID, err)
            continue
        }
    }

    return nil
}

// CheckDeadLinks リンク切れチェック（バッチ用）
func (s *BookmarkService) CheckDeadLinks(ctx context.Context) error {
    bookmarks, err := s.bookmarkRepo.FindAll(ctx)
    if err != nil {
        return err
    }

    for _, bookmark := range bookmarks {
        if err := s.ogpService.CheckURL(ctx, bookmark.URL); err != nil {
            bookmark.Status = "dead"
            s.bookmarkRepo.Save(ctx, bookmark)
            log.Printf("Dead link detected: %s", bookmark.URL)
        }
    }

    return nil
}
```

### 11. services/bookmark/service/ogp.go

```go
package service

import (
    "context"
    "fmt"
    "net/http"
    "net/url"
    "strings"

    "github.com/PuerkitoBio/goquery"
    "github.com/go-resty/resty/v2"
    "bookmark/model"
)

type OgpService struct {
    httpClient *resty.Client
}

func NewOgpService(httpClient *resty.Client) *OgpService {
    return &OgpService{httpClient: httpClient}
}

// Fetch OGP情報を取得
func (s *OgpService) Fetch(ctx context.Context, targetURL string) (*model.OgpInfo, error) {
    resp, err := s.httpClient.R().
        SetContext(ctx).
        SetHeader("User-Agent", "Mozilla/5.0").
        Get(targetURL)
    
    if err != nil {
        return nil, err
    }

    if resp.StatusCode() != http.StatusOK {
        return nil, fmt.Errorf("HTTP %d", resp.StatusCode())
    }

    doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
    if err != nil {
        return nil, err
    }

    ogp := &model.OgpInfo{}

    // OGPタグから取得
    doc.Find("meta[property^='og:']").Each(func(i int, s *goquery.Selection) {
        prop, _ := s.Attr("property")
        content, _ := s.Attr("content")
        
        switch prop {
        case "og:title":
            ogp.Title = content
        case "og:description":
            ogp.Description = content
        case "og:image":
            ogp.ImageURL = s.resolveURL(targetURL, content)
        }
    })

    // フォールバック: titleタグ
    if ogp.Title == "" {
        ogp.Title = doc.Find("title").Text()
    }

    // favicon取得
    faviconURL := doc.Find("link[rel='icon']").AttrOr("href", "")
    if faviconURL == "" {
        faviconURL = doc.Find("link[rel='shortcut icon']").AttrOr("href", "")
    }
    if faviconURL != "" {
        ogp.FaviconURL = s.resolveURL(targetURL, faviconURL)
    }

    return ogp, nil
}

// CheckURL URLが有効かチェック
func (s *OgpService) CheckURL(ctx context.Context, targetURL string) error {
    resp, err := s.httpClient.R().
        SetContext(ctx).
        Head(targetURL)
    
    if err != nil {
        return err
    }

    if resp.StatusCode() >= 400 {
        return fmt.Errorf("HTTP %d", resp.StatusCode())
    }

    return nil
}

// resolveURL 相対URLを絶対URLに変換
func (s *OgpService) resolveURL(baseURL, relativeURL string) string {
    u, err := url.Parse(relativeURL)
    if err != nil {
        return relativeURL
    }

    base, err := url.Parse(baseURL)
    if err != nil {
        return relativeURL
    }

    return base.ResolveReference(u).String()
}
```

### 12. services/bookmark/repository/bookmark.go

```go
package repository

import (
    "context"
    "errors"
    "time"

    "gorm.io/gorm"
    "bookmark/model"
)

type BookmarkRepository struct {
    db *gorm.DB
}

func NewBookmarkRepository(db *gorm.DB) *BookmarkRepository {
    return &BookmarkRepository{db: db}
}

func (r *BookmarkRepository) Save(ctx context.Context, bookmark *model.Bookmark) error {
    return r.db.WithContext(ctx).Save(bookmark).Error
}

func (r *BookmarkRepository) FindByID(ctx context.Context, id uint) (*model.Bookmark, error) {
    var bookmark model.Bookmark
    err := r.db.WithContext(ctx).First(&bookmark, id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, model.ErrNotFound
    }
    return &bookmark, err
}

func (r *BookmarkRepository) FindAll(ctx context.Context) ([]model.Bookmark, error) {
    var bookmarks []model.Bookmark
    err := r.db.WithContext(ctx).Order("created_at DESC").Find(&bookmarks).Error
    return bookmarks, err
}

func (r *BookmarkRepository) Delete(ctx context.Context, id uint) error {
    return r.db.WithContext(ctx).Delete(&model.Bookmark{}, id).Error
}

// Search タイトル・タグで検索
func (r *BookmarkRepository) Search(ctx context.Context, query string, tags []string) ([]model.Bookmark, error) {
    var bookmarks []model.Bookmark
    db := r.db.WithContext(ctx)

    if query != "" {
        db = db.Where("title ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%")
    }

    if len(tags) > 0 {
        db = db.Where("tags && ?", tags) // PostgreSQL配列の重複チェック
    }

    err := db.Order("created_at DESC").Find(&bookmarks).Error
    return bookmarks, err
}

// FindOldFetched 指定日時より古いfetched_atのブックマークを取得
func (r *BookmarkRepository) FindOldFetched(ctx context.Context, cutoffDate time.Time) ([]model.Bookmark, error) {
    var bookmarks []model.Bookmark
    err := r.db.WithContext(ctx).
        Where("status = ? AND fetched_at < ?", "fetched", cutoffDate).
        Find(&bookmarks).Error
    return bookmarks, err
}
```

### 13. services/bookmark/repository/queue.go

```go
package repository

import (
    "context"
    "encoding/json"

    "github.com/aws/aws-sdk-go-v2/service/sqs"
    "bookmark/model"
)

type QueueRepository struct {
    sqsClient *sqs.Client
    queueURL  string
}

func NewQueueRepository(sqsClient *sqs.Client, queueURL string) *QueueRepository {
    return &QueueRepository{
        sqsClient: sqsClient,
        queueURL:  queueURL,
    }
}

func (r *QueueRepository) SendBookmarkCreated(ctx context.Context, bookmarkID uint, url string) error {
    msg := model.BookmarkCreatedMessage{
        BookmarkID: bookmarkID,
        URL:        url,
    }

    body, err := json.Marshal(msg)
    if err != nil {
        return err
    }

    bodyStr := string(body)
    _, err = r.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
        QueueUrl:    &r.queueURL,
        MessageBody: &bodyStr,
    })

    return err
}
```

### 14. services/bookmark/pkg/database/postgres.go

```go
package database

import (
    "log"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func Connect(dsn string) *gorm.DB {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    return db
}
```

### 15. services/bookmark/pkg/sqs/client.go

```go
package sqs

import (
    "context"
    "log"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

func NewClient(queueURL string) *sqs.Client {
    cfg, err := config.LoadDefaultConfig(context.Background())
    if err != nil {
        log.Fatalf("Failed to load AWS config: %v", err)
    }

    return sqs.NewFromConfig(cfg)
}
```

### 16. services/bookmark/pkg/http/client.go

```go
package http

import (
    "github.com/go-resty/resty/v2"
)

func NewClient() *resty.Client {
    return resty.New().
        SetTimeout(30).
        SetRetryCount(3)
}
```

### 17. services/bookmark/Dockerfile

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 依存関係インストール
COPY go.mod go.sum ./
RUN go mod download

# ソースコードコピー・ビルド
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server .

# Run stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/server .

USER appuser

EXPOSE 8080

CMD ["./server"]
```

### 18. ops/db-migrator/main.go

**用途**: `go run main.go` でマイグレーションを実行できるGo製ツール

```go
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
```

**実行例:**
```bash
# デフォルト設定で全てのマイグレーションをUP
go run main.go

# データベースURLを指定してUP
go run main.go -database "postgresql://postgres:postgres@localhost:5432/bookmark_dev?sslmode=disable"

# 全てのマイグレーションをロールバック
go run main.go -cmd down

# 1ステップだけロールバック
go run main.go -cmd down -steps 1

# 現在のバージョン確認
go run main.go -cmd version
```

### 19. ops/db-migrator/db/mydb/migrations/000001_create_bookmarks.up.sql

```sql
CREATE TABLE bookmarks (
    id SERIAL PRIMARY KEY,
    url VARCHAR(2048) NOT NULL UNIQUE,
    title VARCHAR(500),
    description TEXT,
    image_url VARCHAR(2048),
    favicon_url VARCHAR(2048),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    tags TEXT[],
    fetched_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- インデックス
CREATE INDEX idx_bookmarks_status ON bookmarks(status);
CREATE INDEX idx_bookmarks_created_at ON bookmarks(created_at DESC);
CREATE INDEX idx_bookmarks_fetched_at ON bookmarks(fetched_at);

-- GINインデックス（配列検索用）
CREATE INDEX idx_bookmarks_tags ON bookmarks USING GIN(tags);

-- 全文検索用インデックス（PostgreSQL）
CREATE INDEX idx_bookmarks_title_search ON bookmarks USING GIN(to_tsvector('english', title));
CREATE INDEX idx_bookmarks_description_search ON bookmarks USING GIN(to_tsvector('english', description));
```

### 20. ops/db-migrator/db/mydb/migrations/000001_create_bookmarks.down.sql

```sql
DROP TABLE IF EXISTS bookmarks;
```

---

## 処理フロー

### APIリクエスト → SQSワーカー（OGP取得）

```
1. POST /api/bookmarks → BookmarkHandler.Create
2. BookmarkService.Create
3. BookmarkRepository.Save（DB保存、status=pending）
4. QueueRepository.SendBookmarkCreated（SQS送信）
5. HTTP 202 Accepted 返却（非同期処理開始）

6. Worker.Run（SQSポーリング）
7. メッセージ受信 → goroutineで並列処理
8. BookmarkService.FetchOgp
   - OgpService.Fetch（HTTP + HTMLパース）
   - BookmarkRepository.Save（OGP情報更新、status=fetched）
9. 成功時: DeleteMessage / 失敗時: リトライ（DLQ）
```

### バッチ処理

```
1. EventBridge Scheduler / ecspresso run
2. MODE=batch TYPE=refresh-ogp で起動
3. Batch.Run → BookmarkService.RefreshOldOgp
4. 7日以上古いOGP情報を再取得

または

1. MODE=batch TYPE=check-dead-links で起動
2. Batch.Run → BookmarkService.CheckDeadLinks
3. 全ブックマークのリンク切れチェック
```

---

## 依存関係図

```
┌─────────────────────────────────────────────────────────────┐
│                     main.go                                 │
│                        │                                    │
│          ┌─────────────┼─────────────┐                      │
│          ▼             ▼             ▼                      │
│    ┌─────────┐   ┌─────────┐   ┌─────────┐                  │
│    │ handler │   │ handler │   │ handler │  Presentation   │
│    │ (HTTP)  │   │ (Worker)│   │ (Batch) │                  │
│    └────┬────┘   └────┬────┘   └────┬────┘                  │
│         │             │             │                       │
│         └─────────────┼─────────────┘                       │
│                       ▼                                     │
│          ┌──────────────────────────┐                      │
│          │      service             │   Business Logic     │
│          │  ┌──────────┐            │                      │
│          │  │ bookmark │            │                      │
│          │  └────┬─────┘            │                      │
│          │       │                  │                      │
│          │  ┌────▼─────┐            │                      │
│          │  │   ogp    │            │                      │
│          │  └────┬─────┘            │                      │
│          └───────┼──────────────────┘                      │
│                  │                                         │
│          ┌───────┼───────────────┐                         │
│          ▼                       ▼                         │
│    ┌───────────┐         ┌───────────┐                     │
│    │ repository│         │ repository│   Data Access      │
│    │(Bookmark) │         │  (Queue)  │                     │
│    └─────┬─────┘         └─────┬─────┘                     │
│          │                     │                          │
│          ▼                     ▼                          │
│    ┌──────────┐          ┌──────────┐                       │
│    │   pkg    │          │   pkg    │                       │
│    │(database)│          │  (http)  │                       │
│    └─────┬────┘          └─────┬────┘                       │
│          │                     │                          │
│          ▼                     ▼                          │
│    [PostgreSQL]            [External HTTP]                │
│    [SQS]                                                    │
└─────────────────────────────────────────────────────────────┘
```

---

## 技術的な学びポイント

| 機能 | 学べること |
|------|-----------|
| OGP取得 | HTTPクライアント、HTMLパース（goquery） |
| タグ検索 | PostgreSQL配列、GINインデックス |
| 画像URL | 外部URLのバリデーション、相対URL解決 |
| リトライ | SQS DLQ、エラーハンドリング |
| バッチ | 大量データの分割処理、日時比較 |

## 注意事項

1. **直接依存**: ServiceはRepositoryを直接参照（インターフェース不要）
2. **シンプルなDI**: main.goで全ての依存関係を組み立て
3. **goroutine**: WaitGroupで完了待ち、クロージャ問題に注意
4. **SQSリトライ**: 失敗時はDeleteMessageしない → SQSが自動リトライ → DLQ
5. **OGP取得のタイムアウト**: 外部サイトへのリクエストはタイムアウト設定必須
6. **User-Agent**: 一部サイトはUser-Agentがないと403を返す

---

## 追加プロンプト例

### 基本構成作成

```
上記のディレクトリ構成とファイル一覧に基づいて、
ブックマーク管理サービスのGoバックエンドの三層アーキテクチャを作成してください。

要件:
- POST /api/bookmarks でURL登録 → SQSに送信
- ワーカーがSQSからメッセージを受信 → OGP情報取得 → DB更新
- バッチで古いOGP情報を再取得・リンク切れチェック
- タグ検索機能（PostgreSQL配列）
- db-migratorでテーブル作成
```

### 機能追加

```
上記の構成に以下を追加してください:
- Swagger（swaggo/gin-swagger）
- Prometheusメトリクス
- グレースフルシャットダウン
- CORS設定
```

### テスト追加

```
三層アーキテクチャでのテスト構成を追加してください:
- handler: httptest + テスト用Service
- service: テスト用Repository（モック不要、構造体直接作成）
- repository: testcontainersでPostgreSQL
```

