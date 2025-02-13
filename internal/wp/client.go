package wp

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	BaseURL    string
	BasicAuth  string
	HTTPClient *http.Client
}

func NewClient(baseURL, username, password string) *Client {
	auth := username + ":" + password
	basicAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	return &Client{
		BaseURL:    baseURL,
		BasicAuth:  basicAuth,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) CreatePost(post PostRequest) (*PostResponse, error) {
	jsonData, err := json.Marshal(post)
	if err != nil {
		return nil, fmt.Errorf("JSON変換エラー: %v", err)
	}

	url := c.BaseURL + "/wp-json/wp/v2/posts"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Basic "+c.BasicAuth)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var postResp PostResponse
	if err := c.decodeResponse(resp, &postResp); err != nil {
		return nil, err
	}

	return &postResp, nil
}

func (c *Client) UpdatePost(postID int, post PostRequest) (*PostResponse, error) {
	jsonData, err := json.Marshal(post)
	if err != nil {
		return nil, fmt.Errorf("JSON変換エラー: %v", err)
	}

	url := fmt.Sprintf("%s/wp-json/wp/v2/posts/%d", c.BaseURL, postID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Basic "+c.BasicAuth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-HTTP-Method-Override", "PUT")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var postResp PostResponse
	if err := c.decodeResponse(resp, &postResp); err != nil {
		return nil, err
	}

	return &postResp, nil
}

func (c *Client) decodeResponse(resp *http.Response, v interface{}) error {
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return fmt.Errorf("APIエラー: %d", resp.StatusCode)
		}
		return fmt.Errorf("APIエラー: %s - %s", errorResp.Code, errorResp.Message)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}
