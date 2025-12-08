# ビルドステージ
FROM golang:1.24.2-alpine AS builder

# golang-migrateのインストール
RUN apk add --no-cache git && \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.0

# 実行ステージ
FROM alpine:3.18

# 必要なランタイム依存関係をインストール
RUN apk add --no-cache ca-certificates

# ビルドステージからmigrateバイナリをコピー
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# マイグレーションファイルをコピー
WORKDIR /migrations
COPY db/mydb/migrations /migrations

# エントリーポイント
ENTRYPOINT ["migrate"]
