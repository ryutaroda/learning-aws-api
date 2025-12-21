package http

import (
	"time"
    "github.com/go-resty/resty/v2"
)

func NewClient() *resty.Client {
    return resty.New().
        SetTimeout(10 * time.Second).
        SetRetryCount(3).
        SetRetryWaitTime(2 * time.Second).          // リトライ間隔: 2秒
        SetRetryMaxWaitTime(8 * time.Second).      // 最大待機時間も調整
        SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
            // 5xxエラーとタイムアウトのみリトライ
            if resp.StatusCode() >= 500 || resp.StatusCode() == 0 {
                return 0, nil // デフォルトのリトライロジックを使用
            }
            return -1, nil // 4xx系などはリトライしない
        }).
        SetHeader("User-Agent", "Mozilla/5.0 (compatible; BookmarkBot/1.0)")
}

