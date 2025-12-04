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
    ID          uint       `json:"id"`
    URL         string     `json:"url"`
    Title       string     `json:"title"`
    Description string     `json:"description"`
    ImageURL    string     `json:"image_url"`
    FaviconURL  string     `json:"favicon_url"`
    Status      string     `json:"status"`
    Tags        []string   `json:"tags"`
    FetchedAt   *time.Time `json:"fetched_at"`
    CreatedAt   time.Time  `json:"created_at"`
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