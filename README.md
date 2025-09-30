# Kindle天気ダッシュボード

Kindle Paperwhite用に最適化された天気情報ダッシュボードです。

## 特徴

- **E-ink最適化**: Kindle Paperwhiteの画面に最適化されたデザイン
- **完全静的サイト**: 高速で安定したパフォーマンス
- **自動更新**: GitHub Actionsで6時間ごとに天気情報を更新
- **省電力**: Kindleのバッテリーを節約する設計

## セットアップ

### 1. リポジトリをフォーク

このリポジトリをGitHubでフォークしてください。

### 2. OpenWeatherMap APIキーを取得

1. [OpenWeatherMap](https://openweathermap.org/api)でアカウントを作成
2. 無料のAPIキーを取得

### 3. GitHub Secretsを設定

リポジトリの Settings > Secrets and variables > Actions で以下を設定:

#### Secrets
- `OPENWEATHER_API_KEY`: OpenWeatherMapのAPIキー

#### Variables (オプション)
- `CITY`: 都市名 (デフォルト: Tokyo)
- `COUNTRY_CODE`: 国コード (デフォルト: JP)

### 4. GitHub Pagesを有効化

1. リポジトリの Settings > Pages
2. Source: "GitHub Actions" を選択

### 5. 初回ビルドを実行

1. Actions タブで "天気情報更新" ワークフロー
2. "Run workflow" をクリックして手動実行

## ローカル開発

```bash
# サンプルデータでビルド
go run main.go

# ローカルサーバーで確認 (別ターミナルで実行)
python -m http.server 8000 --directory docs
# http://localhost:8000 でアクセス
```

## 環境変数

```bash
# .envファイルを作成
OPENWEATHER_API_KEY=your_api_key_here
CITY=Tokyo
COUNTRY_CODE=JP
```

## Kindleでの使用方法

1. Kindle Paperwhiteで実験的ブラウザを起動
2. 生成されたGitHub PagesのURLにアクセス
3. ブックマークに追加して定期的にアクセス

## 更新頻度

- **自動更新**: 6時間ごと (0, 6, 12, 18時)
- **手動更新**: GitHub Actionsから随時実行可能

## カスタマイズ

### デザインの変更
- `src/styles/kindle.css`: スタイルシートを編集
- `src/templates/index.html`: HTMLテンプレートを編集

### 表示データの変更
- `main.go`: データ処理ロジックを編集

### 更新頻度の変更
- `.github/workflows/update-weather.yml`: cronスケジュールを編集

## ライセンス

MIT License