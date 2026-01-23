# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

Kindle Paperwhite用に最適化された天気情報ダッシュボード。Go言語で静的HTMLを生成し、GitHub Pagesで配信する。

## ビルドとテストコマンド

### 開発
```bash
# アプリケーションをビルドして実行 (天気データ取得 + HTML生成)
go run main.go

# 生成されたHTMLをローカルサーバーで確認
python -m http.server 8000 --directory dist

# http://localhost:8000 にアクセス
```

### テスト
```bash
# すべてのテストを実行
go test -v

# カバレッジ付きで実行
go test -cover

# カバレッジレポートをHTMLで表示
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### デバッグ
```bash
# 天気API のレスポンスを確認
curl -s https://weather.tsukumijima.net/api/forecast/city/130010 | jq

# ニュースRSS のレスポンスを確認
curl -s https://www3.nhk.or.jp/rss/news/cat0.xml | xmllint --format -

# 生成されたHTMLを確認
cat dist/index.html
```

## アーキテクチャ

### データフロー
```
GitHub Actions (3時間ごと)
  └→ main.go 実行 (グローバルHTTPクライアントで接続を再利用)
      ├→ 天気API呼び出し (weather.tsukumijima.net)
      │   └→ processWeatherData() → WeatherData (3日間予報 + 48時間グラフデータ生成)
      ├→ 【並列実行】4つのAPI呼び出し (goroutine + channel)
      │   ├→ ニュースRSS (主要) → []NewsItem
      │   ├→ ニュースRSS (経済) → []NewsItem → filterDuplicateNews() (重複除外)
      │   ├→ はてなブックマークRSS (総合) → []HatenaEntry
      │   └→ はてなブックマークRSS (学び) → []HatenaEntry → filterDuplicateHatenaEntries()
      └→ HTMLテンプレート + データ
          ├→ dist/index.html 生成
          └→ dist/styles/kindle.css コピー
```

### 主要コンポーネント

**main.go** - 単一ファイルで完結する構成
- `httpClient` - グローバルHTTPクライアント (Keep-Alive接続の再利用、HTTP/2対応)
- `fetchWeatherData()` - 天気APIからデータ取得、4つのRSS APIを並列呼び出し、失敗時はサンプルデータ使用
- `processWeatherData()` - 3日間の日別予報と48時間分の時間別予報を生成、気温グラフ用の高さ計算 (SVG座標系対応)
- `fetchNewsData()` - 主要ニュースRSSから最新5件取得
- `fetchEconomyNewsData()` - 経済ニュースRSSから最新10件取得
- `fetchHatenaBookmarks()` - はてなブックマーク(総合)から最新5件取得
- `fetchKnowledgeHatenaBookmarks()` - はてなブックマーク(学び)から最新5件取得
- `filterDuplicateNews()` - ニュースの重複除外 (主要と経済で重複するタイトルを除外)
- `filterDuplicateHatenaEntries()` - はてなブックマークの重複除外 (総合と学びで重複するタイトルを除外)
- `generateHTML()` - html/templateを使用してHTML生成
- `parseTemperature()` - 気温文字列を整数に変換、null/空文字列のハンドリング
- `getWeatherIcon()` - 天気説明からUnicode絵文字を返す
- `containsAny()` - 文字列に指定されたいずれかの部分文字列が含まれるかチェック

**src/templates/index.html** - HTMLテンプレート
- カスタム関数: `mul`, `sub` (算術演算)

**src/styles/kindle.css** - E-ink最適化CSS
- モノクロ、高コントラスト
- 游ゴシック体
- レスポンシブデザイン (758x1024px対応)

### データ構造

```go
type WeatherData struct {
    Location               string           // 都市名
    Temperature            int              // 現在の気温(℃)
    MinTemp                int              // 最低気温(℃)
    MaxTemp                int              // 最高気温(℃)
    FeelsLike              int              // 体感温度(℃)
    Description            string           // 天気概況
    WeatherIcon            string           // 天気アイコン(絵文字)
    Wind                   string           // 風の情報
    ChanceOfRain           []string         // 6時間ごとの降水確率
    UpdateTime             string           // 更新時刻
    HourlyForecast         []HourlyForecast // 時間別予報 (最大20件)
    News                   []NewsItem       // ニュース(主要)
    EconomyNews            []NewsItem       // ニュース(経済)
    HatenaEntries          []HatenaEntry    // はてなブックマーク(総合)
    KnowledgeHatenaEntries []HatenaEntry    // はてなブックマーク(学び)
    DailyForecasts         []DailyForecast  // 3日間の予報
    IsUsingFallbackData    bool             // フォールバックデータを使用しているか
    HasMinTemp             bool             // 最低気温データが有効かどうか (夜間でデータがない場合false)
}

type HourlyForecast struct {
    Time        string // 時刻 (HH:MM)
    Temp        int    // 気温(℃)
    Desc        string // 天気
    WeatherIcon string // 天気アイコン(絵文字)
    RainChance  string // 降水確率
    ChartHeight int    // グラフ高さ(%) SVG座標系: 最高気温=20、最低気温=75
}

type DailyForecast struct {
    Date        string // 日付ラベル(今日/明日/明後日)
    WeatherIcon string // 天気アイコン(絵文字)
    Description string // 天気概況
    MaxTemp     int    // 最高気温(℃)
    MinTemp     int    // 最低気温(℃)
    RainChance  string // 降水確率(最大値)
}

type NewsItem struct {
    Title       string // タイトル
    Link        string // URL
    Description string // 概要
    PubDate     string // 公開日時 (MM/DD HH:MM形式)
}

type HatenaEntry struct {
    Title       string // タイトル
    Link        string // URL
    Description string // 概要
    PubDate     string // 公開日時 (MM/DD HH:MM形式)
    Category    string // カテゴリ
}
```

## コーディング規約

### 命名規則 - 禁止事項
以下の意味のない単語は使用禁止:
- `common`, `util`, `helper`, `manager`

具体的な処理内容を示す名前を使用すること。

**悪い例**: `utilProcessData()`, `commonHelper()`
**良い例**: `parseWeatherResponse()`, `formatTemperature()`

### エラーハンドリング
- 必ずエラーを処理する
- フォールバック処理を実装する (API失敗時はサンプルデータ使用)
- ログメッセージは具体的に記述

```go
data, err := fetchWeatherData()
if err != nil {
    log.Printf("⚠️  天気データの取得に失敗しました: %v", err)
    log.Println("   サンプルデータを使用します")
    return getSampleData()
}
```

### Git操作
- **Claude Codeからgit commitを実行しないこと**
- ユーザーが自分でコミットを作成する運用を基本とする
- コード変更はClaude Codeが行い、コミットはユーザーが手動で実施
- **コミットメッセージの提案はClaude Codeが行うこと**
  - 変更内容を説明する適切なコミットメッセージを提示する
  - 下記のコミットメッセージ規約に従った形式で提案する

### コミットメッセージ (ユーザー向け)
- 日本語で記述
- 体言止め禁止、完全な文章で終わらせる
- 動詞の形: 「機能を追加した」「バグを修正した」「ドキュメントを更新した」

**悪い例**: 「機能を追加」「バグの修正」
**良い例**: 「機能を追加した」「バグを修正した」

## 環境変数

| 変数名 | デフォルト値 | 説明 |
|--------|-------------|------|
| `CITY_CODE` | `130010` | 天気APIの都市コード (130010=東京) |

ローカル開発では `.env` ファイルで設定可能。

## ディレクトリ構成

```
kindle-tenki-dashboard/
├── .github/workflows/      # GitHub Actions設定
│   └── update-weather.yml # 3時間ごとに自動実行
├── dist/                   # ビルド成果物 (GitHub Pages公開ディレクトリ)
│   ├── index.html         # 生成されたHTML
│   └── styles/
│       └── kindle.css     # コピーされたCSS
├── docs/                   # プロジェクトドキュメント
│   ├── ARCHITECTURE.md    # システム設計とデータフロー
│   ├── CONTRIBUTING.md    # コーディング規約詳細
│   ├── EXTERNAL_API.md    # 外部API仕様
│   ├── DEVELOPMENT.md     # 開発環境セットアップ詳細
│   ├── CODE_REVIEW.md     # コードレビュー
│   ├── MODERNIZATION.md   # Goモダン化提案
│   └── FEATURE_IDEAS.md   # 機能アイデア
├── src/
│   ├── templates/         # HTMLテンプレート
│   │   └── index.html
│   └── styles/            # CSSソースファイル
│       └── kindle.css
├── main.go                # メインアプリケーション (単一ファイル)
├── main_test.go           # ユニットテスト
├── go.mod                 # Go依存関係管理 (外部依存なし)
└── README.md              # プロジェクト概要
```

## よくあるタスク

### 都市を変更する
```bash
# ローカル開発
CITY_CODE=270000 go run main.go  # 大阪

# GitHub Actions
# Settings > Secrets and variables > Actions > Variables
# CITY_CODE を設定
```

主要な都市コード: 東京=130010、大阪=270000、福岡=400010

### CSSデザインを変更する
```bash
# 1. CSSを編集
vi src/styles/kindle.css

# 2. ビルド (CSSがdist/にコピーされる)
go run main.go

# 3. 確認
python -m http.server 8000 --directory dist
```

### HTMLテンプレートを変更する
```bash
# 1. テンプレートを編集
vi src/templates/index.html

# 2. ビルド
go run main.go

# 3. 確認
python -m http.server 8000 --directory dist
```

### データ処理ロジックを変更する
```bash
# 1. main.go を編集
vi main.go

# 2. テストを実行
go test -v

# 3. ビルドして確認
go run main.go
python -m http.server 8000 --directory dist
```

### 更新頻度を変更する
```bash
# .github/workflows/update-weather.yml を編集
# cron スケジュールを変更
# 例: 6時間ごと → '0 */6 * * *'
# 例: 1時間ごと → '0 * * * *'
```

## 技術的な制約事項

### E-inkディスプレイの制限
- カラー表示不可 → モノクロデザイン必須
- リフレッシュレート低い → JavaScriptなし、最小限のCSS
- 画面サイズ: 758x1024px (Kindle Paperwhite)

### Kindleブラウザの制限
- JavaScript実行が不安定 → 使用しない
- CSS機能が制限されている → Flexbox/Gridは避ける
- フォントが制限されている → システムフォントを使用

### GitHub Actionsの制限
- cron実行は最短1時間間隔 (現在は3時間ごと)
- タイムゾーンはUTC (TZ環境変数でJSTに変更済み)

## デバッグのヒント

### API失敗時の動作確認
アプリケーションはグレースフルデグラデーションを実装しており、API失敗時は自動的にサンプルデータを使用する。ネットワークを切断してテスト可能:
```bash
# Wi-Fiをオフにして実行
go run main.go
# → サンプルデータで正常にHTMLが生成される
```

### Kindleでの表示確認
ブラウザの開発者ツールで:
- デバイスツールバー有効化 (Cmd+Shift+M)
- カスタムサイズ: 758x1024
- モノクロシミュレーション: Rendering > Emulate CSS media feature > prefers-color-scheme: monochrome

## 参考資料

詳細な情報は以下のドキュメントを参照:
- [ARCHITECTURE.md](./docs/ARCHITECTURE.md) - システム設計書
- [CONTRIBUTING.md](./docs/CONTRIBUTING.md) - 開発ガイドライン
- [EXTERNAL_API.md](./docs/EXTERNAL_API.md) - 外部API仕様
- [DEVELOPMENT.md](./docs/DEVELOPMENT.md) - 開発環境セットアップ
- [MODERNIZATION.md](./docs/MODERNIZATION.md) - Goモダン化提案とアップデート計画
