package wp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// UpdateMetadata はマークダウンファイルのメタデータを更新します
func UpdateMetadata(filename string, metadata ArticleMetadata) error {
	// .md拡張子を追加
	mdFilename := filename + ".md"

	// ファイルの内容を読み込む
	content, err := os.ReadFile(fmt.Sprintf("internal/articles/%s", mdFilename))
	if err != nil {
		return fmt.Errorf("ファイル読み取りエラー: %w", err)
	}

	// 新しいメタデータをJSON形式に変換（プリティプリント）
	newMetadata, err := json.MarshalIndent(metadata, "", "    ")
	if err != nil {
		return fmt.Errorf("メタデータのJSONパースエラー: %w", err)
	}

	// 本文を抽出
	parts := bytes.Split(content, []byte("\n---\n"))
	if len(parts) < 2 {
		// メタデータセクションが見つからない場合は、元のコンテンツを維持
		return fmt.Errorf("メタデータセクションが見つかりません。ファイル形式を確認してください: %s", mdFilename)
	}

	// 本文部分を結合（複数の---がある場合に対応）
	body := bytes.Join(parts[1:], []byte("\n---\n"))

	// 新しいファイルの内容を構築
	var newContent bytes.Buffer
	newContent.Write(newMetadata)
	newContent.WriteString("\n\n---\n")
	newContent.Write(body)

	// ファイルに書き込む
	err = os.WriteFile(fmt.Sprintf("internal/articles/%s", mdFilename), newContent.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("ファイル書き込みエラー: %w", err)
	}

	return nil
}
