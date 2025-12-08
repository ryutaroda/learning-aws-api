# ブックマーク管理サービス 段階的実装ガイド

このドキュメントは、Go + Gin + SQS + ECS のブックマーク管理サービスを段階的に実装するためのガイドです。

## 構成方針

**モノリポ（Monorepo）+ マイクロサービス**構成を採用します。

```
learning-aws-api/              # 1つのリポジトリ
├── services/
│   ├── bookmark/              # bookmarkサービス（今回実装）
│   ├── user/                  # userサービス（将来追加）
│   └── notification/          # notificationサービス（将来追加）
├── shared/                    # 共有コード（将来）
├── ops/                       # デプロイ設定
└── docs/                      # ドキュメント
```

各サービスは独立したコンテナとしてECSにデプロイされます。

詳細は [docs/architecture/go-three-layer-architecture.md](docs/architecture/go-three-layer-architecture.md) を参照してください。

---

## 使い方

1. 各ステップを上から順に実装
2. 実装完了後、AIに「ステップX完了」と報告
3. AIが記録欄を更新
4. 次のステップへ進む

**AIへの依頼例**: 「ステップ1.1完了しました。次のステップの詳細を教えてください」

---

## 進捗サマリー

| Phase | 内容 | 状態 |
|-------|------|------|
| Phase 1 | 環境構築・基本設定 | ✅ 完了 |
| Phase 2 | API基本実装（CRUD） | 🔄 進行中 |
| Phase 3 | SQSワーカー実装 | ⬜ 未着手 |
| Phase 4 | バッチ処理実装 | ⬜ 未着手 |
| Phase 5 | ECSデプロイ | ⬜ 未着手 |
| Phase 6 | Slack連携（将来） | ⬜ 未着手 |

---

# Phase 1: 環境構築・基本設定

## ステップ 1.1: プロジェクト初期化

### やること
- [ ] `services/bookmark/` ディレクトリ作成
- [ ] `go mod init bookmark` 実行
- [ ] 基本的なディレクトリ構造作成

### 作成するディレクトリ
```
services/bookmark/
├── config/
├── model/
├── handler/
├── service/
├── repository/
└── pkg/
    ├── database/
    ├── sqs/
    └── http/
```

### AIへの依頼プロンプト
```
ステップ1.1を実装したいです。
Go 1.24でプロジェクトを初期化し、
基本的なディレクトリ構造を作成するコマンドを教えてください。
```

### 記録欄
<!-- AI記入欄: ステップ完了時に更新 -->
| 項目 | 内容 |
|------|------|
| 状態 | ✅ 完了 |
| 完了日 | 2025-12-04 |
| 備考 | プロジェクト初期化完了、ディレクトリ構造作成済み |

---

## ステップ 1.2: 依存関係インストール

### やること
- [ ] `go mod init` でモジュール初期化（既に完了している場合はスキップ）
- [ ] モデル定義で使用する `github.com/lib/pq` を追加

### 依存関係の追加方法

**重要**: Goの依存関係管理は「必要になったら追加する」方式を推奨します。

```bash
# ステップ1.4でmodel/bookmark.goを作成する際に必要
go get github.com/lib/pq

# 依存関係を整理（go.sum作成・不要な依存関係を削除）
go mod tidy
```

### 今後の依存関係追加タイミング

| Phase | ステップ | 追加する依存関係 |
|-------|---------|-----------------|
| Phase 1 | 1.4 | `github.com/lib/pq` |
| Phase 2 | 2.1 | `gorm.io/gorm`, `gorm.io/driver/postgres` |
| Phase 2 | 2.3 | `github.com/gin-gonic/gin` |
| Phase 3 | 3.1 | `github.com/aws/aws-sdk-go-v2/...` |
| Phase 3 | 3.3 | `github.com/go-resty/resty/v2`, `github.com/PuerkitoBio/goquery` |

### AIへの依頼プロンプト
```
ステップ1.2を実装したいです。
go.modの初期化と、model/bookmark.goで使用するgithub.com/lib/pqの追加方法を教えてください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ✅ 完了 |
| 完了日 | 2025-12-04 |
| 備考 | go.mod初期化完了。依存関係は必要になったら追加する方式で進める |

---

## ステップ 1.3: 設定ファイル作成

### やること
- [ ] `config/config.go` 作成
- [ ] 環境変数から設定読み込み

### AIへの依頼プロンプト
```
ステップ1.3を実装したいです。
docs/architecture/go-three-layer-architecture.md の config/config.go を参考に、
設定ファイルを作成してください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ✅ 完了 |
| 完了日 | 2025-12-04 |
| 備考 | config/config.go作成完了 |

---

## ステップ 1.4: モデル定義

### やること
- [ ] `model/bookmark.go` 作成
- [ ] `model/errors.go` 作成

### AIへの依頼プロンプト
```
ステップ1.4を実装したいです。
docs/architecture/go-three-layer-architecture.md の model/bookmark.go と model/errors.go を参考に、
モデルファイルを作成してください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ✅ 完了 |
| 完了日 | 2025-12-04 |
| 備考 | model/bookmark.go, model/errors.go作成完了。Bookmarkエンティティ、DTO、SQSメッセージ、OGP情報の定義完了 |

---

## ステップ 1.5: DB接続設定

### やること
- [ ] `pkg/database/postgres.go` 作成
- [ ] ローカルPostgreSQL起動確認

### AIへの依頼プロンプト
```
ステップ1.5を実装したいです。
docs/architecture/go-three-layer-architecture.md の pkg/database/postgres.go を参考に、
DB接続ファイルを作成してください。
また、docker-composeでローカルPostgreSQLを起動する方法も教えてください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ✅ 完了 |
| 完了日 | 2025-12-07 |
| 備考 | pkg/database/postgres.go作成完了、docker-compose.ymlでローカルPostgreSQL起動確認完了 |

---

# Phase 2: API基本実装（CRUD）

## ステップ 2.1: Repository実装

### やること
- [ ] GORM依存関係を追加
- [ ] `repository/bookmark.go` 作成
- [ ] CRUD操作の実装

### 依存関係追加
```bash
go get gorm.io/gorm
go get gorm.io/driver/postgres
go mod tidy
```

### AIへの依頼プロンプト
```
ステップ2.1を実装したいです。
docs/architecture/go-three-layer-architecture.md の repository/bookmark.go を参考に、
リポジトリファイルを作成してください。
GORMを使用するので、必要な依存関係も追加してください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ✅ 完了 |
| 完了日 | 2025-12-07 |
| 備考 | repository/bookmark.go作成完了。Save, FindByID, FindAll, Delete, Search, FindOldFetchedの6メソッド実装。GORM依存関係追加完了 |

---

## ステップ 2.2: Service実装（OGP抜き）

### やること
- [ ] `service/bookmark.go` 作成（OGP取得以外）
- [ ] Create, GetByID, GetAll, Delete, Search 実装

### AIへの依頼プロンプト
```
ステップ2.2を実装したいです。
docs/architecture/go-three-layer-architecture.md の service/bookmark.go を参考に、
OGP取得以外のメソッドを実装してください。
FetchOgp, RefreshOldOgp, CheckDeadLinks は後で実装するのでスキップしてください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ✅ 完了 |
| 完了日 | 2025-12-07 |
| 備考 | service/bookmark.go作成完了。Create, GetByID, GetAll, Delete, Searchの5メソッド実装。OGP関連メソッドは後で実装 |

---

## ステップ 2.3: Handler実装

### やること
- [ ] Gin依存関係を追加
- [ ] `handler/router.go` 作成
- [ ] `handler/bookmark.go` 作成
- [ ] `handler/health.go` 作成

### 依存関係追加
```bash
go get github.com/gin-gonic/gin
go mod tidy
```

### AIへの依頼プロンプト
```
ステップ2.3を実装したいです。
docs/architecture/go-three-layer-architecture.md の handler/ 配下のファイルを参考に、
HTTPハンドラーを作成してください。
Ginフレームワークを使用するので、必要な依存関係も追加してください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ✅ 完了 |
| 完了日 | 2025-12-07 |
| 備考 | handler/health.go, router.go, bookmark.go作成完了。6つのエンドポイント実装。Gin依存関係追加完了 |

---

## ステップ 2.4: main.go実装（API起動のみ）

### やること
- [ ] `main.go` 作成
- [ ] APIサーバー起動確認

### AIへの依頼プロンプト
```
ステップ2.4を実装したいです。
docs/architecture/go-three-layer-architecture.md の main.go を参考に、
APIサーバーのみ起動するmain.goを作成してください。
SQSクライアントとワーカー関連は後で追加するのでスキップしてください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ✅ 完了 |
| 完了日 | 2025-12-07 |
| 備考 | main.go作成完了。設定読込→DB接続→Repository→Service→Router→サーバー起動の流れを実装 |

---

## ステップ 2.5: マイグレーション作成・実行（Docker版）

### やること
- [ ] `ops/db-migrator/Dockerfile` 作成
- [ ] `ops/db-migrator/db/mydb/migrations/` ディレクトリ作成
- [ ] マイグレーションファイル作成（up/down）
- [ ] docker-compose.ymlにdb-migratorサービス追加
- [ ] マイグレーション実行

### AIへの依頼プロンプト
```
ステップ2.5を実装したいです。
Dockerコンテナでマイグレーションを実行できるようにしたいです。
以下を作成してください：
1. ops/db-migrator/Dockerfile
2. ops/db-migrator/db/mydb/migrations/000001_create_bookmarks.up.sql
3. ops/db-migrator/db/mydb/migrations/000001_create_bookmarks.down.sql
4. docker-compose.ymlへのdb-migratorサービス追加
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 2.5B: Goコードでマイグレーション実行（推奨）

### やること
- [ ] `ops/db-migrator/go.mod` 初期化
- [ ] `ops/db-migrator/main.go` 作成
- [ ] golang-migrate依存関係追加
- [ ] `go run main.go` でマイグレーション実行

### 実装手順

#### 1. Go moduleを初期化
```bash
cd /Users/kanaokaryuutarou/Documents/GitHub/personal/learning-aws-workspace/learning-aws-api/ops/db-migrator

# Go moduleを初期化
go mod init db-migrator

# 必要な依存関係を追加
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/postgres
go get -u github.com/golang-migrate/migrate/v4/source/file
go get -u github.com/lib/pq
```

#### 2. main.goを作成

**作成するファイル:** `ops/db-migrator/main.go`

**参考:** `docs/architecture/go-three-layer-architecture.md` の「### 18. ops/db-migrator/main.go」を参照してコードを実装してください。

**実装内容のポイント:**
- コマンドライン引数（`flag`パッケージ）でパスやコマンドを指定可能
- 環境変数 `DATABASE_URL` からデータベース接続情報を取得
- デフォルト値として `postgresql://postgres:postgres@localhost:5432/bookmark_dev?sslmode=disable` を使用
- コマンド: `up` / `down` / `version`
- ステップ指定: `-steps` で段階的なマイグレーション実行可能

#### 3. 実行方法

**A. マイグレーションUP（全て適用）**
```bash
cd /Users/kanaokaryuutarou/Documents/GitHub/personal/learning-aws-workspace/learning-aws-api/ops/db-migrator

# デフォルト設定で実行（localhost:5432のPostgreSQLに接続）
go run main.go

# データベースURLを指定して実行
go run main.go -database "postgresql://postgres:postgres@localhost:5432/bookmark_dev?sslmode=disable"
```

**B. マイグレーションDOWN（ロールバック）**
```bash
# 全てのマイグレーションをロールバック
go run main.go -cmd down

# 1ステップだけロールバック
go run main.go -cmd down -steps 1
```

**C. 現在のバージョン確認**
```bash
go run main.go -cmd version
```

**D. ステップ指定でUP**
```bash
# 1ステップだけマイグレーション
go run main.go -cmd up -steps 1
```

#### 4. 実行前の準備

PostgreSQLコンテナが起動していることを確認：
```bash
cd /Users/kanaokaryuutarou/Documents/GitHub/personal/learning-aws-workspace/learning-aws-api/services/bookmark
docker-compose up -d postgres
```

#### 5. 動作確認

```bash
# 1. 既存テーブル削除（クリーンな状態から開始）
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "DROP TABLE IF EXISTS bookmarks CASCADE;"
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "DROP TABLE IF EXISTS schema_migrations CASCADE;"

# 2. マイグレーション実行
cd /Users/kanaokaryuutarou/Documents/GitHub/personal/learning-aws-workspace/learning-aws-api/ops/db-migrator
go run main.go

# 3. テーブル確認
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "\dt"
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "\d bookmarks"
```

**期待される出力:**
- `bookmarks` テーブルが存在
- `schema_migrations` テーブルが存在（マイグレーション履歴管理用）
- bookmarksテーブルに `id`, `url`, `title`, `description`, `tags` などのカラムが存在

### AIへの依頼プロンプト
```
ステップ2.5Bを実装したいです。
docs/architecture/go-three-layer-architecture.md の「### 18. ops/db-migrator/main.go」を参考に、
ops/db-migrator/main.go を作成してください。

要件:
- コマンドライン引数（flag）でパス、データベースURL、コマンドを指定可能
- 環境変数DATABASE_URLに対応
- コマンド: up/down/version
- ステップ指定可能（-steps）
```

### 補足: docker-composeとの比較

| アプローチ | メリット | デメリット |
|-----------|---------|-----------|
| **docker-compose** | CI/CD統合が簡単、環境の再現性が高い | ローカル実行が複雑 |
| **go run main.go** | ローカル開発が簡単、デバッグしやすい | 本番環境への適用には工夫が必要 |

**推奨:**
- **開発中**: `go run main.go` で快適に開発（ステップ2.5B）
- **本番/CI/CD**: docker-composeまたはビルド済みバイナリを使用（ステップ2.5）

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ✅ 完了 |
| 完了日 | 2025-12-08 |
| 備考 | ops/db-migrator/main.go作成完了。go run main.goでマイグレーション実行成功。bookmarksテーブルとschema_migrationsテーブルが作成された。README.mdも作成してコマンド集を整理 |

---

## ステップ 2.6: 動作確認（CRUD）

### やること
- [ ] `go run main.go` で起動
- [ ] curl/Postman で CRUD 動作確認

### 確認するエンドポイント
```bash
# ヘルスチェック
curl http://localhost:8080/up

# ブックマーク作成
curl -X POST http://localhost:8080/api/bookmarks \
  -H "Content-Type: application/json" \
  -d '{"url":"https://go.dev","tags":["go","programming"]}'

# 一覧取得
curl http://localhost:8080/api/bookmarks

# 詳細取得
curl http://localhost:8080/api/bookmarks/1

# 削除
curl -X DELETE http://localhost:8080/api/bookmarks/1
```

### AIへの依頼プロンプト
```
ステップ2.6の動作確認をしています。
[エラー内容やログを貼り付け]
原因と解決方法を教えてください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

# Phase 3: SQSワーカー実装

## ステップ 3.1: SQSクライアント作成

### やること
- [ ] AWS SDK依存関係を追加
- [ ] `pkg/sqs/client.go` 作成
- [ ] ローカルSQS（LocalStack）起動

### 依存関係追加
```bash
go get github.com/aws/aws-sdk-go-v2
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/sqs
go mod tidy
```

### AIへの依頼プロンプト
```
ステップ3.1を実装したいです。
docs/architecture/go-three-layer-architecture.md の pkg/sqs/client.go を参考に、
SQSクライアントを作成してください。
AWS SDKを使用するので、必要な依存関係も追加してください。
また、ローカル開発用にLocalStackでSQSを起動する方法も教えてください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 3.2: Queue Repository実装

### やること
- [ ] `repository/queue.go` 作成
- [ ] SendBookmarkCreated 実装

### AIへの依頼プロンプト
```
ステップ3.2を実装したいです。
docs/architecture/go-three-layer-architecture.md の repository/queue.go を参考に、
キューリポジトリを作成してください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 3.3: HTTPクライアント・OGPサービス実装

### やること
- [ ] HTTPクライアント・HTMLパース依存関係を追加
- [ ] `pkg/http/client.go` 作成
- [ ] `service/ogp.go` 作成

### 依存関係追加
```bash
go get github.com/go-resty/resty/v2
go get github.com/PuerkitoBio/goquery
go mod tidy
```

### AIへの依頼プロンプト
```
ステップ3.3を実装したいです。
docs/architecture/go-three-layer-architecture.md の pkg/http/client.go と service/ogp.go を参考に、
OGP取得機能を作成してください。
restyとgoqueryを使用するので、必要な依存関係も追加してください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 3.4: BookmarkService拡張

### やること
- [ ] `service/bookmark.go` にOGP関連メソッド追加
- [ ] FetchOgp 実装
- [ ] SQS送信処理追加

### AIへの依頼プロンプト
```
ステップ3.4を実装したいです。
ステップ2.2で作成した service/bookmark.go に、
FetchOgp メソッドと SQS送信処理を追加してください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 3.5: Worker実装

### やること
- [ ] `handler/worker.go` 作成
- [ ] SQSポーリング・並列処理実装

### AIへの依頼プロンプト
```
ステップ3.5を実装したいです。
docs/architecture/go-three-layer-architecture.md の handler/worker.go を参考に、
SQSワーカーを作成してください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 3.6: main.go拡張（MODE対応）

### やること
- [ ] main.go にMODE切り替え追加
- [ ] SQSクライアント初期化追加

### AIへの依頼プロンプト
```
ステップ3.6を実装したいです。
ステップ2.4で作成した main.go に、
MODE環境変数でAPIとワーカーを切り替える処理を追加してください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 3.7: 動作確認（Worker）

### やること
- [ ] API起動（MODE未設定）
- [ ] Worker起動（MODE=sqs）
- [ ] ブックマーク作成 → OGP取得確認

### 確認手順
```bash
# ターミナル1: API起動
go run main.go

# ターミナル2: Worker起動
MODE=sqs go run main.go

# ターミナル3: ブックマーク作成
curl -X POST http://localhost:8080/api/bookmarks \
  -H "Content-Type: application/json" \
  -d '{"url":"https://go.dev","tags":["go"]}'

# 少し待ってから確認（OGP取得完了後）
curl http://localhost:8080/api/bookmarks/1
# title, description, image_url が設定されているはず
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

# Phase 4: バッチ処理実装

## ステップ 4.1: バッチハンドラー実装

### やること
- [ ] `handler/batch.go` 作成
- [ ] refresh-ogp, check-dead-links 実装

### AIへの依頼プロンプト
```
ステップ4.1を実装したいです。
docs/architecture/go-three-layer-architecture.md の handler/batch.go を参考に、
バッチハンドラーを作成してください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 4.2: BookmarkService拡張（バッチ用）

### やること
- [ ] RefreshOldOgp 実装
- [ ] CheckDeadLinks 実装

### AIへの依頼プロンプト
```
ステップ4.2を実装したいです。
service/bookmark.go に RefreshOldOgp と CheckDeadLinks メソッドを追加してください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 4.3: main.go拡張（batch対応）

### やること
- [ ] main.go にMODE=batch 追加

### AIへの依頼プロンプト
```
ステップ4.3を実装したいです。
main.go に MODE=batch の処理を追加してください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 4.4: 動作確認（Batch）

### やること
- [ ] バッチ実行確認

### 確認手順
```bash
# OGP更新バッチ
MODE=batch TYPE=refresh-ogp go run main.go

# リンク切れチェックバッチ
MODE=batch TYPE=check-dead-links go run main.go
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

# Phase 5: ECSデプロイ

## ステップ 5.1: Dockerfile作成

### やること
- [ ] `Dockerfile` 作成
- [ ] ローカルでビルド確認

### AIへの依頼プロンプト
```
ステップ5.1を実装したいです。
docs/architecture/go-three-layer-architecture.md の Dockerfile を参考に、
Dockerfileを作成してください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 5.2: db-migrator Dockerfile作成

### やること
- [ ] `ops/db-migrator/Dockerfile` 作成

### AIへの依頼プロンプト
```
ステップ5.2を実装したいです。
db-migrator用のDockerfileを作成してください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 5.3: ECRプッシュ

### やること
- [ ] ECRリポジトリ作成
- [ ] イメージビルド・プッシュ

### AIへの依頼プロンプト
```
ステップ5.3を実装したいです。
ECRにイメージをプッシュする手順を教えてください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 5.4: bookmark-api用ECSタスク定義・サービス作成

### やること
- [ ] ECSクラスター確認
- [ ] `bookmark-api-task-stg` タスク定義作成
- [ ] `bookmark-api-stg` サービス作成
- [ ] ALB Target Group設定

### AIへの依頼プロンプト
```
ステップ5.4を実装したいです。
bookmark-api用のECSタスク定義とサービスを作成する手順を教えてください。
- サービス名: bookmark-api-stg
- タスク定義: bookmark-api-task-stg
- コンテナ名: api
- MODE環境変数: api（デフォルト）
- ポート: 8080
- ALB Target Groupに接続
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 5.5: bookmark-worker用ECSタスク定義・サービス作成

### やること
- [ ] `bookmark-worker-task-stg` タスク定義作成
- [ ] `bookmark-worker-stg` サービス作成
- [ ] SQS Queue URL設定

### AIへの依頼プロンプト
```
ステップ5.5を実装したいです。
bookmark-worker用のECSタスク定義とサービスを作成する手順を教えてください。
- サービス名: bookmark-worker-stg
- タスク定義: bookmark-worker-task-stg
- コンテナ名: worker
- MODE環境変数: sqs
- SQS Queue URLを環境変数で設定
- ポートマッピングは不要（HTTPサーバーではない）
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 5.6: bookmark-batch用ECSタスク定義・サービス作成（オプション）

### やること
- [ ] `bookmark-batch-task-stg` タスク定義作成
- [ ] `bookmark-batch-stg` サービス作成（タスク数0）
- [ ] EventBridge Scheduler設定（オプション）

### AIへの依頼プロンプト
```
ステップ5.6を実装したいです。
bookmark-batch用のECSタスク定義とサービスを作成する手順を教えてください。
- サービス名: bookmark-batch-stg
- タスク定義: bookmark-batch-task-stg
- コンテナ名: batch
- MODE環境変数: batch
- TYPE環境変数: refresh-ogp または check-dead-links
- タスク数: 0（スケジュール実行時のみ起動）
- EventBridge Schedulerで定期実行する設定も教えてください
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 5.7: 動作確認（ECS）

### やること
- [ ] ALB経由でAPI疎通確認
- [ ] ブックマーク作成 → OGP取得確認（Worker動作確認）
- [ ] バッチ実行確認（オプション）

### 確認手順
```bash
# API疎通確認
curl https://your-alb-url/api/bookmarks

# ブックマーク作成
curl -X POST https://your-alb-url/api/bookmarks \
  -H "Content-Type: application/json" \
  -d '{"url":"https://go.dev","tags":["go"]}'

# 少し待ってから確認（OGP取得完了後）
curl https://your-alb-url/api/bookmarks/1
# title, description, image_url が設定されているはず
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

# Phase 6: Slack連携（将来）

## ステップ 6.1: マイグレーション追加

### やること
- [ ] Slack用カラム追加マイグレーション作成

### AIへの依頼プロンプト
```
ステップ6.1を実装したいです。
Slack連携用に source, slack_channel_id, slack_user_id カラムを追加する
マイグレーションファイルを作成してください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 6.2: Slack App作成

### やること
- [ ] Slack App作成（api.slack.com）
- [ ] Slash Command設定

### AIへの依頼プロンプト
```
ステップ6.2を実装したいです。
Slack Appの作成手順と /bookmark コマンドの設定方法を教えてください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 6.3: Slackハンドラー実装

### やること
- [ ] `handler/slack.go` 作成
- [ ] Slash Command 処理実装

### AIへの依頼プロンプト
```
ステップ6.3を実装したいです。
Slack Slash Command を受け取る handler/slack.go を作成してください。
```

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

## ステップ 6.4: 動作確認（Slack）

### やること
- [ ] Slackから /bookmark コマンド実行
- [ ] ブックマーク登録確認

### 記録欄
| 項目 | 内容 |
|------|------|
| 状態 | ⬜ 未着手 |
| 完了日 | - |
| 備考 | - |

---

# 完了チェックリスト

## Phase 1: 環境構築・基本設定
- [x] 1.1 プロジェクト初期化
- [x] 1.2 依存関係インストール
- [x] 1.3 設定ファイル作成
- [x] 1.4 モデル定義
- [x] 1.5 DB接続設定

## Phase 2: API基本実装（CRUD）
- [x] 2.1 Repository実装
- [x] 2.2 Service実装（OGP抜き）
- [x] 2.3 Handler実装
- [x] 2.4 main.go実装
- [ ] 2.5 マイグレーション作成・実行（Docker版）
- [x] 2.5B Goコードでマイグレーション実行（推奨）
- [ ] 2.6 動作確認（CRUD）

## Phase 3: SQSワーカー実装
- [ ] 3.1 SQSクライアント作成
- [ ] 3.2 Queue Repository実装
- [ ] 3.3 HTTPクライアント・OGPサービス実装
- [ ] 3.4 BookmarkService拡張
- [ ] 3.5 Worker実装
- [ ] 3.6 main.go拡張（MODE対応）
- [ ] 3.7 動作確認（Worker）

## Phase 4: バッチ処理実装
- [ ] 4.1 バッチハンドラー実装
- [ ] 4.2 BookmarkService拡張（バッチ用）
- [ ] 4.3 main.go拡張（batch対応）
- [ ] 4.4 動作確認（Batch）

## Phase 5: ECSデプロイ
- [ ] 5.1 Dockerfile作成
- [ ] 5.2 db-migrator Dockerfile作成
- [ ] 5.3 ECRプッシュ
- [ ] 5.4 bookmark-api用ECSタスク定義・サービス作成
- [ ] 5.5 bookmark-worker用ECSタスク定義・サービス作成
- [ ] 5.6 bookmark-batch用ECSタスク定義・サービス作成（オプション）
- [ ] 5.7 動作確認（ECS）

## Phase 6: Slack連携（将来）
- [ ] 6.1 マイグレーション追加
- [ ] 6.2 Slack App作成
- [ ] 6.3 Slackハンドラー実装
- [ ] 6.4 動作確認（Slack）

---

# 実装メモ

<!-- AI記入欄: 実装中に気づいた点や学びを記録 -->

## 痛みポイント記録

| Phase | 痛みポイント | 感じたこと | DDDでの解決策 |
|-------|-------------|-----------|--------------|
| - | - | - | - |

## エラー・解決記録

| 日付 | ステップ | エラー内容 | 解決方法 |
|------|---------|-----------|---------|
| - | - | - | - |

## 学んだこと

| 日付 | 内容 |
|------|------|
| - | - |

