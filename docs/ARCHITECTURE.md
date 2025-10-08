# アーキテクチャ設計書

## システム概要

Kindle Paperwhiteで表示するための天気情報ダッシュボード。E-inkディスプレイに最適化された静的HTMLを生成する。

## アーキテクチャ図

```
┌─────────────────┐
│  GitHub Actions │
│   (定期実行)    │
└────────┬────────┘
         │
         v
┌─────────────────┐     ┌──────────────────┐
│   main.go       │────>│  外部API         │
│  (データ取得    │     │ - 天気API        │
│   & HTML生成)   │     │ - ニュースRSS    │
└────────┬────────┘     └──────────────────┘
         │
         v
┌─────────────────┐
│  docs/          │
│  - index.html   │────> GitHub Pages
│  - styles/      │
│    - kindle.css │
└─────────────────┘
         │
         v
┌─────────────────┐
│ Kindle Browser  │
│ (E-ink表示)     │
└─────────────────┘
```

## コンポーネント構成

### 1. データ取得層 (main.go)

#### 1.1 天気データ取得 (`fetchWeatherData`)
- **API**: `weather.tsukumijima.net`
- **機能**: 指定された都市コードの天気情報を取得
- **フォールバック**: API失敗時はサンプルデータを使用
- **データ構造**: `TsukumijimaWeatherResponse` -> `WeatherData`

#### 1.2 ニュースデータ取得 (`fetchNewsData`)
- **API**: NHK ニュースRSS (XML)
- **機能**: 最新ニュース5件を取得
- **フォールバック**: API失敗時はサンプルニュースを使用
- **データ構造**: `NHKNewsRSS` -> `[]NewsItem`

### 2. データ処理層

#### 2.1 天気データ処理 (`processWeatherData`)
- 今日と明日の予報から48時間分の時間別予報を生成
- 気温グラフ用の高さ計算 (20%〜100%にマッピング)
- 時間帯による気温の推定ロジック

#### 2.2 温度パース (`parseTemperature`)
- 文字列の気温データを整数に変換
- null値や空文字列のハンドリング

#### 2.3 天気アイコン変換 (`getWeatherIcon`)
- 天気の説明文からUnicode絵文字を返す
- 対応パターン: 晴れ(☀️)、曇り(☁️)、雨(☔)、雪(⛄)、雷(⚡)、霧(🌫️)など
- パターンマッチングに`containsAny`関数を使用

### 3. HTML生成層 (`generateHTML`)

#### 3.1 テンプレートエンジン
- Go標準の `html/template` を使用
- カスタム関数: `mul`, `sub` (算術演算)

#### 3.2 出力構造
```
docs/
├── index.html (生成されたHTML)
└── styles/
    └── kindle.css (コピーされたCSS)
```

### 4. プレゼンテーション層

#### 4.1 HTMLテンプレート (`src/templates/index.html`)
- 現在の天気情報表示
- 48時間予報グラフ
- ニュースフィード

#### 4.2 スタイルシート (`src/styles/kindle.css`)
- E-ink最適化: モノクロ、高コントラスト
- 游ゴシック体を使用
- レスポンシブデザイン

## データフロー

```
1. GitHub Actions (cron: 6時間ごと)
   └─> main.go 実行

2. データ取得
   ├─> 天気API呼び出し
   │   └─> TsukumijimaWeatherResponse 取得
   │       └─> processWeatherData()
   │           └─> WeatherData 生成
   │
   └─> ニュースRSS呼び出し
       └─> NHKNewsRSS 取得
           └─> []NewsItem 生成

3. HTML生成
   └─> テンプレート + WeatherData
       └─> docs/index.html 生成
       └─> docs/styles/kindle.css コピー

4. GitHub Pages デプロイ
   └─> 静的ファイル公開

5. Kindle ブラウザ
   └─> https://username.github.io/repo-name/ アクセス
```

## データ構造

### WeatherData
```go
type WeatherData struct {
    Location        string           // 都市名
    Temperature     int              // 現在の気温(℃)
    MinTemp         int              // 最低気温(℃)
    MaxTemp         int              // 最高気温(℃)
    FeelsLike       int              // 体感温度(℃)
    Description     string           // 天気概況
    WeatherIcon     string           // 天気アイコン(絵文字)
    Wind            string           // 風の情報
    ChanceOfRain    []string         // 6時間ごとの降水確率
    UpdateTime      string           // 更新時刻
    HourlyForecast  []HourlyForecast // 時間別予報
    News            []NewsItem       // ニュース
}
```

### HourlyForecast
```go
type HourlyForecast struct {
    Time        string // 時刻 (HH:MM)
    Temp        int    // 気温(℃)
    Desc        string // 天気
    WeatherIcon string // 天気アイコン(絵文字)
    RainChance  string // 降水確率
    ChartHeight int    // グラフ高さ(%) 20-100
}
```

### NewsItem
```go
type NewsItem struct {
    Title       string // ニュースタイトル
    Link        string // URL
    Description string // 概要
    PubDate     string // 公開日時
}
```

## 環境変数

| 変数名 | デフォルト値 | 説明 |
|--------|-------------|------|
| `CITY_CODE` | `130010` | 天気APIの都市コード (130010=東京) |

## エラーハンドリング戦略

### 1. グレースフルデグラデーション
- API失敗時はサンプルデータを使用
- ユーザーには常に表示可能なコンテンツを提供

### 2. ログ出力
- エラー発生箇所と原因を記録
- フォールバックの使用を明示

### 3. ゼロダウンタイム
- GitHub Pagesは前回生成したHTMLを保持
- ビルド失敗時も既存のコンテンツが利用可能

## パフォーマンス最適化

### 1. 静的サイト生成
- サーバーサイド処理なし
- CDN配信による高速ロード

### 2. E-ink最適化
- 最小限のCSS
- JavaScriptなし
- 画像なし (Unicode絵文字のみ使用)

### 3. バッテリー節約
- サーバー側更新頻度: 6時間ごと
- ページ自動リロード: 30分ごと (meta refresh)

## セキュリティ考慮事項

### 1. API認証
- 天気API: 認証不要の無料API使用
- ニュースRSS: 公開RSS使用

### 2. GitHub Secrets
- 将来的にAPIキーが必要になった場合に対応可能

### 3. XSS対策
- Go標準の `html/template` による自動エスケープ

## 拡張性

### 追加可能な機能
1. 複数都市対応
2. 天気アラート機能
3. 週間予報の追加
4. カスタマイズ可能なニュースソース
5. 降水確率の表示

### 制約事項
- Kindleブラウザの制限: JavaScript実行が不安定
- E-inkの制限: カラー表示不可、リフレッシュレート低い
- GitHub Actionsの制限: 実行時間、頻度