package wp

import (
	"bytes"
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

	// コードブロック変換（トリプルバッククォート）を最初に処理
	html = regexp.MustCompile("(?s)```(.*?)\n(.*?)```").ReplaceAllStringFunc(html, func(match string) string {
		re := regexp.MustCompile("(?s)```(.*?)\n(.*?)```")
		parts := re.FindStringSubmatch(match)
		if len(parts) == 3 {
			lang := strings.TrimSpace(parts[1])
			code := strings.TrimSpace(parts[2])
			if lang == "" {
				return "\n<pre><code>" + code + "</code></pre>\n"
			}
			return "\n<pre><code class=\"language-" + lang + "\">" + code + "</code></pre>\n"
		}
		return match
	})

	// 画像変換（WordPressにアップロード済みの画像URLを使用）
	html = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`).ReplaceAllString(html, "<img src=\"$2\" alt=\"$1\">")

	// 見出し変換
	html = regexp.MustCompile(`(?m)^# (.+)$`).ReplaceAllString(html, "<h1>$1</h1>")
	html = regexp.MustCompile(`(?m)^## (.+)$`).ReplaceAllString(html, "<h2>$1</h2>")
	html = regexp.MustCompile(`(?m)^### (.+)$`).ReplaceAllString(html, "<h3>$1</h3>")
	html = regexp.MustCompile(`(?m)^#### (.+)$`).ReplaceAllString(html, "<h4>$1</h4>")

	// リンク変換
	html = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(html, "<a href=\"$2\">$1</a>")

	// 箇条書き変換
	html = regexp.MustCompile(`(?m)^- (.+)$`).ReplaceAllStringFunc(html, func(match string) string {
		items := regexp.MustCompile(`(?m)^- (.+)$`).FindAllStringSubmatch(match, -1)
		if len(items) > 0 {
			return "<ul><li>" + items[0][1] + "</li></ul>"
		}
		return match
	})
	// 連続する</ul><ul>を削除
	html = regexp.MustCompile(`</ul>\s*<ul>`).ReplaceAllString(html, "")

	// インラインコード変換（シングルバッククォート）
	html = regexp.MustCompile("`([^`]+)`").ReplaceAllString(html, "<code>$1</code>")

	// 段落タグの処理を最後に行う
	// コードブロックを段落から除外
	paragraphs := strings.Split(html, "\n\n")
	for i, p := range paragraphs {
		if !strings.Contains(p, "<pre>") && !strings.Contains(p, "</pre>") && strings.TrimSpace(p) != "" {
			paragraphs[i] = "<p>" + strings.TrimSpace(p) + "</p>"
		}
	}
	html = strings.Join(paragraphs, "\n")

	// 太字変換
	html = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(html, "<strong>$1</strong>")

	return html
}
