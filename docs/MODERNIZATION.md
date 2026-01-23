# モダン化提案 (Modernization Proposals)

**作成日**: 2026-01-23
**参照元**: Context7 Go Documentation (golang/go, go.dev)
**対象バージョン**: Go 1.21 → Go 1.23+

このドキュメントは、Context7のGoドキュメントレビューに基づき、最新のGoベストプラクティスとAPIを適用するための提案をまとめたものです。

---

## 目次

1. [概要](#概要)
2. [優先度別改善提案](#優先度別改善提案)
3. [詳細な実装ガイド](#詳細な実装ガイド)
4. [参考資料](#参考資料)

---

## 概要

### レビューの目的

最新のGo標準ライブラリのベストプラクティスを適用し、以下の点を改善する:
- パフォーマンスの向上
- リソース管理の最適化
- エラーハンドリングの改善
- セキュリティの強化
- 保守性の向上

### レビュー対象

- `net/http` パッケージ (HTTPクライアント)
- `encoding/json` パッケージ (JSONデコーダ)
- `encoding/xml` パッケージ (XMLデコーダ)
- `html/template` パッケージ (テンプレートエンジン)
- `context` パッケージ (コンテキスト管理)

### 主要な発見事項

現在の実装は基本的に良好だが、以下の点でGo 1.13以降のベストプラクティスが適用されていない:
1. HTTPリクエストでコンテキストを使用していない
2. ストリーミングデコード (Decoder) を活用していない
3. エラーラッピング (`%w`) が一部で未使用
4. Go言語のバージョンが古い (1.21)

---

## 優先度別改善提案

### 🔴 高優先度 (セキュリティ・信頼性に直結)

#### 1. HTTPリクエストでのコンテキスト利用

**問題点**: `httpClient.Get()` を直接呼び出しており、個別リクエストごとのタイムアウト制御やキャンセルができない。

**影響**:
- リクエストのキャンセルができない
- タイムアウトがグローバルHTTPクライアントの設定 (10秒) に固定される
- Go 1.13以降のベストプラクティスに準拠していない

**対象箇所**:
- `fetchWeatherData()` (main.go:216)
- `fetchNewsData()` (main.go:608)
- `fetchEconomyNewsData()` (main.go:659)
- `fetchHatenaBookmarks()` (main.go:711)
- `fetchKnowledgeHatenaBookmarks()` (main.go:771)

**推奨実装**: 詳細は [実装ガイド: コンテキスト付きHTTPリクエスト](#1-コンテキスト付きhttpリクエスト) を参照

---

### 🟡 中優先度 (パフォーマンス・効率性に影響)

#### 2. JSON/XMLストリーミングデコードの採用

**問題点**: `io.ReadAll()` で全体を読み込んでから `Unmarshal()` を使用している。

**影響**:
- 大きなレスポンスでメモリ使用量が増加
- パフォーマンス面で最適ではない
- 2回のステップ (読み込み + パース) が必要

**対象箇所**:
- JSON: main.go:230-242
- XML (NHKニュース): main.go:619-626
- XML (経済ニュース): main.go:670-677
- XML (はてなブックマーク): main.go:722-729
- XML (はてなブックマーク学び): main.go:782-789

**推奨実装**: 詳細は [実装ガイド: ストリーミングデコード](#2-jsonxmlストリーミングデコード) を参照

---

#### 3. エラーハンドリングの改善

**問題点**: エラーメッセージが一般的で、デバッグ情報が不足している箇所がある。

**影響**:
- トラブルシューティング時に原因特定が困難
- エラーチェーンが途切れる
- Go 1.13以降の `%w` によるエラーラッピングが活用されていない

**対象箇所**:
- `parseTemperature()` (main.go:576-585)
- その他エラー処理全般

**推奨実装**: 詳細は [実装ガイド: エラーハンドリング](#3-エラーハンドリングの改善) を参照

---

### 🟢 低優先度 (将来的な改善)

#### 4. HTTP Transportの設定追加

**提案**: `ResponseHeaderTimeout` の追加

**影響**: サーバーがヘッダーを返すまでの待機時間を制御できる

**対象箇所**: main.go:28-40

**推奨実装**: 詳細は [実装ガイド: HTTP Transport設定](#4-http-transport設定の追加) を参照

---

#### 5. テンプレート実行エラーの詳細化

**提案**: `template.ExecError` 型による詳細なエラー情報の取得

**影響**: テンプレート実行エラー時のデバッグが容易になる

**対象箇所**: main.go:974-976

**推奨実装**: 詳細は [実装ガイド: テンプレートエラー](#5-テンプレート実行エラーの詳細化) を参照

---

#### 6. Go言語バージョンのアップデート

**提案**: Go 1.23 以降へのアップグレード

**メリット**:
- スライスとマップのジェネリック操作が改善
- イテレータサポート
- テストAPIの改善 (`b.Loop()` など)
- セキュリティとパフォーマンスの向上

**対象箇所**: go.mod:3

**推奨実装**: 詳細は [実装ガイド: Goバージョンアップデート](#6-go言語バージョンのアップデート) を参照

---

## 詳細な実装ガイド

### 1. コンテキスト付きHTTPリクエスト

#### 背景

Go 1.7で`context`パッケージが標準ライブラリに追加され、Go 1.13で`http.NewRequestWithContext()`が導入されました。これにより、個別のリクエストごとにタイムアウトやキャンセルを制御できるようになりました。

#### Context7のベストプラクティス

```go
// 推奨される実装パターン
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
if err != nil {
    return nil, fmt.Errorf("リクエストの作成に失敗: %w", err)
}

resp, err := httpClient.Do(req)
if err != nil {
    return nil, fmt.Errorf("HTTPリクエストに失敗: %w", err)
}
defer resp.Body.Close()
```

#### 実装例: fetchWeatherData()

**現在の実装** (main.go:216-222):
```go
resp, err := httpClient.Get(weatherURL)
if err != nil {
    log.Printf("⚠️  天気APIの取得に失敗しました: %v", err)
    log.Println("   サンプルデータを使用します")
    return getSampleData()
}
defer resp.Body.Close()
```

**推奨される実装**:
```go
// コンテキストを作成 (5秒のタイムアウト)
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// コンテキスト付きリクエストを作成
req, err := http.NewRequestWithContext(ctx, "GET", weatherURL, nil)
if err != nil {
    log.Printf("⚠️  天気APIリクエストの作成に失敗しました: %v", err)
    log.Println("   サンプルデータを使用します")
    return getSampleData()
}

// リクエストを実行
resp, err := httpClient.Do(req)
if err != nil {
    log.Printf("⚠️  天気APIの取得に失敗しました: %v", err)
    log.Println("   サンプルデータを使用します")
    return getSampleData()
}
defer resp.Body.Close()
```

#### 実装例: fetchNewsData()

**現在の実装** (main.go:608-612):
```go
resp, err := httpClient.Get(url)
if err != nil {
    return nil, fmt.Errorf("ニュースRSSの取得に失敗しました: %w", err)
}
defer resp.Body.Close()
```

**推奨される実装**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
if err != nil {
    return nil, fmt.Errorf("ニュースRSSリクエストの作成に失敗しました: %w", err)
}

resp, err := httpClient.Do(req)
if err != nil {
    return nil, fmt.Errorf("ニュースRSSの取得に失敗しました: %w", err)
}
defer resp.Body.Close()
```

#### メリット

1. **個別のタイムアウト制御**: リクエストごとに異なるタイムアウトを設定可能
2. **キャンセル可能**: 必要に応じてリクエストをキャンセルできる
3. **ベストプラクティス準拠**: Go 1.13以降の推奨パターン
4. **デバッグ情報**: コンテキストでリクエストのトレースが可能

#### 注意点

- `defer cancel()` を必ず呼び出す (リソースリーク防止)
- タイムアウト時間は API の特性に合わせて調整する
- エラーメッセージは `%w` を使ってラップする

---

### 2. JSON/XMLストリーミングデコード

#### 背景

`io.ReadAll()` + `Unmarshal()` のパターンは、レスポンス全体をメモリに読み込む必要があります。`Decoder`を使用すると、ストリーミング処理により効率的にパースできます。

#### Context7のベストプラクティス

```go
// JSON の場合
var data MyStruct
if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
    return nil, fmt.Errorf("JSONのパースに失敗: %w", err)
}

// XML の場合
var rss RSSFeed
if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
    return nil, fmt.Errorf("XMLのパースに失敗: %w", err)
}
```

#### 実装例: JSON (天気API)

**現在の実装** (main.go:230-242):
```go
body, err := io.ReadAll(resp.Body)
if err != nil {
    log.Printf("⚠️  天気データの読み込みに失敗しました: %v", err)
    log.Println("   サンプルデータを使用します")
    return getSampleData()
}

var weatherResponse TsukumijimaWeatherResponse
if err := json.Unmarshal(body, &weatherResponse); err != nil {
    log.Printf("⚠️  天気データのパースに失敗しました: %v", err)
    log.Println("   サンプルデータを使用します")
    return getSampleData()
}
```

**推奨される実装**:
```go
var weatherResponse TsukumijimaWeatherResponse
if err := json.NewDecoder(resp.Body).Decode(&weatherResponse); err != nil {
    log.Printf("⚠️  天気データのパースに失敗しました: %v", err)
    log.Println("   サンプルデータを使用します")
    return getSampleData()
}
```

#### 実装例: XML (NHKニュースRSS)

**現在の実装** (main.go:619-626):
```go
body, err := io.ReadAll(resp.Body)
if err != nil {
    return nil, fmt.Errorf("ニュースRSSの読み込みに失敗しました: %w", err)
}

var rss NHKNewsRSS
if err := xml.Unmarshal(body, &rss); err != nil {
    return nil, fmt.Errorf("ニュースRSSのパースに失敗しました: %w", err)
}
```

**推奨される実装**:
```go
var rss NHKNewsRSS
if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
    return nil, fmt.Errorf("ニュースRSSのパースに失敗しました: %w", err)
}
```

#### メリット

1. **メモリ効率**: 全体を読み込まずにストリーミング処理
2. **パフォーマンス**: `io.ReadAll()` の呼び出しが不要
3. **コード簡潔化**: 2ステップが1ステップになる
4. **大きなレスポンス対応**: メモリ使用量を抑えられる

#### 適用箇所

以下の5箇所に適用可能:

| 関数 | 行番号 | 形式 | 優先度 |
|------|--------|------|--------|
| `fetchWeatherData()` | 230-242 | JSON | 高 |
| `fetchNewsData()` | 619-626 | XML | 高 |
| `fetchEconomyNewsData()` | 670-677 | XML | 高 |
| `fetchHatenaBookmarks()` | 722-729 | XML | 高 |
| `fetchKnowledgeHatenaBookmarks()` | 782-789 | XML | 高 |

#### 注意点

- `Decoder`は`io.Reader`から直接読み込むため、`resp.Body`を直接渡す
- エラーハンドリングは `%w` を使ってラップする
- デコードエラーは元のエラーを保持する

---

### 3. エラーハンドリングの改善

#### 背景

Go 1.13で`fmt.Errorf`に`%w`動詞が追加され、エラーをラップして元のエラーを保持できるようになりました。これにより、エラーチェーンを辿ってデバッグしやすくなります。

#### Context7のベストプラクティス

```go
// エラーラッピングの推奨パターン
if err != nil {
    return fmt.Errorf("操作の説明: %w", err)
}

// 詳細なコンテキスト情報を追加
if err != nil {
    return fmt.Errorf("無効な値 %q の処理に失敗: %w", value, err)
}
```

#### 実装例: parseTemperature()

**現在の実装** (main.go:576-585):
```go
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

**推奨される実装**:
```go
func parseTemperature(tempStr string) (int, error) {
    if tempStr == "" {
        return 0, fmt.Errorf("temperature is empty")
    }
    if tempStr == "null" {
        return 0, fmt.Errorf("temperature is null")
    }
    temp, err := strconv.Atoi(tempStr)
    if err != nil {
        // エラーメッセージに入力値を含める
        return 0, fmt.Errorf("invalid temperature format %q: %w", tempStr, err)
    }
    return temp, nil
}
```

#### メリット

1. **デバッグ情報の充実**: 失敗した値を含めることで、原因特定が容易
2. **エラーチェーンの保持**: `errors.Is()` や `errors.As()` でエラーの種類を判定可能
3. **保守性の向上**: エラーメッセージが明確になる

#### 適用箇所

以下の関数でエラーハンドリングを改善可能:

| 関数 | 行番号 | 改善内容 |
|------|--------|----------|
| `parseTemperature()` | 576-585 | 入力値をエラーメッセージに含める |
| `fetchWeatherData()` | 218, 232, 238 | エラーラッピングの一貫性向上 |
| `fetchNewsData()` | 610, 620, 625 | エラーラッピングの一貫性向上 |
| `generateHTML()` | 948, 958, 971, 976, 980 | エラーラッピングの一貫性向上 |

#### 注意点

- エラーメッセージは小文字で始める (Goの慣例)
- 元のエラーを保持するため、必ず `%w` を使用する
- エラーメッセージには失敗した値や操作の詳細を含める

---

### 4. HTTP Transport設定の追加

#### 背景

`http.Transport`には、リクエストの各段階でタイムアウトを制御するための設定があります。`ResponseHeaderTimeout`は、サーバーがレスポンスヘッダーを返すまでの待機時間を制御します。

#### Context7のベストプラクティス

```go
var httpClient = &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:          100,
        MaxIdleConnsPerHost:   10,
        IdleConnTimeout:       90 * time.Second,
        TLSHandshakeTimeout:   5 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
        ResponseHeaderTimeout: 5 * time.Second,  // 追加
        DisableKeepAlives:     false,
        DisableCompression:    false,
        ForceAttemptHTTP2:     true,
    },
}
```

#### 実装例

**現在の実装** (main.go:28-40):
```go
var httpClient = &http.Client{
    Timeout: HTTPClientTimeout,
    Transport: &http.Transport{
        MaxIdleConns:          100,
        MaxIdleConnsPerHost:   10,
        IdleConnTimeout:       90 * time.Second,
        TLSHandshakeTimeout:   5 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
        DisableKeepAlives:     false,
        DisableCompression:    false,
        ForceAttemptHTTP2:     true,
    },
}
```

**推奨される実装**:
```go
var httpClient = &http.Client{
    Timeout: HTTPClientTimeout,
    Transport: &http.Transport{
        MaxIdleConns:          100,
        MaxIdleConnsPerHost:   10,
        IdleConnTimeout:       90 * time.Second,
        TLSHandshakeTimeout:   5 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
        ResponseHeaderTimeout: 5 * time.Second,  // 追加
        DisableKeepAlives:     false,
        DisableCompression:    false,
        ForceAttemptHTTP2:     true,
    },
}
```

#### メリット

1. **タイムアウト制御の細分化**: ヘッダー待機時間を個別に制御
2. **ハングアップ防止**: サーバーが応答を返さない場合のタイムアウト
3. **ベストプラクティス**: Go標準ライブラリの推奨設定

#### 推奨値

| 設定項目 | 推奨値 | 説明 |
|---------|--------|------|
| `ResponseHeaderTimeout` | 5秒 | レスポンスヘッダー待機のタイムアウト |
| `TLSHandshakeTimeout` | 5秒 | TLSハンドシェイクのタイムアウト |
| `ExpectContinueTimeout` | 1秒 | 100-continueレスポンスの待機時間 |

---

### 5. テンプレート実行エラーの詳細化

#### 背景

Go 1.6で`text/template`パッケージに`ExecError`型が追加され、テンプレート実行時のエラーをより詳細に取得できるようになりました。

#### Context7のベストプラクティス

```go
if err := tmpl.Execute(w, data); err != nil {
    if execErr, ok := err.(*template.ExecError); ok {
        // テンプレート名と元のエラーを取得
        log.Printf("Template %s: %v", execErr.Name, execErr.Err)
    }
    return fmt.Errorf("テンプレートの実行に失敗: %w", err)
}
```

#### 実装例

**現在の実装** (main.go:974-976):
```go
if err := tmpl.Execute(outputFile, data); err != nil {
    return fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
}
```

**推奨される実装**:
```go
if err := tmpl.Execute(outputFile, data); err != nil {
    // ExecError型にキャスト可能かチェック
    if execErr, ok := err.(*template.ExecError); ok {
        return fmt.Errorf("テンプレート %s の実行に失敗しました: %w", execErr.Name, execErr.Err)
    }
    return fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
}
```

#### メリット

1. **詳細なエラー情報**: テンプレート名と元のエラーを取得
2. **デバッグ容易性**: どのテンプレートで失敗したか特定可能
3. **エラーチェーンの保持**: `%w` でエラーをラップ

#### 注意点

- `text/template`の`ExecError`と`html/template`の`ExecError`は異なる型
- 本プロジェクトでは`html/template`を使用しているため、正しいパッケージをインポートする

```go
import "html/template"

// 型アサーション
if execErr, ok := err.(*template.ExecError); ok {
    // ...
}
```

---

### 6. Go言語バージョンのアップデート

#### 背景

Go 1.21以降、多くの改善とセキュリティ修正が追加されています。Context7のドキュメントによると、Go 1.23以降では以下の機能が追加されています:
- ジェネリックスの改善
- イテレータサポート
- テストAPIの改善
- セキュリティ強化

#### 推奨バージョン

**現在**: Go 1.21
**推奨**: Go 1.23 以降 (最新安定版)

#### 実装例

**現在の設定** (go.mod:3):
```go
go 1.21
```

**推奨される設定**:
```go
go 1.23
```

#### 変更が必要な箇所

1. **go.mod**:
```bash
# コマンドラインで実行
go mod edit -go=1.23
go mod tidy
```

2. **GitHub Actions** (.github/workflows/update-weather.yml):
```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.23'  # 1.21 から変更
```

3. **ローカル開発環境**:
```bash
# Go 1.23をインストール
go install golang.org/dl/go1.23@latest
go1.23 download
```

#### メリット

1. **パフォーマンス向上**: 最新のランタイム最適化
2. **セキュリティ強化**: 最新のセキュリティパッチ
3. **新機能の利用**: ジェネリックス、イテレータなど
4. **長期サポート**: 最新バージョンはサポート期間が長い

#### 互換性の確認

Go 1.23は下位互換性があるため、既存のコードはそのまま動作します。以下の手順で確認:

```bash
# テストを実行
go test -v

# ビルドを実行
go build

# 実行して確認
./kindle-tenki-dashboard
```

#### 注意点

- GitHub Actionsの実行環境も合わせて更新する
- テストが全て通過することを確認する
- 本番環境にデプロイする前にステージング環境で検証する

---

## 参考資料

### Context7ドキュメント

- **Go標準ライブラリ**: https://context7.com/golang/go
- **Go公式ドキュメント**: https://go.dev/doc/

### 関連するGoのバージョンとリリースノート

- **Go 1.13**: `context`と`http.NewRequestWithContext()`の追加
- **Go 1.19**: `encoding/csv.Reader.InputOffset()`, `encoding/xml.Decoder.InputPos()`の追加
- **Go 1.20**: `context.WithCancelCause()`の追加
- **Go 1.23**: ジェネリックスとイテレータの改善

### Go標準ライブラリのベストプラクティス

1. **エラーハンドリング**: `%w`を使ったエラーラッピング
2. **HTTPリクエスト**: `http.NewRequestWithContext()`の使用
3. **デコーディング**: `json.NewDecoder()`, `xml.NewDecoder()`の使用
4. **コンテキスト**: `context.WithTimeout()`, `context.WithCancel()`の活用
5. **リソース管理**: `defer`によるクリーンアップ

### 実装時の参考コード例

Context7から提供された以下のコード例を参考にできます:

1. **HTTP Client with Context**: コンテキスト付きHTTPリクエストの実装
2. **Stream JSON Decoding**: ストリーミングJSONデコード
3. **Error Handling**: エラーラッピングとチェーン
4. **Context Timeout**: タイムアウト付きコンテキスト

---

## 実装の優先順位とロードマップ

### Phase 1: 高優先度 (セキュリティ・信頼性)

**目標**: リクエストのタイムアウト制御とキャンセル機能の追加

- [ ] 1. HTTPリクエストにコンテキストを追加 (5つの関数)
  - [ ] `fetchWeatherData()`
  - [ ] `fetchNewsData()`
  - [ ] `fetchEconomyNewsData()`
  - [ ] `fetchHatenaBookmarks()`
  - [ ] `fetchKnowledgeHatenaBookmarks()`

**見積もり**: 1-2時間
**テスト**: 既存のテストが通過することを確認

---

### Phase 2: 中優先度 (パフォーマンス)

**目標**: メモリ効率とパフォーマンスの向上

- [ ] 2. JSON/XMLストリーミングデコードへの変更 (5箇所)
  - [ ] `fetchWeatherData()` - JSON
  - [ ] `fetchNewsData()` - XML
  - [ ] `fetchEconomyNewsData()` - XML
  - [ ] `fetchHatenaBookmarks()` - XML
  - [ ] `fetchKnowledgeHatenaBookmarks()` - XML

- [ ] 3. エラーハンドリングの改善
  - [ ] `parseTemperature()` - 入力値を含める
  - [ ] その他の関数 - エラーラッピングの一貫性向上

**見積もり**: 2-3時間
**テスト**: ユニットテストを追加し、エラーケースをカバー

---

### Phase 3: 低優先度 (将来的な改善)

**目標**: コードの保守性と将来性の向上

- [ ] 4. HTTP Transport設定の追加
  - [ ] `ResponseHeaderTimeout` の追加

- [ ] 5. テンプレート実行エラーの詳細化
  - [ ] `ExecError` による詳細なエラー情報取得

- [ ] 6. Go言語バージョンのアップデート
  - [ ] go.modをGo 1.23に更新
  - [ ] GitHub Actionsの設定を更新
  - [ ] テストとビルドの確認

**見積もり**: 1-2時間
**テスト**: 全てのテストが通過することを確認

---

## テストとデプロイの推奨手順

### 1. ローカル環境でのテスト

```bash
# 変更を適用する前にブランチを作成
git checkout -b feature/modernization

# 変更を適用

# テストを実行
go test -v

# ビルドを実行
go run main.go

# 生成されたHTMLを確認
python -m http.server 8000 --directory dist
```

### 2. GitHub Actionsでのテスト

```bash
# 変更をコミット
git add .
git commit -m "HTTPリクエストにコンテキストを追加した"

# リモートにプッシュ
git push origin feature/modernization

# GitHub ActionsでCIが通過することを確認
```

### 3. 段階的なデプロイ

1. 開発環境で動作確認
2. ステージング環境で検証
3. 本番環境にデプロイ

---

## まとめ

このモダン化提案を実装することで、以下の改善が期待できます:

### 改善効果の見積もり

| 項目 | 改善内容 | 効果 |
|------|----------|------|
| **信頼性** | コンテキストによるタイムアウト制御 | ⭐⭐⭐⭐⭐ |
| **パフォーマンス** | ストリーミングデコード | ⭐⭐⭐⭐ |
| **保守性** | エラーハンドリングの改善 | ⭐⭐⭐⭐ |
| **セキュリティ** | 最新Goバージョンへの更新 | ⭐⭐⭐ |
| **効率性** | HTTP Transport設定の追加 | ⭐⭐⭐ |

### 総合的な改善

- **コードの品質**: ベストプラクティスに準拠
- **保守性**: エラーメッセージが詳細で、デバッグが容易
- **パフォーマンス**: メモリ効率とリクエスト制御の向上
- **将来性**: 最新のGo機能を活用可能

### 次のアクション

1. このドキュメントをチームでレビュー
2. 優先度に基づいて実装計画を立てる
3. Phase 1から順次実装を開始
4. 各フェーズでテストとデプロイを実施

---

**最終更新日**: 2026-01-23
**作成者**: Claude Code (Context7レビューに基づく)
**参照元**: Context7 Go Documentation (golang/go, go.dev)
