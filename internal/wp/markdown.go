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

	// コードブロック内のテキストを一時的に保存
	codeBlocks := make(map[string]string)
	placeholder := "CODE_BLOCK_PLACEHOLDER_%d"
	blockCount := 0

	// コードブロック変換（トリプルバッククォート）を処理し、プレースホルダーで置き換え
	html = regexp.MustCompile("(?s)```(.*?)\n(.*?)```").ReplaceAllStringFunc(html, func(match string) string {
		re := regexp.MustCompile("(?s)```(.*?)\n(.*?)```")
		parts := re.FindStringSubmatch(match)
		if len(parts) == 3 {
			lang := strings.TrimSpace(parts[1])
			code := strings.TrimSpace(parts[2])

			// HTMLエスケープ処理
			code = strings.ReplaceAll(code, "<", "&lt;")
			code = strings.ReplaceAll(code, ">", "&gt;")

			currentPlaceholder := fmt.Sprintf(placeholder, blockCount)

			// 新しいテンプレートを使用
			codeBlock := fmt.Sprintf(`<div class="code-wrap"><button class="code-copy"><i class="fa fa-copy"></i></button><pre class="wp-block-code hljs %s"><code>%s</code></pre></div>`, lang, code)
			codeBlocks[currentPlaceholder] = codeBlock

			blockCount++
			return currentPlaceholder
		}
		return match
	})

	// インラインコードを一時的に保存
	inlineCodeBlocks := make(map[string]string)
	inlinePlaceholder := "INLINE_CODE_PLACEHOLDER_%d"
	inlineCount := 0

	html = regexp.MustCompile("`([^`]+)`").ReplaceAllStringFunc(html, func(match string) string {
		content := regexp.MustCompile("`([^`]+)`").FindStringSubmatch(match)[1]
		content = strings.ReplaceAll(content, "<", "&lt;")
		content = strings.ReplaceAll(content, ">", "&gt;")
		currentPlaceholder := fmt.Sprintf(inlinePlaceholder, inlineCount)
		inlineCodeBlocks[currentPlaceholder] = "<code>" + content + "</code>"
		inlineCount++
		return currentPlaceholder
	})

	// テーブル変換を処理
	html = convertTables(html)

	// 水平線変換
	html = regexp.MustCompile(`(?m)^---\n`).ReplaceAllString(html, "<hr>\n")

	// 画像変換
	html = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`).ReplaceAllString(html, "<img src=\"$2\" alt=\"$1\">")

	// 見出し変換
	html = regexp.MustCompile(`(?m)^# (.+)$`).ReplaceAllString(html, "<h1>$1</h1>")
	html = regexp.MustCompile(`(?m)^## (.+)$`).ReplaceAllString(html, "<h2>$1</h2>")
	html = regexp.MustCompile(`(?m)^### (.+)$`).ReplaceAllString(html, "<h3>$1</h3>")
	html = regexp.MustCompile(`(?m)^#### (.+)$`).ReplaceAllString(html, "<h4>$1</h4>")

	// リンク変換
	html = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(html, "<a href=\"$2\">$1</a>")

	// 箇条書き変換
	html = processListItems(html)

	// 段落タグの処理
	paragraphs := strings.Split(html, "\n\n")
	for i, p := range paragraphs {
		if !strings.Contains(p, "CODE_BLOCK_PLACEHOLDER") && strings.TrimSpace(p) != "" {
			paragraphs[i] = "<p>" + strings.TrimSpace(p) + "</p>"
		}
	}
	html = strings.Join(paragraphs, "\n")

	// 太字変換
	html = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(html, "<strong>$1</strong>")

	// プレースホルダーを元のコードブロックに置き換え
	for placeholder, codeBlock := range codeBlocks {
		html = strings.Replace(html, placeholder, codeBlock, 1)
	}

	// インラインコードのプレースホルダーを置き換え
	for placeholder, code := range inlineCodeBlocks {
		html = strings.Replace(html, placeholder, code, 1)
	}

	return html
}

// convertTables はMarkdownのテーブルをHTMLテーブルに変換します
func convertTables(markdown string) string {
	lines := strings.Split(markdown, "\n")
	var result strings.Builder
	inTable := false
	var tableRows []string

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// テーブル行の判定
		if strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") {
			if !inTable {
				inTable = true
				tableRows = make([]string, 0)
			}
			tableRows = append(tableRows, line)
		} else if inTable {
			// テーブルの終了
			if len(tableRows) > 0 {
				result.WriteString(processTable(tableRows))
			}
			inTable = false
			result.WriteString(line + "\n")
		} else {
			result.WriteString(lines[i] + "\n")
		}
	}

	// 最後のテーブルの処理
	if inTable && len(tableRows) > 0 {
		result.WriteString(processTable(tableRows))
	}

	return result.String()
}

// processTable は収集したテーブル行をHTMLテーブルに変換します
func processTable(rows []string) string {
	if len(rows) < 3 {
		return strings.Join(rows, "\n") + "\n"
	}

	var result strings.Builder
	result.WriteString("<table class=\"wp-table\">\n")

	// ヘッダー行の処理
	headerCells := splitTableRow(rows[0])
	result.WriteString("<thead>\n<tr>\n")
	for _, cell := range headerCells {
		// 太字(**) の処理
		cell = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(cell, "$1")
		result.WriteString("<th>" + strings.TrimSpace(cell) + "</th>\n")
	}
	result.WriteString("</tr>\n</thead>\n")

	// 区切り行をスキップ

	// ボディ行の処理
	result.WriteString("<tbody>\n")
	for i := 2; i < len(rows); i++ {
		cells := splitTableRow(rows[i])
		result.WriteString("<tr>\n")
		for _, cell := range cells {
			// 太字(**) の処理
			cell = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(cell, "$1")
			result.WriteString("<td>" + strings.TrimSpace(cell) + "</td>\n")
		}
		result.WriteString("</tr>\n")
	}
	result.WriteString("</tbody>\n")
	result.WriteString("</table>\n")

	return result.String()
}

// splitTableRow はテーブル行を個々のセルに分割します
func splitTableRow(row string) []string {
	// 先頭と末尾の | を削除
	row = strings.Trim(row, "|")
	// セルを分割
	cells := strings.Split(row, "|")
	// 各セルをトリム
	for i, cell := range cells {
		cells[i] = strings.TrimSpace(cell)
	}
	return cells
}
