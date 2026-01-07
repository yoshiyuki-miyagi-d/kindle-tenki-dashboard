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
┌─────────────────┐     ┌──────────────────────┐
│   main.go       │────>│  外部API             │
│  (データ取得    │     │ - 天気API            │
│   & HTML生成)   │     │ - ニュースRSS        │
└────────┬────────┘     │ - はてなブックマーク │
                        └──────────────────────┘
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

#### 1.2 ニュースデータ取得 (`fetchNewsData`, `fetchEconomyNewsData`)
- **API**: NHK ニュースRSS (XML)
- **機能**: 主要ニュース5件、経済ニュース10件を取得
- **フォールバック**: API失敗時はサンプルニュースを使用
- **データ構造**: `NHKNewsRSS` -> `[]NewsItem`
- **重複除外**: `filterDuplicateNews()`で主要と経済の重複を除外

#### 1.3 はてなブックマークデータ取得 (`fetchHatenaBookmarks`, `fetchKnowledgeHatenaBookmarks`)
- **API**: はてなブックマークRSS (RDF/XML)
- **機能**: 総合5件、学び5件の人気エントリーを取得
- **フォールバック**: API失敗時はサンプルデータを使用
- **データ構造**: `HatenaBookmarkRSS` -> `[]HatenaEntry`
- **重複除外**: `filterDuplicateHatenaEntries()`で総合と学びの重複を除外

### 2. データ処理層

#### 2.1 天気データ処理 (`processWeatherData`)
- 今日と明日の予報から48時間分の時間別予報を生成
- 気温グラフ用の高さ計算:
  - SVG座標系(上が小さい値、下が大きい値)に対応
  - 最高気温を上部(y=20)、最低気温を下部(y=75)に配置
  - 気温範囲を55ポイント(75-20)にマッピング
  - 全て同じ気温の場合は中央(y=47)に配置
- 時間帯による気温の推定ロジック
- 3日間の日別予報を生成:
  - 今日/明日/明後日の最高・最低気温
  - 各日の降水確率の最大値を計算
  - 天気アイコンと概況を設定

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
1. GitHub Actions (cron: 3時間ごと)
   └─> main.go 実行

2. データ取得
   ├─> 天気API呼び出し
   │   └─> TsukumijimaWeatherResponse 取得
   │       └─> processWeatherData()
   │           └─> WeatherData 生成
   │
   └─> 【並列実行】4つのAPIを同時に呼び出し (goroutineで並列化)
       ├─> 主要ニュースRSS (NHK)
       │   └─> []NewsItem 生成
       ├─> 経済ニュースRSS (NHK)
       │   └─> []NewsItem 生成
       │   └─> filterDuplicateNews() (重複除外)
       ├─> はてなブックマーク(総合)RSS
       │   └─> []HatenaEntry 生成
       └─> はてなブックマーク(学び)RSS
           └─> []HatenaEntry 生成
           └─> filterDuplicateHatenaEntries() (重複除外)

       ※ 並列化の実装方法:
         - 各API呼び出しを個別のgoroutineで実行
         - channelを使用して結果を受け取る(バッファサイズ1)
         - 結果構造体(newsResult, hatenaResult)でエラーとデータを返す
         - 全てのchannelから結果を受信後、エラーハンドリングを実行

       ※ 並列化により実行時間が最大75%短縮 (直列40-50秒 → 並列10-15秒)
       ※ HTTPクライアントの再利用により、さらに接続オーバーヘッドが削減

3. HTML生成
   └─> テンプレート + WeatherData
       └─> dist/index.html 生成
       └─> dist/styles/kindle.css コピー

4. GitHub Pages デプロイ
   └─> 静的ファイル公開

5. Kindle ブラウザ
   └─> https://username.github.io/repo-name/ アクセス
```

## データ構造

### WeatherData
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
    HourlyForecast         []HourlyForecast // 時間別予報
    News                   []NewsItem       // ニュース(主要)
    EconomyNews            []NewsItem       // ニュース(経済)
    HatenaEntries          []HatenaEntry    // はてなブックマーク(総合)
    KnowledgeHatenaEntries []HatenaEntry    // はてなブックマーク(学び)
    DailyForecasts         []DailyForecast  // 3日間の予報
    IsUsingFallbackData    bool             // フォールバックデータを使用しているか
    HasMinTemp             bool             // 最低気温データが有効かどうか
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

### HatenaEntry
```go
type HatenaEntry struct {
    Title       string // エントリータイトル
    Link        string // URL
    Description string // 概要
    PubDate     string // 公開日時
    Category    string // カテゴリ(学び、テクノロジーなど)
}
```

### DailyForecast
```go
type DailyForecast struct {
    Date        string // 日付ラベル(今日/明日/明後日)
    WeatherIcon string // 天気アイコン(絵文字)
    Description string // 天気概況
    MaxTemp     int    // 最高気温(℃)
    MinTemp     int    // 最低気温(℃)
    RainChance  string // 降水確率(最大値)
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

### 2. HTTPクライアントの最適化
- **グローバルHTTPクライアントの再利用**: Keep-Alive接続を有効化し、複数のAPI呼び出しで接続を再利用
- **接続プールのチューニング**:
  - MaxIdleConns: 100 (アイドル接続数)
  - MaxIdleConnsPerHost: 10 (ホストごとのアイドル接続数)
  - IdleConnTimeout: 90秒 (アイドル接続のタイムアウト)
- **HTTP/2の強制試行**: ForceAttemptHTTP2を有効化してプロトコル最適化
- **タイムアウト設定**: 各種タイムアウトを適切に設定 (TLSハンドシェイク: 5秒、ExpectContinue: 1秒)
- **圧縮の有効化**: DisableCompressionをfalseに設定してデータ転送を最適化

これらの最適化により、API呼び出しのオーバーヘッドが大幅に削減され、実行時間が短縮される。

### 3. E-ink最適化
- 最小限のCSS
- JavaScriptなし
- 画像なし (Unicode絵文字のみ使用)

### 4. バッテリー節約
- サーバー側更新頻度: 3時間ごと
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