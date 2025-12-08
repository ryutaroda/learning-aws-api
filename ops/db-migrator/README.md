# DB Migrator

`golang-migrate` を使用したデータベースマイグレーションツールです。

## 📁 ディレクトリ構成

```
ops/db-migrator/
├── main.go                    # マイグレーション実行用のGoコード
├── go.mod                     # Go依存関係管理
├── go.sum
├── Dockerfile                 # Docker版マイグレーション用（オプション）
└── db/
    └── mydb/
        └── migrations/
            ├── 000001_create_bookmarks.up.sql    # マイグレーションUP
            └── 000001_create_bookmarks.down.sql  # マイグレーションDOWN
```

## 🚀 セットアップ

### 1. Go moduleを初期化（初回のみ）

```bash
cd /path/to/ops/db-migrator

# Go moduleを初期化
go mod init db-migrator

# 必要な依存関係を追加
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/postgres
go get -u github.com/golang-migrate/migrate/v4/source/file
go get -u github.com/lib/pq
```

### 2. PostgreSQLを起動

```bash
cd ../../services/bookmark
docker-compose up -d postgres
```

## 📋 マイグレーションファイルの作成

### 命名規則

```
{version}_{description}.{up|down}.sql
```

- `version`: 連番（例: 000001, 000002）
- `description`: マイグレーションの説明（例: create_bookmarks, add_user_id）
- `up`: テーブル作成・カラム追加などの適用
- `down`: upの逆操作（ロールバック）

### 例

```
db/mydb/migrations/
├── 000001_create_bookmarks.up.sql
├── 000001_create_bookmarks.down.sql
├── 000002_add_user_id_to_bookmarks.up.sql
└── 000002_add_user_id_to_bookmarks.down.sql
```

## ⚡ 基本的な使い方

### マイグレーションUP（全て適用）

```bash
cd /path/to/ops/db-migrator
go run main.go
```

### マイグレーションDOWN（全てロールバック）

```bash
go run main.go -cmd down
```

### 現在のバージョン確認

```bash
go run main.go -cmd version
```

## 🛠️ コマンドオプション

| オプション | デフォルト値 | 説明 |
|-----------|------------|------|
| `-path` | `db/mydb/migrations` | マイグレーションファイルのパス |
| `-database` | 環境変数 `DATABASE_URL` または `postgresql://postgres:postgres@localhost:5432/bookmark_dev?sslmode=disable` | データベース接続URL |
| `-cmd` | `up` | 実行するコマンド (`up` / `down` / `version`) |
| `-steps` | `-1`（全て） | マイグレーションのステップ数 |

## 📖 コマンド集

### ステップ指定でマイグレーション

```bash
# 1ステップだけUP
go run main.go -cmd up -steps 1

# 1ステップだけロールバック
go run main.go -cmd down -steps 1

# 2ステップだけUP
go run main.go -cmd up -steps 2
```

### データベースURL指定

```bash
# ローカルPostgreSQL
go run main.go -database "postgresql://postgres:postgres@localhost:5432/bookmark_dev?sslmode=disable"

# 本番環境（例）
go run main.go -database "postgresql://user:password@prod-db.example.com:5432/production?sslmode=require"
```

### 環境変数を使用

```bash
# 環境変数を設定
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/bookmark_dev?sslmode=disable"

# 実行（DATABASE_URLが自動的に使用される）
go run main.go
```

## 🐳 Docker版との比較

### Approach A: `go run main.go`（推奨：開発環境）

**メリット:**
- ローカル開発が簡単
- デバッグしやすい
- 即座に実行可能

**デメリット:**
- 本番環境への適用には工夫が必要

**使用シーン:**
- ローカル開発
- マイグレーションファイルの動作確認
- デバッグ

### Approach B: Docker Compose（推奨：本番/CI/CD）

**メリット:**
- CI/CD統合が簡単
- 環境の再現性が高い
- 本番環境と同じ構成

**デメリット:**
- ローカル実行が複雑
- ビルド時間がかかる

**使用シーン:**
- 本番環境デプロイ
- CI/CDパイプライン
- ECSタスクでの実行

**Docker版の実行方法:**
```bash
cd ../../services/bookmark
docker-compose up --build db-migrator
```

## 🗄️ データベース操作コマンド

### PostgreSQLコンテナに接続

```bash
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev
```

### テーブル一覧確認

```bash
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "\dt"
```

### テーブル構造確認

```bash
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "\d bookmarks"
```

### テーブル削除（クリーンアップ）

```bash
# bookmarksテーブル削除
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "DROP TABLE IF EXISTS bookmarks CASCADE;"

# マイグレーション履歴テーブル削除
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "DROP TABLE IF EXISTS schema_migrations CASCADE;"

# 全テーブル削除
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
```

### データ確認

```bash
# bookmarksテーブルの全データ確認
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "SELECT * FROM bookmarks;"

# マイグレーション履歴確認
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "SELECT * FROM schema_migrations;"
```

## 🔧 トラブルシューティング

### エラー: `relation "bookmarks" already exists`

**原因:** テーブルが既に存在している

**解決策:**
```bash
# 既存テーブルを削除
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "DROP TABLE IF EXISTS bookmarks CASCADE;"
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "DROP TABLE IF EXISTS schema_migrations CASCADE;"

# 再実行
go run main.go
```

### エラー: `no such file or directory`

**原因:** マイグレーションファイルのパスが間違っている

**解決策:**
```bash
# 現在のディレクトリ確認
pwd
# 出力: /path/to/ops/db-migrator

# マイグレーションファイルの存在確認
ls -la db/mydb/migrations/

# パスを指定して実行
go run main.go -path db/mydb/migrations
```

### エラー: `Dirty database version`

**原因:** 前回のマイグレーションが途中で失敗した

**解決策:**
```bash
# 現在のバージョン確認
go run main.go -cmd version

# schema_migrationsテーブルのdirtyフラグを修正
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "UPDATE schema_migrations SET dirty = false;"

# または、マイグレーションをやり直す
docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "DROP TABLE IF EXISTS schema_migrations CASCADE;"
go run main.go
```

### エラー: `connection refused`

**原因:** PostgreSQLが起動していない

**解決策:**
```bash
cd ../../services/bookmark
docker-compose up -d postgres

# 起動確認
docker-compose ps
```

## 📚 マイグレーション作成例

### 新しいテーブルを追加

**000002_create_users.up.sql:**
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

**000002_create_users.down.sql:**
```sql
DROP TABLE IF EXISTS users;
```

### カラムを追加

**000003_add_user_id_to_bookmarks.up.sql:**
```sql
ALTER TABLE bookmarks ADD COLUMN user_id INTEGER REFERENCES users(id);
CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);
```

**000003_add_user_id_to_bookmarks.down.sql:**
```sql
DROP INDEX IF EXISTS idx_bookmarks_user_id;
ALTER TABLE bookmarks DROP COLUMN IF EXISTS user_id;
```

## 🔗 関連ドキュメント

- [golang-migrate 公式ドキュメント](https://github.com/golang-migrate/migrate)
- [PostgreSQL ドキュメント](https://www.postgresql.org/docs/)
- [三層アーキテクチャガイド](../../docs/architecture/go-three-layer-architecture.md)
- [実装ガイド](../../docs/implementation-guide.md)

## 📝 開発フロー

### 新しいマイグレーションを追加する場合

1. マイグレーションファイルを作成
   ```bash
   # db/mydb/migrations/ に以下を作成
   # 000002_add_new_feature.up.sql
   # 000002_add_new_feature.down.sql
   ```

2. upファイルにSQLを記述
   ```sql
   -- 000002_add_new_feature.up.sql
   ALTER TABLE bookmarks ADD COLUMN priority INTEGER DEFAULT 0;
   ```

3. downファイルにロールバック用SQLを記述
   ```sql
   -- 000002_add_new_feature.down.sql
   ALTER TABLE bookmarks DROP COLUMN IF EXISTS priority;
   ```

4. マイグレーション実行
   ```bash
   go run main.go
   ```

5. 動作確認
   ```bash
   docker exec -it bookmark-postgres psql -U postgres -d bookmark_dev -c "\d bookmarks"
   ```

6. ロールバックのテスト
   ```bash
   go run main.go -cmd down -steps 1
   go run main.go -cmd up -steps 1
   ```
