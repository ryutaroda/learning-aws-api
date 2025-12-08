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