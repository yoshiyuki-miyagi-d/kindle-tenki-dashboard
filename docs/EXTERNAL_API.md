# API仕様書

このドキュメントは、プロジェクトで使用している外部APIの仕様と使用方法をまとめたものです。

## 使用しているAPI

1. **天気予報API** - weather.tsukumijima.net
2. **ニュースRSS** - NHKニュース
3. **はてなブックマークRSS** - はてなブックマーク

---

## 1. 天気予報API

### 基本情報

| 項目 | 内容 |
|------|------|
| **提供元** | weather.tsukumijima.net |
| **認証** | 不要 |
| **料金** | 無料 |
| **レート制限** | 明記なし (過度な使用は控える) |
| **データ形式** | JSON |
| **文字エンコーディング** | UTF-8 |

### エンドポイント

```
GET https://weather.tsukumijima.net/api/forecast/city/{cityCode}
```

#### パラメータ

| パラメータ | 型 | 必須 | 説明 | 例 |
|-----------|-----|------|------|-----|
| `cityCode` | string | ✓ | 都市コード | `130010` (東京) |

#### 主要な都市コード

| 都市 | コード |
|------|--------|
| 札幌 | `016010` |
| 東京 | `130010` |
| 横浜 | `140010` |
| 名古屋 | `230010` |
| 大阪 | `270000` |
| 京都 | `260010` |
| 福岡 | `400010` |
| 那覇 | `471010` |

全都市コードは[こちら](https://weather.tsukumijima.net/primary_area.xml)を参照。

### リクエスト例

```bash
curl https://weather.tsukumijima.net/api/forecast/city/130010
```

### レスポンス構造

```json
{
  "publicTime": "2025-10-02T11:00:00+09:00",
  "publicTimeFormatted": "2025/10/02 11:00:00",
  "publishingOffice": "気象庁",
  "title": "東京都 東京 の天気",
  "location": {
    "area": "関東",
    "prefecture": "東京都",
    "district": "東京地方",
    "city": "東京"
  },
  "forecasts": [
    {
      "date": "2025-10-02",
      "dateLabel": "今日",
      "telop": "晴れ",
      "detail": {
        "weather": "晴れ",
        "wind": "北の風",
        "wave": "0.5メートル"
      },
      "temperature": {
        "min": {
          "celsius": "18"
        },
        "max": {
          "celsius": "28"
        }
      },
      "chanceOfRain": {
        "T00_06": "10%",
        "T06_12": "0%",
        "T12_18": "0%",
        "T18_24": "10%"
      },
      "image": {
        "title": "晴れ",
        "url": "https://www.jma.go.jp/bosai/forecast/img/100.svg"
      }
    },
    {
      "date": "2025-10-03",
      "dateLabel": "明日",
      "telop": "晴時々曇",
      "detail": {
        "weather": "晴れ時々くもり",
        "wind": "北の風後南の風",
        "wave": "0.5メートル"
      },
      "temperature": {
        "min": {
          "celsius": "19"
        },
        "max": {
          "celsius": "27"
        }
      },
      "chanceOfRain": {
        "T00_06": "10%",
        "T06_12": "10%",
        "T12_18": "20%",
        "T18_24": "20%"
      },
      "image": {
        "title": "晴時々曇",
        "url": "https://www.jma.go.jp/bosai/forecast/img/101.svg"
      }
    },
    {
      "date": "2025-10-04",
      "dateLabel": "明後日",
      "telop": "曇時々晴",
      "detail": {},
      "temperature": {
        "min": {
          "celsius": "20"
        },
        "max": {
          "celsius": "25"
        }
      },
      "chanceOfRain": {},
      "image": {
        "title": "曇時々晴",
        "url": "https://www.jma.go.jp/bosai/forecast/img/102.svg"
      }
    }
  ]
}
```

### レスポンスフィールド説明

#### ルートレベル

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `publicTime` | string | 発表時刻 (ISO 8601形式) |
| `publicTimeFormatted` | string | 発表時刻 (表示用) |
| `publishingOffice` | string | 発表機関 |
| `title` | string | タイトル |
| `location` | object | 地域情報 |
| `forecasts` | array | 予報配列 (3日分) |

#### location オブジェクト

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `area` | string | 地方 (例: "関東") |
| `prefecture` | string | 都道府県 |
| `district` | string | 地域 |
| `city` | string | 市区町村 |

#### forecasts 配列の各要素

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `date` | string | 日付 (YYYY-MM-DD) |
| `dateLabel` | string | 日付ラベル ("今日", "明日", "明後日") |
| `telop` | string | 天気概況 |
| `detail` | object | 詳細情報 |
| `temperature` | object | 気温情報 |
| `chanceOfRain` | object | 降水確率 |
| `image` | object | 天気アイコン情報 |

#### temperature オブジェクト

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `min.celsius` | string | 最低気温 (℃) ※null の場合あり |
| `max.celsius` | string | 最高気温 (℃) ※null の場合あり |

#### chanceOfRain オブジェクト

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `T00_06` | string | 0-6時の降水確率 |
| `T06_12` | string | 6-12時の降水確率 |
| `T12_18` | string | 12-18時の降水確率 |
| `T18_24` | string | 18-24時の降水確率 |

### 注意事項

1. **null値の処理**
   - `temperature.min.celsius` や `temperature.max.celsius` がnullまたは空文字列の場合がある
   - 特に当日の夜間など、既に過ぎた時刻のデータはnullになる
   - 必ずnullチェックを実装する

   **最低気温データがnullになるケース:**
   - 当日の午後〜夜間: 最低気温は朝に記録されるため、既に過ぎた場合はデータがない
   - 今日の天気情報を夕方以降に取得すると、`min.celsius`が空文字列やnullになる

   **対処方法:**
   - プロジェクトでは`HasMinTemp`フラグを使用して最低気温データの有効性を追跡
   - 今日の最低気温がnullの場合、明日の最低気温を使用するフォールバック処理を実装
   - nullの場合は0を返すのではなく、適切なデフォルト値または代替データを使用

2. **データの更新頻度**
   - 1日3回程度更新される (気象庁の発表タイミングに依存)
   - 5:00, 11:00, 17:00頃に更新されることが多い

3. **エラーハンドリング**
   - 不正な都市コードの場合は404エラー
   - サーバーエラーの場合は500エラー
   - タイムアウトに備えてフォールバック処理を実装する

### 実装例 (Go)

```go
// グローバルHTTPクライアント (Keep-Alive接続を再利用)
var httpClient = &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        ForceAttemptHTTP2:   true,
    },
}

func fetchWeatherData() (*WeatherData, error) {
    cityCode := getEnv("CITY_CODE", "130010")
    url := fmt.Sprintf("https://weather.tsukumijima.net/api/forecast/city/%s", cityCode)

    resp, err := httpClient.Get(url)
    if err != nil {
        return nil, fmt.Errorf("API呼び出しエラー: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("APIエラー: %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("レスポンス読み込みエラー: %w", err)
    }

    var weatherResponse TsukumijimaWeatherResponse
    if err := json.Unmarshal(body, &weatherResponse); err != nil {
        return nil, fmt.Errorf("JSONパースエラー: %w", err)
    }

    return processWeatherData(weatherResponse), nil
}

// 気温パース処理 (null値のハンドリング)
func parseTemperature(tempStr string) (int, error) {
    if tempStr == "" || tempStr == "null" {
        return 0, fmt.Errorf("empty temperature")
    }
    temp, err := strconv.Atoi(tempStr)
    if err != nil {
        return 0, err
    }
    return temp, nil
}
```

---

## 2. ニュースRSS (NHK)

### 基本情報

| 項目 | 内容 |
|------|------|
| **提供元** | NHK (日本放送協会) |
| **認証** | 不要 |
| **料金** | 無料 |
| **レート制限** | 明記なし (過度な使用は控える) |
| **データ形式** | RSS 2.0 (XML) |
| **文字エンコーディング** | UTF-8 |

### エンドポイント

```
GET https://www3.nhk.or.jp/rss/news/cat0.xml
```

#### カテゴリ別RSS

| カテゴリ | URL |
|----------|-----|
| 主要ニュース | `https://www3.nhk.or.jp/rss/news/cat0.xml` |
| 社会 | `https://www3.nhk.or.jp/rss/news/cat1.xml` |
| 気象・災害 | `https://www3.nhk.or.jp/rss/news/cat2.xml` |
| 科学・文化 | `https://www3.nhk.or.jp/rss/news/cat3.xml` |
| 政治 | `https://www3.nhk.or.jp/rss/news/cat4.xml` |
| 経済 | `https://www3.nhk.or.jp/rss/news/cat5.xml` |
| 国際 | `https://www3.nhk.or.jp/rss/news/cat6.xml` |
| スポーツ | `https://www3.nhk.or.jp/rss/news/cat7.xml` |

### リクエスト例

```bash
curl https://www3.nhk.or.jp/rss/news/cat0.xml
```

### レスポンス構造

```xml
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>NHKニュース</title>
    <link>https://www.nhk.or.jp/news/</link>
    <description>NHKニュース</description>
    <language>ja</language>
    <pubDate>Wed, 02 Oct 2025 12:00:00 +0900</pubDate>

    <item>
      <title>新浪氏の処遇 経済同友会が協議 審査会は"辞任勧告が相当"</title>
      <link>http://www3.nhk.or.jp/news/html/20250930/k10014936121000.html</link>
      <description>経済同友会は、サプリメントをめぐる警察の捜査を受けて活動を自粛している、新浪剛史代表幹事の処遇について30日、理事会を開いて協議しています。</description>
      <pubDate>Mon, 30 Sep 2025 12:19:00 +0900</pubDate>
    </item>

    <item>
      <title>10月 値上げの食品 半年ぶり3000品目超 7割が「酒類・飲料」</title>
      <link>http://www3.nhk.or.jp/news/html/20250930/k10014935951000.html</link>
      <description>10月に値上げされる食品は3000品目を超え、ことし4月以来、半年ぶりの高い水準になることが民間の調査でわかりました。</description>
      <pubDate>Mon, 30 Sep 2025 11:26:00 +0900</pubDate>
    </item>

    <!-- 以下、複数のitem要素が続く -->
  </channel>
</rss>
```

### レスポンスフィールド説明

#### channel レベル

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `title` | string | チャンネルタイトル |
| `link` | string | チャンネルURL |
| `description` | string | チャンネル説明 |
| `language` | string | 言語コード |
| `pubDate` | string | 最終更新日時 |

#### item 要素

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `title` | string | ニュースタイトル |
| `link` | string | ニュース記事URL |
| `description` | string | ニュース概要 |
| `pubDate` | string | 公開日時 (RFC 822形式) |

### 日付フォーマット

**入力形式 (RFC 822):**
```
Mon, 30 Sep 2025 12:19:00 +0900
```

**変換後 (表示用):**
```
09/30 12:19
```

### 注意事項

1. **更新頻度**
   - リアルタイムで更新される
   - 新しいニュースが発表されるたびに追加

2. **アイテム数**
   - 通常20-30件程度のニュースが含まれる
   - 上位5件程度を表示することを推奨

3. **XMLパース**
   - Go標準の `encoding/xml` を使用
   - 文字エンコーディングはUTF-8

4. **エラーハンドリング**
   - ネットワークエラーに備える
   - XMLパースエラーに備える

### 実装例 (Go)

```go
type NHKNewsRSS struct {
    XMLName xml.Name `xml:"rss"`
    Channel struct {
        Title       string    `xml:"title"`
        Description string    `xml:"description"`
        Link        string    `xml:"link"`
        Items       []RSSItem `xml:"item"`
    } `xml:"channel"`
}

type RSSItem struct {
    Title       string `xml:"title"`
    Link        string `xml:"link"`
    Description string `xml:"description"`
    PubDate     string `xml:"pubDate"`
}

func fetchNewsData() ([]NewsItem, error) {
    url := "https://www3.nhk.or.jp/rss/news/cat0.xml"

    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("RSS取得エラー: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("レスポンス読み込みエラー: %w", err)
    }

    var rss NHKNewsRSS
    if err := xml.Unmarshal(body, &rss); err != nil {
        return nil, fmt.Errorf("XMLパースエラー: %w", err)
    }

    var news []NewsItem
    maxItems := 5
    for i := 0; i < maxItems && i < len(rss.Channel.Items); i++ {
        item := rss.Channel.Items[i]

        // 日付をフォーマット
        pubTime, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", item.PubDate)
        formattedDate := pubTime.Format("01/02 15:04")

        news = append(news, NewsItem{
            Title:       item.Title,
            Link:        item.Link,
            Description: item.Description,
            PubDate:     formattedDate,
        })
    }

    return news, nil
}
```

---

## エラーハンドリング戦略

### 共通のエラー処理

1. **HTTPクライアントの最適化**
   ```go
   // グローバルHTTPクライアントで接続を再利用
   var httpClient = &http.Client{
       Timeout: 10 * time.Second,
       Transport: &http.Transport{
           MaxIdleConns:          100,
           MaxIdleConnsPerHost:   10,
           IdleConnTimeout:       90 * time.Second,
           TLSHandshakeTimeout:   5 * time.Second,
           ExpectContinueTimeout: 1 * time.Second,
           DisableKeepAlives:     false,
           ForceAttemptHTTP2:     true,
       },
   }
   ```

2. **並列API呼び出し**
   - goroutineとchannelを使用して複数のAPIを同時に呼び出し
   - 各API呼び出しは独立してエラーハンドリングを実行
   - 実行時間を最大75%短縮

3. **フォールバック**
   - API失敗時はサンプルデータを使用
   - `IsUsingFallbackData`フラグでフォールバックの使用を追跡
   - ユーザーには必ず表示可能なコンテンツを提供

4. **ログ記録**
   - エラー内容を詳細に記録
   - フォールバックの使用を明示

### 実装例

```go
func fetchWithRetry(url string, maxRetries int) (*http.Response, error) {
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        resp, err := http.Get(url)
        if err == nil && resp.StatusCode == http.StatusOK {
            return resp, nil
        }
        lastErr = err
        time.Sleep(time.Duration(1<<i) * time.Second) // 指数バックオフ
    }
    return nil, fmt.Errorf("最大リトライ回数を超えました: %w", lastErr)
}
```

---

## 環境変数

### 設定可能な環境変数

| 変数名 | デフォルト値 | 説明 |
|--------|-------------|------|
| `CITY_CODE` | `130010` | 天気APIの都市コード |

### 設定方法

#### ローカル開発
`.env` ファイル:
```bash
CITY_CODE=270000  # 大阪
```

#### GitHub Actions
Variables設定:
```yaml
vars.CITY_CODE || '130010'
```

---

## テストとモニタリング

### 手動テスト

#### 天気API
```bash
# 東京の天気を取得
curl -s https://weather.tsukumijima.net/api/forecast/city/130010 | jq

# 大阪の天気を取得
curl -s https://weather.tsukumijima.net/api/forecast/city/270000 | jq
```

#### ニュースRSS
```bash
# 主要ニュースを取得
curl -s https://www3.nhk.or.jp/rss/news/cat0.xml | xmllint --format -

# 気象ニュースを取得
curl -s https://www3.nhk.or.jp/rss/news/cat2.xml | xmllint --format -
```

### モニタリング

1. **APIの可用性チェック**
   - GitHub Actionsでの定期実行
   - 失敗時はIssueを自動作成

2. **レスポンスタイムの監視**
   - 10秒以上かかる場合は警告

3. **データの妥当性チェック**
   - 気温が-50℃〜50℃の範囲外の場合は警告
   - ニュースが0件の場合は警告

---

## 3. はてなブックマークRSS

### 基本情報

| 項目 | 内容 |
|------|------|
| **提供元** | はてなブックマーク |
| **認証** | 不要 |
| **料金** | 無料 |
| **レート制限** | 明記なし (過度な使用は控える) |
| **データ形式** | RDF/RSS 1.0 (XML) |
| **文字エンコーディング** | UTF-8 |

### エンドポイント

#### 総合カテゴリ
```
GET https://b.hatena.ne.jp/hotentry/all.rss
```

#### 学びカテゴリ
```
GET https://b.hatena.ne.jp/hotentry/knowledge.rss
```

#### その他の主要カテゴリ

| カテゴリ | URL |
|----------|-----|
| 総合 | `https://b.hatena.ne.jp/hotentry/all.rss` |
| テクノロジー | `https://b.hatena.ne.jp/hotentry/it.rss` |
| 学び | `https://b.hatena.ne.jp/hotentry/knowledge.rss` |
| 暮らし | `https://b.hatena.ne.jp/hotentry/life.rss` |
| エンタメ | `https://b.hatena.ne.jp/hotentry/entertainment.rss` |
| おもしろ | `https://b.hatena.ne.jp/hotentry/fun.rss` |
| 政治と経済 | `https://b.hatena.ne.jp/hotentry/economics.rss` |

### リクエスト例

```bash
curl https://b.hatena.ne.jp/hotentry/all.rss
```

### レスポンス構造

```xml
<?xml version="1.0" encoding="UTF-8"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns="http://purl.org/rss/1.0/"
         xmlns:dc="http://purl.org/dc/elements/1.1/"
         xmlns:hatena="http://www.hatena.ne.jp/info/xmlns#">

  <channel rdf:about="https://b.hatena.ne.jp/hotentry/all">
    <title>はてなブックマーク - 人気エントリー - 総合</title>
    <link>https://b.hatena.ne.jp/hotentry/all</link>
    <description>最近の人気エントリー</description>
  </channel>

  <item rdf:about="https://example.com/article">
    <title>記事のタイトル</title>
    <link>https://example.com/article</link>
    <description>記事の概要</description>
    <dc:date>2025-10-30T16:24:16Z</dc:date>
    <dc:subject>学び</dc:subject>
    <dc:subject>テクノロジー</dc:subject>
  </item>

  <!-- 以下、複数のitem要素が続く -->
</rdf:RDF>
```

### レスポンスフィールド説明

#### channel レベル

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `title` | string | チャンネルタイトル |
| `link` | string | チャンネルURL |
| `description` | string | チャンネル説明 |

#### item 要素

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `title` | string | エントリータイトル |
| `link` | string | エントリーURL |
| `description` | string | エントリー概要 |
| `dc:date` | string | ブックマーク日時 (ISO 8601形式) |
| `dc:subject` | string | カテゴリタグ (複数可) |

### 日付フォーマット

**入力形式 (ISO 8601):**
```
2025-10-30T16:24:16Z
```

**変換後 (表示用):**
```
10/30 16:24
```

### プロジェクトでの使用方法

#### データ取得

```go
// 総合カテゴリの人気エントリーを取得
func fetchHatenaBookmarks() ([]HatenaEntry, error) {
    url := "https://b.hatena.ne.jp/hotentry/all.rss"
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Get(url)
    // ... エラーハンドリング ...
}

// 学びカテゴリの人気エントリーを取得
func fetchKnowledgeHatenaBookmarks() ([]HatenaEntry, error) {
    url := "https://b.hatena.ne.jp/hotentry/knowledge.rss"
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Get(url)
    // ... エラーハンドリング ...
}
```

#### 重複除外

総合と学びで重複するエントリーは学びカラムから除外:

```go
func filterDuplicateHatenaEntries(knowledgeEntries []HatenaEntry, generalEntries []HatenaEntry) []HatenaEntry {
    generalTitles := make(map[string]bool)
    for _, item := range generalEntries {
        generalTitles[item.Title] = true
    }

    var filtered []HatenaEntry
    for _, item := range knowledgeEntries {
        if !generalTitles[item.Title] {
            filtered = append(filtered, item)
            if len(filtered) >= MaxHatenaItems {
                break
            }
        }
    }
    return filtered
}
```

### 注意事項

1. **RDF形式**
   - はてなブックマークはRDF/RSS 1.0形式を使用
   - 標準的なRSS 2.0とは構造が異なる
   - 名前空間の処理が必要

2. **レート制限**
   - 明示的な制限はないが、過度なリクエストは控える
   - プロジェクトでは3時間ごとに更新

3. **カテゴリタグ**
   - 複数の`dc:subject`要素が存在する場合がある
   - プロジェクトでは最初のカテゴリのみを使用

---

## 参考リンク

- [天気API (Tsukumijima)](https://weather.tsukumijima.net/)
- [NHK ニュースRSS一覧](https://www.nhk.or.jp/toppage/rss/index.html)
- [はてなブックマーク](https://b.hatena.ne.jp/)
- [気象庁](https://www.jma.go.jp/)
- [RFC 822 (日付フォーマット)](https://www.ietf.org/rfc/rfc822.txt)
- [ISO 8601 (日付フォーマット)](https://www.iso.org/iso-8601-date-and-time-format.html)