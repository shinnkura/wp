package wp

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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

func ExtractAndUploadImages(client *Client, content string) (string, error) {
	re := regexp.MustCompile(`!\[([^\]]*)\]\(internal/images/([^)]+)\)`)

	result := re.ReplaceAllStringFunc(content, func(match string) string {
		matches := re.FindStringSubmatch(match)
		if len(matches) >= 3 {
			alt := matches[1]
			imagePath := matches[2]

			mediaID, err := UploadFeaturedImage(client, imagePath)
			if err != nil {
				return match // エラーの場合は元のまま
			}

			// WordPressメディアのURLを取得
			url := fmt.Sprintf("%s/wp-json/wp/v2/media/%d", client.BaseURL, mediaID)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return match
			}

			req.Header.Set("Authorization", "Basic "+client.BasicAuth)
			resp, err := client.HTTPClient.Do(req)
			if err != nil {
				return match
			}
			defer resp.Body.Close()

			var mediaResp MediaResponse
			if err := client.decodeResponse(resp, &mediaResp); err != nil {
				return match
			}

			return fmt.Sprintf("![%s](%s)", alt, mediaResp.URL)
		}
		return match
	})

	return result, nil
}

