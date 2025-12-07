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