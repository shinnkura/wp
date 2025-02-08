package main

import (
	// "bytes"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

type PostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Status  string `json:"status"`
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

// Markdownファイルから記事内容を読み取る関数
func readArticleFromMd(filename string) (string, string, error) {
	content, err := os.ReadFile(fmt.Sprintf("internal/%s.md", filename))
	if err != nil {
		return "", "", fmt.Errorf("ファイル読み取りエラー: %v", err)
	}

	lines := bytes.Split(content, []byte("\n"))
	if len(lines) < 2 {
		return "", "", fmt.Errorf("ファイルフォーマットが不正です")
	}

	// 1行目をタイトルとして扱う
	title := string(bytes.TrimPrefix(lines[0], []byte("# ")))

	// 残りの内容を本文として扱う
	body := string(bytes.Join(lines[1:], []byte("\n")))

	return title, body, nil
}

// 新しい関数を追加
func convertMarkdownToHTML(markdown string) string {
	// 基本的なMarkdown→HTML変換ルール
	html := markdown

	// 見出し変換
	html = regexp.MustCompile(`(?m)^# (.+)$`).ReplaceAllString(html, "<h1>$1</h1>")
	html = regexp.MustCompile(`(?m)^## (.+)$`).ReplaceAllString(html, "<h2>$1</h2>")
	html = regexp.MustCompile(`(?m)^### (.+)$`).ReplaceAllString(html, "<h3>$1</h3>")

	// 画像変換
	html = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`).ReplaceAllString(html, "<img src=\"$2\" alt=\"$1\">")

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
	title, content, err := readArticleFromMd("article1") // internal/article1.md を読み込む
	if err != nil {
		fmt.Printf("記事読み取りエラー: %v\n", err)
		return
	}

	// 投稿するコンテンツを準備
	post := PostRequest{
		Title:   title,
		Content: convertMarkdownToHTML(content), // Markdownを変換
		Status:  "publish", // 下書き: draft
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
