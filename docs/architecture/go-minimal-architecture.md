# Go最小構成アーキテクチャ プロンプト

このドキュメントは、Go + Gin + SQS + ECS を使用した最小構成のバックエンドアプリケーションを新規作成するためのプロンプトです。

---

## 概要

以下の機能を持つGoバックエンドを構築します：

- **APIサーバー**: Gin フレームワーク
- **ワーカー**: SQSメッセージを並列処理
- **バッチ**: 定期実行タスク
- **DBマイグレーション**: golang-migrate

同一イメージを環境変数（MODE）で切り替えて運用します。

---

## ディレクトリ構成

```
myapp/
├── services/
│   └── myapp/                      # メインアプリケーション
│       ├── main.go                 # エントリーポイント（MODE切り替え）
│       ├── go.mod
│       ├── Dockerfile
│       ├── Makefile
│       ├── config/
│       │   ├── config.go           # 設定読み込み
│       │   └── config.toml.tmpl    # 設定テンプレート
│       ├── domain/
│       │   ├── entity.go           # エンティティ
│       │   ├── repository.go       # リポジトリインターフェース
│       │   ├── service.go          # ドメインサービス
│       │   └── errors.go           # ドメインエラー定義
│       ├── application/
│       │   └── usecase.go          # ユースケース
│       ├── infrastructure/
│       │   ├── postgres/
│       │   │   ├── postgres.go     # DB接続
│       │   │   └── repository.go   # リポジトリ実装
│       │   └── sqs/
│       │       ├── client.go       # SQSクライアント
│       │       └── producer.go     # キュー送信
│       └── interface/
│           ├── http/
│           │   ├── http.go         # Gin初期化・共通処理
│           │   └── handler.go      # HTTPハンドラー
│           ├── sqs/
│           │   └── consumer.go     # ワーカー（SQS受信）
│           └── cli/
│               └── command.go      # バッチ処理
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
│   │               ├── 000001_create_users.up.sql
│   │               └── 000001_create_users.down.sql
│   │
│   └── ecspresso/                  # ECSデプロイ設定
│       ├── myapp-api/
│       │   └── stg/
│       │       ├── ecspresso.yml
│       │       ├── ecs-task-def.json
│       │       └── ecs-service-def.json
│       ├── myapp-batch/
│       │   └── stg/
│       │       ├── ecspresso.yml
│       │       ├── ecs-task-def.json
│       │       └── ecs-task-def.overrides.json
│       └── db-migrator/
│           └── stg/
│               ├── ecspresso.yml
│               └── ecs-task-def.json
│
├── .github/
│   └── workflows/
│       └── deploy.yaml             # CI/CD
│
├── Makefile                        # ルートMakefile
└── README.md
```

---

## 技術スタック

| カテゴリ | 技術 |
|----------|------|
| 言語 | Go 1.24+ |
| Webフレームワーク | Gin |
| ORM | GORM |
| DB | PostgreSQL |
| キュー | Amazon SQS |
| マイグレーション | golang-migrate |
| コンテナ | Docker + ECS Fargate |
| デプロイ | ecspresso |
| CI/CD | GitHub Actions |

---

## 実行モード

| MODE | 用途 | 実行方法 |
|------|------|----------|
| (default) | APIサーバー | ECS Service |
| sqs | SQSワーカー | ECS Service（サイドカー） |
| batch | バッチ処理 | ecspresso run / EventBridge Scheduler |

---

## 作成してほしいファイル一覧

### 1. services/myapp/main.go

```go
package main

// エントリーポイント
// - 環境変数MODEで動作を切り替え
// - 共通の依存関係を初期化（DB, SQS, etc）
// - MODE=api: Ginサーバー起動
// - MODE=sqs: SQSワーカー起動
// - MODE=batch: バッチ処理実行
```

### 2. services/myapp/config/config.go

```go
package config

// 設定読み込み
// - 環境変数から設定を読み込む
// - TOMLテンプレートをサポート
```

### 3. services/myapp/domain/entity.go

```go
package domain

// エンティティ定義
// - Task構造体（ID, Name, Status, CreatedAt, UpdatedAt）
```

### 4. services/myapp/domain/repository.go

```go
package domain

// リポジトリインターフェース
// - TaskRepository: Save, FindByID, FindAll
// - TaskEventRepository: SendCreated
```

### 5. services/myapp/domain/service.go

```go
package domain

// ドメインサービス
// - TaskService: Create, Process, ProcessAll
```

### 6. services/myapp/domain/errors.go

```go
package domain

// ドメインエラー定義
// - ErrNotFound, ErrUnauthorized, ErrInvalidInput, etc
```

### 7. services/myapp/application/usecase.go

```go
package application

// ユースケース
// - TaskUsecase: Create, Process, ProcessAll
```

### 8. services/myapp/infrastructure/postgres/postgres.go

```go
package postgres

// DB接続
// - Connect関数
// - GORMの初期化
```

### 9. services/myapp/infrastructure/postgres/repository.go

```go
package postgres

// リポジトリ実装
// - TaskRepositoryImpl
```

### 10. services/myapp/infrastructure/sqs/client.go

```go
package sqs

// SQSクライアント
// - newClient関数
```

### 11. services/myapp/infrastructure/sqs/producer.go

```go
package sqs

// キュー送信
// - TaskEventRepositoryImpl
// - SendCreated: タスク作成イベントを送信
```

### 12. services/myapp/interface/http/http.go

```go
package http

// Gin初期化
// - NewGinEngine: ミドルウェア設定
// - handleError: 共通エラーハンドラー
```

### 13. services/myapp/interface/http/handler.go

```go
package http

// HTTPハンドラー
// - TaskHandler: CreateTask, GetTask
```

### 14. services/myapp/interface/sqs/consumer.go

```go
package sqs

// SQSワーカー
// - Consumer: Run（ロングポーリング + 並列処理）
// - sync.WaitGroupでgoroutine管理
// - エラー時はDeleteMessageしない（SQSリトライ）
```

### 15. services/myapp/interface/cli/command.go

```go
package cli

// バッチ処理
// - Command: Run
// - TYPE環境変数で処理を切り替え
```

### 16. services/myapp/Dockerfile

```dockerfile
# マルチステージビルド
# - builder: Goビルド
# - runner: Alpine + バイナリ
# - 非rootユーザー実行
```

### 17. ops/db-migrator/main.go

```go
package main

// DBマイグレーション
// - golang-migrateを使用
// - MODE=up: マイグレーション適用
// - MODE=down: ロールバック
```

### 18. ops/db-migrator/Dockerfile

```dockerfile
# マルチステージビルド
# - マイグレーションファイルをコピー
```

### 19. ops/db-migrator/db/mydb/migrations/000001_create_tasks.up.sql

```sql
-- tasksテーブル作成
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 20. ops/ecspresso/myapp-api/stg/ecspresso.yml

```yaml
# ecspresso設定
# - cluster, service, task_definition
```

### 21. ops/ecspresso/myapp-api/stg/ecs-task-def.json

```json
// ECSタスク定義
// - apiコンテナ: MODE未設定
// - workerコンテナ: MODE=sqs
// - 同一イメージを使用
```

### 22. .github/workflows/deploy.yaml

```yaml
# CI/CD
# 1. 変更があったDockerfileを検出
# 2. イメージビルド・ECRプッシュ
# 3. db-migrator実行（ecspresso run）
# 4. アプリデプロイ（ecspresso deploy）
```

---

## 処理フロー

### APIリクエスト → SQSワーカー

```
1. POST /tasks → TaskHandler.CreateTask
2. TaskUsecase.Create → TaskService.Create
3. TaskRepository.Save（DB保存）
4. TaskEventRepository.SendCreated（SQS送信）
5. HTTP 202 Accepted 返却

6. Consumer.Run（SQSポーリング）
7. メッセージ受信 → goroutineで並列処理
8. TaskUsecase.Process
9. 成功時: DeleteMessage / 失敗時: リトライ（DLQ）
```

### バッチ処理

```
1. EventBridge Scheduler / ecspresso run
2. MODE=batch TYPE=process-all で起動
3. Command.Run → TaskUsecase.ProcessAll
4. 全タスクを順次処理
```

---

## 注意事項

1. **クリーンアーキテクチャ**: 依存の方向は内側へ（interface → application → domain ← infrastructure）
2. **DI**: ハンドラーはUsecaseを受け取る（テスト容易性）
3. **エラーハンドリング**: ドメインエラーをHTTPステータスに変換
4. **goroutine**: WaitGroupで完了待ち、クロージャ問題に注意
5. **SQSリトライ**: 失敗時はDeleteMessageしない → SQSが自動リトライ → DLQ

---

## 追加プロンプト例

### 基本構成作成

```
上記のディレクトリ構成とファイル一覧に基づいて、
Taskエンティティを管理するGoバックエンドの最小構成を作成してください。

要件:
- POST /api/tasks でタスク作成 → SQSに送信
- ワーカーがSQSからメッセージを受信して処理
- バッチで全タスクを一括処理
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

### CI/CD追加

```
GitHub Actionsで以下のCI/CDを構築してください:
- developブランチ → stg環境
- mainブランチ → prd環境
- 変更があったDockerfileのみビルド
- ecspressoでデプロイ
```

