package wp

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func ReadArticleFromMd(filename string) (ArticleMetadata, string, error) {
	content, err := os.ReadFile(fmt.Sprintf("internal/articles/%s.md", filename))
	if err != nil {
		return ArticleMetadata{}, "", fmt.Errorf("ファイル読み取りエラー: %v", err)
	}

	// JSONメタデータと本文を分離
	parts := bytes.SplitN(content, []byte("\n---\n"), 2)
	if len(parts) != 2 {
		return ArticleMetadata{}, "", fmt.Errorf("ファイルフォーマットが不正です。JSONメタデータと本文を'---'で区切ってください")
	}

	// JSONメタデータをパース
	var metadata ArticleMetadata
	if err := json.Unmarshal(parts[0], &metadata); err != nil {
		return ArticleMetadata{}, "", fmt.Errorf("メタデータのJSONパースエラー: %v", err)
	}

	// 本文を取得
	return metadata, string(parts[1]), nil
}

func ConvertMarkdownToHTML(markdown string) string {
	html := markdown

	// 画像変換を修正（内部画像の場合はBase64エンコードした画像を使用）
	html = regexp.MustCompile(`!\[([^\]]*)\]\(internal/([^)]+)\)`).ReplaceAllStringFunc(html, func(match string) string {
		re := regexp.MustCompile(`!\[([^\]]*)\]\(internal/([^)]+)\)`)
		matches := re.FindStringSubmatch(match)
		if len(matches) == 3 {
			alt := matches[1]
			path := matches[2]
			imageData, err := readImageFile(path)
			if err == nil {
				return fmt.Sprintf(`<img src="data:image/jpeg;base64,%s" alt="%s">`, imageData, alt)
			}
		}
		return match
	})

	// 外部画像のための既存の変換ルールを維持
	html = regexp.MustCompile(`!\[([^\]]*)\]\(http[^)]+\)`).ReplaceAllString(html, "<img src=\"$2\" alt=\"$1\">")

	// 見出し変換
	html = regexp.MustCompile(`(?m)^# (.+)$`).ReplaceAllString(html, "<h1>$1</h1>")
	html = regexp.MustCompile(`(?m)^## (.+)$`).ReplaceAllString(html, "<h2>$1</h2>")
	html = regexp.MustCompile(`(?m)^### (.+)$`).ReplaceAllString(html, "<h3>$1</h3>")

	// リンク変換
	html = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(html, "<a href=\"$2\">$1</a>")

	// 改行変換
	html = strings.ReplaceAll(html, "\n\n", "</p><p>")
	html = "<p>" + html + "</p>"

	// コードブロック変換
	html = regexp.MustCompile("```([^`]+)```").ReplaceAllString(html, "<pre><code>$1</code></pre>")

	// 太字変換
	html = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(html, "<strong>$1</strong>")

	return html
}

func readImageFile(imagePath string) (string, error) {
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("画像ファイル読み取りエラー: %v", err)
	}
	return base64.StdEncoding.EncodeToString(imageData), nil
}
