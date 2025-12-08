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