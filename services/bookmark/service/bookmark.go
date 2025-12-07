package service

import (
    "context"

    "bookmark/model"
    "bookmark/repository"
)

type BookmarkService struct {
    bookmarkRepo *repository.BookmarkRepository
}

func NewBookmarkService(bookmarkRepo *repository.BookmarkRepository) *BookmarkService {
    return &BookmarkService{
        bookmarkRepo: bookmarkRepo,
    }
}

// Create ブックマーク作成
func (s *BookmarkService) Create(ctx context.Context, url string, tags []string) (*model.Bookmark, error) {
    bookmark := &model.Bookmark{
        URL:    url,
        Status: "pending",
        Tags:   tags,
    }

    if err := s.bookmarkRepo.Save(ctx, bookmark); err != nil {
        return nil, err
    }

    return bookmark, nil
}

// GetByID ブックマーク取得
func (s *BookmarkService) GetByID(ctx context.Context, id uint) (*model.Bookmark, error) {
    return s.bookmarkRepo.FindByID(ctx, id)
}

// GetAll 全ブックマーク取得
func (s *BookmarkService) GetAll(ctx context.Context) ([]model.Bookmark, error) {
    return s.bookmarkRepo.FindAll(ctx)
}

// Delete ブックマーク削除
func (s *BookmarkService) Delete(ctx context.Context, id uint) error {
    return s.bookmarkRepo.Delete(ctx, id)
}

// Search 検索（タイトル・タグ）
func (s *BookmarkService) Search(ctx context.Context, query string, tags []string) ([]model.Bookmark, error) {
    return s.bookmarkRepo.Search(ctx, query, tags)
}