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

// 箇条書き変換（インデント対応）
func processListItems(html string) string {
	// 箇条書きの階層構造を保持する
	var currentIndent int
	var result strings.Builder

	lines := strings.Split(html, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if match := regexp.MustCompile(`^(\s*)- (.+)$`).FindStringSubmatch(line); match != nil {
			indent := len(match[1])
			content := match[2]

			// インデントレベルの変化に応じてタグを追加
			if indent > currentIndent {
				// インデントが深くなった場合、新しい<ul>を開始
				result.WriteString("<ul>")
			} else if indent < currentIndent {
				// インデントが浅くなった場合、必要な数だけ</ul>を追加
				for j := 0; j < (currentIndent-indent)/2; j++ {
					result.WriteString("</ul>")
				}
			}

			currentIndent = indent
			result.WriteString("<li>" + content + "</li>")
		} else {
			// 箇条書き以外の行の処理
			if currentIndent > 0 {
				// 箇条書きが終わった場合、必要な数だけ</ul>を追加
				for j := 0; j < currentIndent/2; j++ {
					result.WriteString("</ul>")
				}
				currentIndent = 0
			}
			result.WriteString(line + "\n")
		}
	}

	// 残りの</ul>タグを追加
	for j := 0; j < currentIndent/2; j++ {
		result.WriteString("</ul>")
	}

	return result.String()
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

			// HTMLエスケープ処理
			code = strings.ReplaceAll(code, "<", "&lt;")
			code = strings.ReplaceAll(code, ">", "&gt;")

			if lang == "" {
				return "\n<pre><code>" + code + "</code></pre>\n"
			}
			return "\n<pre><code class=\"language-" + lang + "\">" + code + "</code></pre>\n"
		}
		return match
	})

	// 水平線変換
	html = regexp.MustCompile(`(?m)^---\n`).ReplaceAllString(html, "<hr>\n")

	// 画像変換（WordPressにアップロード済みの画像URLを使用）
	html = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`).ReplaceAllString(html, "<img src=\"$2\" alt=\"$1\">")

	// 見出し変換
	html = regexp.MustCompile(`(?m)^# (.+)$`).ReplaceAllString(html, "<h1>$1</h1>")
	html = regexp.MustCompile(`(?m)^## (.+)$`).ReplaceAllString(html, "<h2>$1</h2>")
	html = regexp.MustCompile(`(?m)^### (.+)$`).ReplaceAllString(html, "<h3>$1</h3>")
	html = regexp.MustCompile(`(?m)^#### (.+)$`).ReplaceAllString(html, "<h4>$1</h4>")

	// リンク変換
	html = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(html, "<a href=\"$2\">$1</a>")

	// 箇条書き変換を新しい関数で処理
	html = processListItems(html)

	// インラインコード変換（シングルバッククォート）
	html = regexp.MustCompile("`([^`]+)`").ReplaceAllStringFunc(html, func(match string) string {
		// バッククォートの中身を取得
		content := regexp.MustCompile("`([^`]+)`").FindStringSubmatch(match)[1]

		// HTMLエスケープ処理
		content = strings.ReplaceAll(content, "<", "&lt;")
		content = strings.ReplaceAll(content, ">", "&gt;")

		return "<code>" + content + "</code>"
	})

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
