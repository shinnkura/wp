package wp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func GetCategoryIDs(client *Client, categoryNames []string) ([]int, error) {
	url := client.BaseURL + "/wp-json/wp/v2/categories"
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

	var categories []Category
	if err := client.decodeResponse(resp, &categories); err != nil {
		return nil, err
	}

	var categoryIDs []int
	for _, name := range categoryNames {
		found := false
		for _, cat := range categories {
			if cat.Name == name {
				categoryIDs = append(categoryIDs, cat.ID)
				found = true
				break
			}
		}

		if !found {
			// カテゴリーが存在しない場合は新規作成
			newCat, err := CreateCategory(client, name)
			if err != nil {
				return nil, fmt.Errorf("カテゴリー作成エラー: %v", err)
			}
			categoryIDs = append(categoryIDs, newCat.ID)
		}
	}

	return categoryIDs, nil
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
}

func CreateCategory(client *Client, name string) (*Category, error) {
	categoryReq := CreateCategoryRequest{
		Name: name,
	}

	jsonData, err := json.Marshal(categoryReq)
	if err != nil {
		return nil, fmt.Errorf("JSON変換エラー: %v", err)
	}

	url := client.BaseURL + "/wp-json/wp/v2/categories"
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

	var category Category
	if err := client.decodeResponse(resp, &category); err != nil {
		return nil, err
	}

	return &category, nil
}