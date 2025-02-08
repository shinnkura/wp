package wp

import (
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
		for _, cat := range categories {
			if cat.Name == name {
				categoryIDs = append(categoryIDs, cat.ID)
				break
			}
		}
	}

	return categoryIDs, nil
}