package wp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func CreateTag(client *Client, name string) (*Tag, error) {
	tagReq := CreateTagRequest{
		Name: name,
	}

	jsonData, err := json.Marshal(tagReq)
	if err != nil {
		return nil, fmt.Errorf("JSON変換エラー: %v", err)
	}

	url := client.BaseURL + "/wp-json/wp/v2/tags"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Basic "+client.BasicAuth)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// エラーレスポンスの詳細を取得
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("APIエラー: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("APIエラー: %s - %s", errorResp.Code, errorResp.Message)
	}

	var tag Tag
	if err := json.NewDecoder(resp.Body).Decode(&tag); err != nil {
		return nil, err
	}

	return &tag, nil
}

func GetTagID(client *Client, tagName string) (int, error) {
	url := client.BaseURL + "/wp-json/wp/v2/tags"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Basic "+client.BasicAuth)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var tags []Tag
	if err := client.decodeResponse(resp, &tags); err != nil {
		return 0, err
	}

	// 既存のタグを検索
	for _, tag := range tags {
		if tag.Name == tagName {
			return tag.ID, nil
		}
	}

	// タグが見つからない場合は新規作成
	newTag, err := CreateTag(client, tagName)
	if err != nil {
		return 0, fmt.Errorf("タグ作成エラー: %v", err)
	}

	return newTag.ID, nil
}

func GetTagIDs(client *Client, tagNames []string) ([]int, error) {
	url := client.BaseURL + "/wp-json/wp/v2/tags?per_page=100"  // 取得件数を増やす
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Basic "+client.BasicAuth)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tags []Tag
	if err := client.decodeResponse(resp, &tags); err != nil {
		return nil, err
	}

	var tagIDs []int
	for _, name := range tagNames {
		found := false
		for _, tag := range tags {
			if strings.EqualFold(tag.Name, name) {  // 大文字小文字を区別しない比較
				tagIDs = append(tagIDs, tag.ID)
				found = true
				break
			}
		}

		if !found {
			// タグが存在しない場合は新規作成
			newTag, err := CreateTag(client, name)
			if err != nil {
				// 作成に失敗した場合は、もう一度検索を試みる
				for _, tag := range tags {
					if strings.EqualFold(tag.Name, name) {
						tagIDs = append(tagIDs, tag.ID)
						found = true
						break
					}
				}
				if !found {
					return nil, fmt.Errorf("タグ作成エラー: %v", err)
				}
			} else {
				tagIDs = append(tagIDs, newTag.ID)
			}
		}
	}

	return tagIDs, nil
}
