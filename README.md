# WordPress 記事管理 CLI ツール

WordPress の記事をマークダウンファイルで管理し、API を通じて投稿・更新するための CLI ツールです。

## 機能

- マークダウンファイルから WordPress への記事投稿
- 既存記事の更新
- 画像の自動アップロード
- カテゴリーとタグの自動作成

## プロジェクト構成

```
project-root/
├── cmd/ # CLI ツールのエントリーポイント
│ └── cli/
│ └── main.go
├── internal/ # 内部パッケージ
│ ├── articles/ # マークダウン記事ファイル
│ ├── images/ # 記事で使用する画像ファイル
│ └── wp/ # WordPress API 関連の実装
│ ├── client.go
│ ├── category.go
│ ├── markdown.go
│ ├── media.go
│ ├── metadata.go
│ ├── tag.go
│ └── types.go
├── go.mod
├── go.sum
└── .env # 環境変数設定ファイル
```

## セットアップ

1. リポジトリをクローン
2. 必要なパッケージをインストール

```bash
go mod tidy
```

3. 環境変数を設定

```env:README.md
WP_URL=https://your-wordpress-site.com
USER_NAME=your-username
USER_PASSWORD=your-password
```

## 使い方

### 新規記事の投稿

```bash
go run cmd/cli create article-name
```

### 既存記事の更新

```bash
go run cmd/cli update article-name
```

※ `article-name`は、たとえば、`internal/articles/1.md`のような記事の場合は`1`となります。
```bash
go run cmd/cli create 1
```

※ internal/articles/ ディレクトリに新しいディレクトリを切った場合は、以下のように新しいディレクトリを指定してください。
```bash
go run cmd/cli create dir-name/article-name

# 例
go run cmd/cli create posts_001-050/1
```

## 記事ファイルの形式

記事は`internal/articles/`ディレクトリに`.md`ファイルとして保存します。
各記事ファイルは以下の形式で記述します：

```markdown
{
"Title": "記事タイトル",
"Image": "アイキャッチ画像ファイル名",
"Permalink": "記事のスラッグ",
"Category": ["カテゴリー 1", "カテゴリー 2"],
"Tag": ["タグ 1", "タグ 2"]
}

---

記事本文をマークダウン形式で記述...
```

## 画像の管理

記事で使用する画像は`internal/images/`ディレクトリに配置します。
マークダウン内で参照された画像は、投稿時に自動的に WordPress にアップロードされます。

## 注意事項

- 環境変数は必ず`.env`ファイルで管理してください
- 画像ファイルは`internal/images/`ディレクトリに配置してください
- 記事の更新には、記事メタデータに`post_id`が必要です

## ライセンス

このプロジェクトは MIT ライセンスの下で公開されています。
