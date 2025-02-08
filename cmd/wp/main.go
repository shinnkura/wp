package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

type PostRequest struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Status     string   `json:"status"`
	Slug       string   `json:"slug"`
	Categories []int    `json:"categories"`
	FeaturedMedia int    `json:"featured_media"`
}

type PostResponse struct {
	ID      int    `json:"id"`
	Link    string `json:"link"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type Post struct {
	ID    int    `json:"id"`
	Link  string `json:"link"`
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
}

type ArticleMetadata struct {
	Title     string   `json:"Title"`
	Image     string   `json:"Image"`
	Permalink string   `json:"Permalink"`
	Tag       string   `json:"Tag"`
	Category  []string `json:"Category"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type MediaResponse struct {
	ID  int    `json:"id"`
	URL string `json:"source_url"`
}

// Markdownファイルから記事内容を読み取る関数
func readArticleFromMd(filename string) (ArticleMetadata, string, error) {
	content, err := os.ReadFile(fmt.Sprintf("internal/article/%s.md", filename))
	if err != nil {
		return ArticleMetadata{}, "", fmt.Errorf("ファイル読み取りエラー: %v", err)
	}

	// JSONメタデータと本文を分離
	parts := bytes.SplitN(content, []byte("\n---\n"), 2)
	if len(parts) != 2 {
		return ArticleMetadata{}, "", fmt.Errorf("ファイルフォーマットが不正です。JSONメタデータと本文を'---'で区切ってください")
	}

	// JSONメタデータをパース
	var metadata ArticleMetadata
	if err := json.Unmarshal(parts[0], &metadata); err != nil {
		return ArticleMetadata{}, "", fmt.Errorf("メタデータのJSONパースエラー: %v", err)
	}

	// 本文を取得
	body := string(parts[1])

	return metadata, body, nil
}

// 画像ファイルを読み込んでBase64エンコードする関数を追加
func readImageFile(imagePath string) (string, error) {
	// internal/images/からの相対パスで画像を読み込む
	imageData, err := os.ReadFile(fmt.Sprintf("internal/%s", imagePath))
	if err != nil {
		return "", fmt.Errorf("画像ファイル読み取りエラー: %v", err)
	}

	// Base64エンコード
	return base64.StdEncoding.EncodeToString(imageData), nil
}

// 新しい関数を追加
func convertMarkdownToHTML(markdown string) string {
	html := markdown

	// 画像変換を修正（内部画像の場合はBase64エンコードした画像を使用）
	html = regexp.MustCompile(`!\[([^\]]*)\]\(internal/([^)]+)\)`).ReplaceAllStringFunc(html, func(match string) string {
		re := regexp.MustCompile(`!\[([^\]]*)\]\(internal/([^)]+)\)`)
		matches := re.FindStringSubmatch(match)
		if len(matches) == 3 {
			alt := matches[1]
			path := matches[2]
			imageData, err := readImageFile(path)
			if err == nil {
				return fmt.Sprintf(`<img src="data:image/jpeg;base64,%s" alt="%s">`, imageData, alt)
			}
		}
		return match
	})

	// 外部画像のための既存の変換ルールを維持
	html = regexp.MustCompile(`!\[([^\]]*)\]\(http[^)]+\)`).ReplaceAllString(html, "<img src=\"$2\" alt=\"$1\">")

	// 見出し変換
	html = regexp.MustCompile(`(?m)^# (.+)$`).ReplaceAllString(html, "<h1>$1</h1>")
	html = regexp.MustCompile(`(?m)^## (.+)$`).ReplaceAllString(html, "<h2>$1</h2>")
	html = regexp.MustCompile(`(?m)^### (.+)$`).ReplaceAllString(html, "<h3>$1</h3>")

	// リンク変換
	html = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(html, "<a href=\"$2\">$1</a>")

	// 改行変換
	html = strings.ReplaceAll(html, "\n\n", "</p><p>")
	html = "<p>" + html + "</p>"

	// コードブロック変換
	html = regexp.MustCompile("```([^`]+)```").ReplaceAllString(html, "<pre><code>$1</code></pre>")

	// 太字変換
	html = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(html, "<strong>$1</strong>")

	return html
}

func getCategoryIDs(baseURL, basicAuth string, categoryNames []string) ([]int, error) {
	// カテゴリー一覧を取得
	url := baseURL + "/wp-json/wp/v2/categories"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Basic "+basicAuth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var categories []Category
	if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
		return nil, err
	}

	// カテゴリー名からIDを検索
	var categoryIDs []int
	for _, name := range categoryNames {
		for _, cat := range categories {
			if cat.Name == name {
				categoryIDs = append(categoryIDs, cat.ID)
				break
			}
		}
	}

	return categoryIDs, nil
}

func uploadFeaturedImage(baseURL, basicAuth, imagePath string) (int, error) {
	fullPath := fmt.Sprintf("internal/images/%s", imagePath)
	imageData, err := os.ReadFile(fullPath)
	if err != nil {
		return 0, fmt.Errorf("画像ファイル読み取りエラー: %v", err)
	}

	// マルチパートフォームデータを作成
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "image.jpg")
	if err != nil {
		return 0, err
	}
	part.Write(imageData)
	writer.Close()

	// メディアアップロードのリクエストを作成
	url := baseURL + "/wp-json/wp/v2/media"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Basic "+basicAuth)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// リクエストを送信
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// レスポンスを処理
	var mediaResp MediaResponse
	if err := json.NewDecoder(resp.Body).Decode(&mediaResp); err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("画像アップロードエラー: %d", resp.StatusCode)
	}

	return mediaResp.ID, nil
}

func main() {
	// .envファイルを読み込む
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		return
	}

	// 環境変数から認証情報とURLを取得
	username := os.Getenv("USER_NAME")
	password := os.Getenv("USER_PASSWORD")
	baseURL := os.Getenv("WP_URL")

	// Basic認証のヘッダーを作成
	auth := username + ":" + password
	basicAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	// Markdownファイルから記事内容を読み取る
	metadata, content, err := readArticleFromMd("article1")
	if err != nil {
		fmt.Printf("記事読み取りエラー: %v\n", err)
		return
	}

	// カテゴリー名からIDを取得
	categoryIDs, err := getCategoryIDs(baseURL, basicAuth, metadata.Category)
	if err != nil {
		fmt.Printf("カテゴリーID取得エラー: %v\n", err)
		return
	}

	if metadata.Image != "" {
		// 画像をアップロード
		mediaID, err := uploadFeaturedImage(baseURL, basicAuth, metadata.Image)
		if err != nil {
			fmt.Printf("画像アップロードエラー: %v\n", err)
			return
		}

		// 投稿するコンテンツを準備
		post := PostRequest{
			Title:         metadata.Title,
			Content:       convertMarkdownToHTML(content),
			Status:        "publish",
			Slug:          metadata.Permalink,
			Categories:    categoryIDs,
			FeaturedMedia: mediaID,  // アップロードした画像のIDを設定
		}

		// JSONに変換
		jsonData, err := json.Marshal(post)
		if err != nil {
			fmt.Printf("Error marshaling JSON: %v\n", err)
			return
		}

		// リクエストを作成
		url := baseURL + "/wp-json/wp/v2/posts"
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}

		// ヘッダーを設定
		req.Header.Set("Authorization", "Basic "+basicAuth)
		req.Header.Set("Content-Type", "application/json")

		// リクエストを送信
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
			return
		}
		defer resp.Body.Close()

		// レスポンスを処理
		var postResp PostResponse
		if err := json.NewDecoder(resp.Body).Decode(&postResp); err != nil {
			fmt.Printf("Error decoding response: %v\n", err)
			return
		}

		if resp.StatusCode == http.StatusCreated {
			fmt.Printf("投稿が成功しました。投稿ID: %d\n", postResp.ID)
			fmt.Printf("投稿URL: %s\n", postResp.Link)
		} else {
			fmt.Printf("投稿に失敗しました。ステータスコード: %d\n", resp.StatusCode)
			fmt.Printf("エラーメッセージ: %s\n", postResp.Message)
		}
	}
}
