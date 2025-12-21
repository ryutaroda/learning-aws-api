package service

import (
    "context"
    "fmt"
    "net/http"
    "net/url"
    "strings"

    "github.com/PuerkitoBio/goquery"
    "github.com/go-resty/resty/v2"
    "bookmark/model"
)

type OgpService struct {
    httpClient *resty.Client
}

func NewOgpService(httpClient *resty.Client) *OgpService {
    return &OgpService{httpClient: httpClient}
}

// Fetch OGP情報を取得
func (s *OgpService) Fetch(ctx context.Context, targetURL string) (*model.OgpInfo, error) {
    resp, err := s.httpClient.R().
        SetContext(ctx).
        Get(targetURL)
    
    if err != nil {
        return nil, err
    }

    if resp.StatusCode() != http.StatusOK {
        return nil, fmt.Errorf("HTTP %d", resp.StatusCode())
    }

    doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
    if err != nil {
        return nil, err
    }

    ogp := &model.OgpInfo{}

    // OGPタグから取得
    doc.Find("meta[property^='og:']").Each(func(i int, sel *goquery.Selection) {
        prop, _ := sel.Attr("property")
        content, _ := sel.Attr("content")
        
        switch prop {
        case "og:title":
            ogp.Title = content
        case "og:description":
            ogp.Description = content
        case "og:image":
            ogp.ImageURL = s.resolveURL(targetURL, content)
        }
    })

    // フォールバック: titleタグ
    if ogp.Title == "" {
        ogp.Title = doc.Find("title").Text()
    }

    // favicon取得
    faviconURL := doc.Find("link[rel='icon']").AttrOr("href", "")
    if faviconURL == "" {
        faviconURL = doc.Find("link[rel='shortcut icon']").AttrOr("href", "")
    }
    if faviconURL != "" {
        ogp.FaviconURL = s.resolveURL(targetURL, faviconURL)
    }

    return ogp, nil
}

// CheckURL URLが有効かチェック
func (s *OgpService) CheckURL(ctx context.Context, targetURL string) error {
    resp, err := s.httpClient.R().
        SetContext(ctx).
        Head(targetURL)
    
    if err != nil {
        return err
    }

    if resp.StatusCode() >= 400 {
        return fmt.Errorf("HTTP %d", resp.StatusCode())
    }

    return nil
}

// resolveURL 相対URLを絶対URLに変換
func (s *OgpService) resolveURL(baseURL, relativeURL string) string {
    u, err := url.Parse(relativeURL)
    if err != nil {
        return relativeURL
    }

    base, err := url.Parse(baseURL)
    if err != nil {
        return relativeURL
    }

    return base.ResolveReference(u).String()
}
