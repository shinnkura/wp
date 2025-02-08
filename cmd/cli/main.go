package main

import (
	"fmt"
	"os"

	"wp/internal/wp"

	"github.com/joho/godotenv"
)

func main() {
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

	metadata, content, err := wp.ReadArticleFromMd("article1")
	if err != nil {
		fmt.Printf("記事読み取りエラー: %v\n", err)
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

	post := wp.PostRequest{
		Title:         metadata.Title,
		Content:       wp.ConvertMarkdownToHTML(content),
		Status:        "publish",
		Slug:          metadata.Permalink,
		Categories:    categoryIDs,
		FeaturedMedia: mediaID,
	}

	resp, err := client.CreatePost(post)
	if err != nil {
		fmt.Printf("投稿エラー: %v\n", err)
		return
	}

	fmt.Printf("投稿が成功しました。投稿ID: %d\n", resp.ID)
	fmt.Printf("投稿URL: %s\n", resp.Link)
}
