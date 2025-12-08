# Bookmark Service

ブックマーク管理サービス（Go + Gin + PostgreSQL）

## 🚀 クイックスタート

### 前提条件

- Go 1.25+
- Docker Desktop
- direnv（推奨）

### 1. direnvのセットアップ（推奨）

direnvをインストールして、ディレクトリに入ると自動的に環境変数を設定します。

```bash
# Macの場合
brew install direnv

# シェルの設定に追加
echo 'eval "$(direnv hook zsh)"' >> ~/.zshrc  # zshの場合
echo 'eval "$(direnv hook bash)"' >> ~/.bashrc  # bashの場合

# 設定を反映
source ~/.zshrc  # または source ~/.bashrc
```

### 2. 環境変数の設定

```bash
cd services/bookmark

# .envrcファイルを作成
cat > .envrc << 'EOF'
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/bookmark_dev?sslmode=disable"
export APP_ENV="development"
EOF

# direnvを許可（初回のみ）
direnv allow
```

**✅ これでディレクトリに入ると自動的に環境変数が設定されます！**

確認方法：
```bash
# ディレクトリに入る
cd services/bookmark

# 環境変数が設定されていることを確認
echo $DATABASE_URL
# 出力: postgresql://postgres:postgres@localhost:5432/bookmark_dev?sslmode=disable
```

### 3. 依存関係のインストール

```bash
go mod download
```

### 4. PostgreSQLの起動

```bash
docker-compose up -d postgres

# 起動確認
docker-compose ps
```

### 5. マイグレーション実行

```bash
cd ../../ops/db-migrator
go run main.go
cd -  # services/bookmark に戻る
```

### 6. アプリケーション起動

```bash
go run main.go
```

**期待される出力:**
```
[GIN-debug] Listening and serving HTTP on :8080
```

### 7. 動作確認

別のターミナルで：

```bash
# ヘルスチェック
curl http://localhost:8080/up

# ブックマーク作成
curl -X POST http://localhost:8080/api/bookmarks \
  -H "Content-Type: application/json" \
  -d '{"url":"https://go.dev","tags":["go","programming"]}'

# 一覧取得
curl http://localhost:8080/api/bookmarks
```

---

## 📁 ディレクトリ構成

```
services/bookmark/
├── main.go                 # エントリーポイント
├── go.mod                  # Go依存関係管理
├── docker-compose.yml      # ローカルPostgreSQL
├── .envrc                  # 環境変数（direnv）
├── .env                    # 環境変数（godotenv、direnvの代替）
│
├── config/
│   └── config.go          # 設定読み込み
│
├── model/
│   ├── bookmark.go        # Bookmarkエンティティ
│   └── errors.go          # エラー定義
│
├── handler/               # Presentation Layer
│   ├── router.go          # ルーティング
│   ├── bookmark.go        # ブックマークハンドラー
│   └── health.go          # ヘルスチェック
│
├── service/               # Business Logic Layer
│   └── bookmark.go        # ブックマークサービス
│
├── repository/            # Data Access Layer
│   └── bookmark.go        # ブックマークリポジトリ
│
└── pkg/                   # 共通ユーティリティ
    └── database/
        └── postgres.go    # DB接続
```

---

## 🔧 開発コマンド

### データベース操作

```bash
# PostgreSQL起動
docker-compose up -d postgres

# PostgreSQL停止
docker-compose down

# PostgreSQLログ確認
docker-compose logs -f postgres

# PostgreSQLに接続
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev

# テーブル一覧確認
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "\dt"

# テーブル構造確認
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "\d bookmarks"
```

### マイグレーション

```bash
# マイグレーション実行
cd ../../ops/db-migrator && go run main.go && cd -

# マイグレーションロールバック
cd ../../ops/db-migrator && go run main.go -cmd down && cd -

# マイグレーションバージョン確認
cd ../../ops/db-migrator && go run main.go -cmd version && cd -
```

詳細は [ops/db-migrator/README.md](../../ops/db-migrator/README.md) を参照

### アプリケーション

```bash
# アプリケーション起動
go run main.go

# ビルド
go build -o bin/server .

# 実行
./bin/server
```

---

## 📋 API仕様

### エンドポイント

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/up` | ヘルスチェック |
| POST | `/api/bookmarks` | ブックマーク作成 |
| GET | `/api/bookmarks` | ブックマーク一覧 |
| GET | `/api/bookmarks/:id` | ブックマーク詳細 |
| DELETE | `/api/bookmarks/:id` | ブックマーク削除 |
| GET | `/api/bookmarks/search` | ブックマーク検索 |

### リクエスト例

**ブックマーク作成:**
```bash
curl -X POST http://localhost:8080/api/bookmarks \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://go.dev",
    "tags": ["go", "programming"]
  }'
```

**タグで検索:**
```bash
curl "http://localhost:8080/api/bookmarks/search?tags=go"
```

**キーワードで検索:**
```bash
curl "http://localhost:8080/api/bookmarks/search?q=golang"
```

---

## 🔐 環境変数

### direnvを使用する場合（推奨）

`.envrc` ファイル:
```bash
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/bookmark_dev?sslmode=disable"
export APP_ENV="development"
```

ディレクトリに入ると自動的に環境変数が設定されます。

### direnvを使用しない場合

`.env` ファイルを作成:
```bash
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/bookmark_dev?sslmode=disable
APP_ENV=development
```

`godotenv`が自動的に読み込みます。

---

## 🐛 トラブルシューティング

### エラー: `bind: address already in use`

**原因:** ポート8080が既に使用されている

**解決策:**
```bash
# 使用中のプロセスを確認
lsof -i :8080

# プロセスを終了
kill -9 <PID>
```

### エラー: `connection refused`

**原因:** PostgreSQLが起動していない

**解決策:**
```bash
docker-compose up -d postgres
```

### エラー: `no such table: bookmarks`

**原因:** マイグレーションが実行されていない

**解決策:**
```bash
cd ../../ops/db-migrator && go run main.go && cd -
```

### direnvが動作しない

**症状:** ディレクトリに入っても環境変数が設定されない

**解決策:**
```bash
# シェルの設定を確認
cat ~/.zshrc | grep direnv

# 設定がない場合は追加
echo 'eval "$(direnv hook zsh)"' >> ~/.zshrc
source ~/.zshrc

# direnvを許可
direnv allow
```

---

## 📚 関連ドキュメント

- [実装ガイド](../../docs/implementation-guide.md)
- [三層アーキテクチャガイド](../../docs/architecture/go-three-layer-architecture.md)
- [マイグレーションガイド](../../ops/db-migrator/README.md)

---

## 🎯 次のステップ

基本的なCRUD操作が動作したら、次は以下を実装します：

- **Phase 3:** SQSワーカー実装（OGP取得）
- **Phase 4:** バッチ処理実装（OGP更新、リンク切れチェック）
- **Phase 5:** ECSデプロイ

詳細は [実装ガイド](../../docs/implementation-guide.md) を参照してください。
