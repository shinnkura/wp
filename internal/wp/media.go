package wp

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func UploadFeaturedImage(client *Client, imagePath string) (int, error) {
	imageData, err := os.ReadFile(fmt.Sprintf("internal/images/%s", imagePath))
	if err != nil {
		return 0, fmt.Errorf("画像ファイル読み取りエラー: %v", err)
	}

	// マルチパートフォームデータを作成
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(imagePath))
	if err != nil {
		return 0, err
	}
	part.Write(imageData)
	writer.Close()

	// メディアアップロードのリクエストを作成
	url := client.BaseURL + "/wp-json/wp/v2/media"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Basic "+client.BasicAuth)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// リクエストを送信
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// レスポンスを処理
	var mediaResp MediaResponse
	if err := client.decodeResponse(resp, &mediaResp); err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("画像アップロードエラー: %d", resp.StatusCode)
	}

	return mediaResp.ID, nil
}

