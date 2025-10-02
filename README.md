# Kindle天気ダッシュボード

Kindle Paperwhite用に最適化された天気情報ダッシュボードです。

## 特徴

- **E-ink最適化**: Kindle Paperwhiteの画面に最適化されたモノクロデザイン
- **完全静的サイト**: GitHub Pagesで高速配信
- **自動更新**: GitHub Actionsで6時間ごとに天気情報を更新
- **48時間予報**: 3時間ごとの気温変化を折れ線グラフで表示
- **ニュースフィード**: NHKニュースの最新5件を表示
- **省電力**: JavaScriptなしで動作、Kindleのバッテリーを節約

## スクリーンショット

現在の天気、48時間の気温予報グラフ、最新ニュースを1画面に表示。
Kindle Paperwhite (758x1024px) に最適化されたレスポンシブデザイン。

## セットアップ

### 1. リポジトリをフォーク

このリポジトリをGitHubでフォークしてください。

### 2. 環境変数の設定 (オプション)

リポジトリの Settings > Secrets and variables > Actions > Variables で以下を設定:

| 変数名 | デフォルト値 | 説明 |
|--------|-------------|------|
| `CITY_CODE` | `130010` | 都市コード (天気APIで使用) |

**主要な都市コード:**
- 札幌: `016010`
- 東京: `130010` (デフォルト)
- 横浜: `140010`
- 名古屋: `230010`
- 大阪: `270000`
- 京都: `260010`
- 福岡: `400010`
- 那覇: `471010`

全都市コードは[こちら](https://weather.tsukumijima.net/primary_area.xml)を参照。

### 3. GitHub Pagesを有効化

1. リポジトリの Settings > Pages
2. Source: "GitHub Actions" を選択

### 4. 初回ビルドを実行

1. Actions タブで "天気情報更新" ワークフロー
2. "Run workflow" をクリックして手動実行
3. 完了後、 `https://<username>.github.io/<repository-name>/` にアクセス

## Kindleでの使用方法

1. Kindle Paperwhiteで実験的ブラウザを起動
2. GitHub PagesのURL (`https://<username>.github.io/<repository-name>/`) にアクセス
3. ブックマークに追加
4. 6時間ごとに自動更新されるため、定期的にページをリフレッシュ

## ローカル開発

### 必要な環境
- Go 1.21以上
- Python 3.x (ローカルサーバー用)

### 開発手順

```bash
# 1. リポジトリをクローン
git clone https://github.com/your-username/kindle-tenki-dashbaord.git
cd kindle-tenki-dashbaord

# 2. 環境変数を設定 (オプション)
cp .env.example .env
# .env を編集して CITY_CODE を設定

# 3. ビルドと実行
go run main.go

# 4. ローカルサーバーで確認
python -m http.server 8000 --directory docs

# 5. ブラウザで http://localhost:8000 にアクセス
```

### 開発者向けドキュメント

詳細なドキュメントを用意しています:

- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - システム設計とデータフロー
- **[CONTRIBUTING.md](./CONTRIBUTING.md)** - コーディング規約とコミットガイドライン
- **[EXTERNAL_API.md](./EXTERNAL_API.md)** - 外部API仕様と使用方法
- **[DEVELOPMENT.md](./DEVELOPMENT.md)** - 開発環境のセットアップとデバッグ方法

## 使用しているAPI

### 1. 天気予報API
- **提供元**: [weather.tsukumijima.net](https://weather.tsukumijima.net/)
- **認証**: 不要
- **データ**: 今日・明日・明後日の天気と気温

### 2. ニュースRSS
- **提供元**: [NHKニュース](https://www.nhk.or.jp/toppage/rss/index.html)
- **認証**: 不要
- **データ**: 主要ニュースの最新5件

## 更新頻度

- **自動更新**: 6時間ごと (0, 6, 12, 18時 JST)
- **手動更新**: GitHub Actions の "天気情報更新" ワークフローから実行可能
- **トリガー**: mainブランチへのプッシュ時も実行

## カスタマイズ

### 都市の変更

GitHub Variables で `CITY_CODE` を設定:
```
Settings > Secrets and variables > Actions > Variables
New repository variable:
  Name: CITY_CODE
  Value: 270000  (例: 大阪)
```

### デザインの変更

```bash
# CSSを編集
vi src/styles/kindle.css

# HTMLテンプレートを編集
vi src/templates/index.html

# ビルドして確認
go run main.go
python -m http.server 8000 --directory docs
```

### データ処理ロジックの変更

```bash
# main.go を編集
vi main.go

# ビルドして確認
go run main.go
```

### 更新頻度の変更

```bash
# GitHub Actions ワークフローを編集
vi .github/workflows/update-weather.yml

# cron スケジュールを変更 (例: 3時間ごと)
cron: '0 */3 * * *'
```

## トラブルシューティング

### ビルドが失敗する

```bash
# Go のバージョンを確認
go version  # 1.21以上であることを確認

# 依存関係を整理
go mod tidy

# エラーログを確認
go run main.go 2>&1
```

### APIからデータが取得できない

- ネットワーク接続を確認
- サンプルデータで動作することを確認: `go run main.go`
- エラーログを確認: GitHub Actions の Logs タブ

### Kindleで表示が崩れる

- ブラウザの開発者ツールでモノクロ表示をシミュレート
- カスタムデバイスサイズ: 758x1024 (Kindle Paperwhite)
- Kindleブラウザは機能が制限されているため、シンプルなCSSを使用

詳細は [DEVELOPMENT.md](./DEVELOPMENT.md) の「よくある問題と解決方法」を参照。

## 技術スタック

- **言語**: Go 1.21
- **テンプレートエンジン**: html/template (Go標準)
- **CI/CD**: GitHub Actions
- **ホスティング**: GitHub Pages
- **外部API**:
  - 天気予報: weather.tsukumijima.net
  - ニュース: NHK RSS

## プロジェクト構成

```
kindle-tenki-dashbaord/
├── .github/workflows/    # GitHub Actions設定
├── docs/                 # 生成されるHTML (GitHub Pages公開ディレクトリ)
├── src/
│   ├── templates/       # HTMLテンプレート
│   └── styles/          # CSSソースファイル
├── main.go              # メインアプリケーション
├── ARCHITECTURE.md      # システム設計書
├── CONTRIBUTING.md      # 開発ガイドライン
├── EXTERNAL_API.md      # 外部API仕様書
├── DEVELOPMENT.md       # 開発環境セットアップ
└── README.md            # このファイル
```

## コントリビューション

プルリクエストを歓迎します。大きな変更の場合は、まずIssueで議論してください。

コントリビューション前に [CONTRIBUTING.md](./CONTRIBUTING.md) をお読みください。

## ライセンス

MIT License

## 参考リンク

- [天気API (Tsukumijima)](https://weather.tsukumijima.net/)
- [NHK ニュースRSS](https://www.nhk.or.jp/toppage/rss/index.html)
- [GitHub Actions ドキュメント](https://docs.github.com/ja/actions)
- [Kindle Paperwhite](https://www.amazon.co.jp/kindle-paperwhite)