package main

import (
	"flag"
	"fmt"
	"os"

	"wp/internal/wp"

	"github.com/joho/godotenv"
)

func main() {
	// コマンドライン引数の解析
	flag.Parse()
	args := flag.Args()

	if len(args) != 2 {
		fmt.Println("使用方法: go run cmd/cli [command] [マークダウンファイル名]")
		fmt.Println("例: go run cmd/cli create article1")
		fmt.Println("    go run cmd/cli update article1")
		os.Exit(1)
	}

	command := args[0]
	filename := args[1]

	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		return
	}

	client := wp.NewClient(
		os.Getenv("WP_URL"),
		os.Getenv("USER_NAME"),
		os.Getenv("USER_PASSWORD"),
	)

	// 指定されたファイル名の記事を読み込む
	metadata, content, err := wp.ReadArticleFromMd(filename)
	if err != nil {
		fmt.Printf("記事読み取りエラー: %v\n", err)
		return
	}

	content, err = wp.ExtractAndUploadImages(client, content)
	if err != nil {
		fmt.Printf("画像アップロードエラー: %v\n", err)
		return
	}

	categoryIDs, err := wp.GetCategoryIDs(client, metadata.Category)
	if err != nil {
		fmt.Printf("カテゴリーID取得エラー: %v\n", err)
		return
	}

	var mediaID int
	if metadata.Image != "" {
		mediaID, err = wp.UploadFeaturedImage(client, metadata.Image)
		if err != nil {
			fmt.Printf("画像アップロードエラー: %v\n", err)
			return
		}
	}

	tagIDs, err := wp.GetTagIDs(client, metadata.Tag)
	if err != nil {
		fmt.Printf("タグID取得エラー: %v\n", err)
		return
	}

	post := wp.PostRequest{
		Title:         metadata.Title,
		Content:       wp.ConvertMarkdownToHTML(content),
		Status:        "publish",
		Slug:          metadata.Permalink,
		Categories:    categoryIDs,
		Tags:          tagIDs,
		FeaturedMedia: mediaID,
	}

	var resp *wp.PostResponse
	switch command {
	case "create":
		resp, err = client.CreatePost(post)
		if err != nil {
			fmt.Printf("投稿エラー: %v\n", err)
			return
		}
		// メタデータにpost_idを追加して保存
		metadata.PostID = resp.ID
		if err := wp.UpdateMetadata(filename, metadata); err != nil {
			fmt.Printf("メタデータ更新エラー: %v\n", err)
			return
		}
	case "update":
		if metadata.PostID == 0 {
			fmt.Println("エラー: この記事はまだ投稿されていません")
			return
		}
		resp, err = client.UpdatePost(metadata.PostID, post)
		if err != nil {
			fmt.Printf("更新エラー: %v\n", err)
			return
		}
	default:
		fmt.Printf("不正なコマンド: %s\n", command)
		return
	}

	fmt.Printf("操作が成功しました。投稿ID: %d\n", resp.ID)
	fmt.Printf("投稿URL: %s\n", resp.Link)
}
