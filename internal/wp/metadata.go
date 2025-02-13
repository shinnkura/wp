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
	if len(parts) != 2 {
		return fmt.Errorf("メタデータセクションが見つかりません")
	}
	body := parts[1]

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
