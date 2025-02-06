package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

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
	baseURL := os.Getenv("WP_URL") // "https://aichi.blog"

	// Basic認証のヘッダーを作成
	auth := username + ":" + password
	basicAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	// 投稿するコンテンツを準備
	post := PostRequest{
		Title:   "GoからのAPI投稿テスト",
		Content: "これはGo言語からWP REST APIを使用して投稿されたコンテンツです。",
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
