package wp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

	var tag Tag
	if err := client.decodeResponse(resp, &tag); err != nil {
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
	url := client.BaseURL + "/wp-json/wp/v2/tags"
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
			if tag.Name == name {
				tagIDs = append(tagIDs, tag.ID)
				found = true
				break
			}
		}

		if !found {
			// タグが存在しない場合は新規作成
			newTag, err := CreateTag(client, name)
			if err != nil {
				return nil, fmt.Errorf("タグ作成エラー: %v", err)
			}
			tagIDs = append(tagIDs, newTag.ID)
		}
	}

	return tagIDs, nil
}
