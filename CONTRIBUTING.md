# 開発ガイドライン

このドキュメントは、プロジェクトに貢献する際のルールとベストプラクティスをまとめたものです。

## コーディング規約

### 命名規則

#### 禁止されている単語
以下の単語は意味が曖昧なため使用禁止:
- `common`
- `util`
- `helper`
- `manager`
- `handler` (具体的な処理内容を示す名前を使用)

**❌ 悪い例:**
```go
func utilProcessData() {}
func commonHelper() {}
```

**✅ 良い例:**
```go
func parseWeatherResponse() {}
func formatTemperature() {}
```

#### ファイル名
- 小文字とアンダースコアを使用
- 内容を明確に表す名前にする

**例:**
- `weather_fetcher.go`
- `news_parser.go`
- `template_renderer.go`

#### 関数名
- キャメルケースを使用
- 動詞から始める
- 具体的な処理内容を示す

**例:**
- `fetchWeatherData()`
- `parseTemperature()`
- `generateHTML()`

#### 変数名
- キャメルケースを使用
- 意味のある名前にする
- 略語は避ける (一般的なものは除く: `id`, `url`, `api`)

**❌ 悪い例:**
```go
var tmp string
var d WeatherData
var resp *http.Response
```

**✅ 良い例:**
```go
var cityCode string
var weatherData WeatherData
var apiResponse *http.Response
```

#### 定数名
- 大文字のスネークケースを使用

**例:**
```go
const DEFAULT_CITY_CODE = "130010"
const MAX_NEWS_ITEMS = 5
```

### Go コーディングスタイル

#### 1. エラーハンドリング
- エラーは必ず処理する
- エラーメッセージは具体的にする
- フォールバック処理を実装する

**例:**
```go
data, err := fetchWeatherData()
if err != nil {
    log.Printf("⚠️  天気データの取得に失敗しました: %v", err)
    log.Println("   サンプルデータを使用します")
    return getSampleData()
}
```

#### 2. ログ出力
- 重要な処理の開始/終了をログに記録
- エラー発生時は原因を明記
- 絵文字を使用して視認性を向上

**例:**
```go
log.Println("天気データを取得中...")
log.Printf("⚠️  API呼び出しに失敗しました: %v", err)
log.Println("✅ ビルドが完了しました")
```

#### 3. 関数設計
- 1つの関数は1つの責務のみ
- 引数は最小限にする
- 長い関数は分割する (目安: 50行以内)

#### 4. コメント
- 複雑なロジックには説明を追加
- 公開関数にはGoDocコメントを記述
- コードを読めば分かる内容はコメント不要

**例:**
```go
// parseTemperature は文字列の気温データを整数に変換する。
// 空文字列やnullの場合はエラーを返す。
func parseTemperature(tempStr string) (int, error) {
    if tempStr == "" || tempStr == "null" {
        return 0, fmt.Errorf("empty temperature")
    }
    return strconv.Atoi(tempStr)
}
```

### HTML/CSS コーディングスタイル

#### 1. HTML
- インデントは2スペース
- 属性値はダブルクォートで囲む
- セマンティックなタグを使用

#### 2. CSS
- クラス名はケバブケースを使用
- E-ink最適化を優先
  - モノクロデザイン
  - 高コントラスト
  - 最小限のスタイル

**例:**
```css
.weather-container {
    background-color: white;
    color: black;
}

.temperature-value {
    font-size: 48px;
    font-weight: bold;
}
```

## Git コミット規約

### コミットメッセージの形式

#### 基本ルール
1. **日本語で記述する**
2. **完全な文章で終わらせる** (体言止め禁止)
3. **何をしたか明確に記述する**

#### 動詞の活用

**❌ 体言止め (禁止):**
```
機能を追加
バグの修正
ドキュメントの更新
```

**✅ 正しい形式:**
```
機能を追加した
バグを修正した
ドキュメントを更新した
```

### コミットメッセージの例

#### 新機能追加
```
48時間予報の気温グラフを実装した

- 時間別予報データから折れ線グラフを生成
- 最低気温20%、最高気温100%にマッピング
- レスポンシブデザインに対応
```

#### バグ修正
```
気温が取得できない場合のエラーハンドリングを修正した

明日の気温データも確認するようフォールバック処理を追加。
これにより夜間でも正しい気温が表示されるようになった。
```

#### リファクタリング
```
天気データ処理ロジックを関数に分割した

fetchWeatherData が肥大化していたため、
processWeatherData 関数を独立させて可読性を向上させた。
```

#### ドキュメント
```
API仕様書を追加した

外部APIの使用方法とレスポンス形式を文書化。
AIによる実装支援のための参考資料として作成した。
```

#### スタイル変更
```
フォントを游ゴシックに変更した

Kindleでの可読性を向上させるため、
游ゴシック体を優先フォントとして設定した。
```

### コミットの粒度

#### 1つのコミットには1つの変更
- 複数の機能を同時にコミットしない
- 無関係な変更は別々にコミット

**❌ 悪い例:**
```
天気グラフとニュース機能を追加し、READMEも更新した
```

**✅ 良い例:**
```
commit 1: 天気グラフ機能を実装した
commit 2: ニュースフィード機能を追加した
commit 3: READMEに新機能の説明を追加した
```

## プルリクエスト

### PR作成前のチェックリスト
- [ ] コードが正常にビルドされる
- [ ] 既存機能が壊れていない
- [ ] コーディング規約に従っている
- [ ] 必要に応じてドキュメントを更新した

### PRの説明
- 変更内容の要約
- 変更理由
- テスト方法
- スクリーンショット (UIに関する変更の場合)

## テスト

### 手動テスト手順

#### 1. ローカルビルド
```bash
go run main.go
```

#### 2. ローカルサーバー起動
```bash
python -m http.server 8000 --directory docs
```

#### 3. ブラウザで確認
- http://localhost:8000 にアクセス
- レスポンシブデザインの確認 (Kindleサイズ: 758x1024)
- データが正しく表示されているか確認

#### 4. Kindleでの確認 (推奨)
- GitHub Pagesにデプロイ後
- 実機で表示とレイアウトを確認

## 環境変数の管理

### ローカル開発
`.env` ファイルを作成:
```bash
CITY_CODE=130010  # 東京
```

### GitHub Actions
リポジトリの Settings > Secrets and variables > Actions で設定:
- Secrets: 機密情報 (APIキーなど)
- Variables: 非機密情報 (都市コードなど)

## トラブルシューティング

### ビルドが失敗する
1. Go のバージョンを確認 (1.21以上推奨)
2. 依存パッケージを確認: `go mod tidy`
3. エラーログを確認

### APIからデータが取得できない
1. ネットワーク接続を確認
2. APIのレート制限を確認
3. サンプルデータでビルドを試す

### Kindleで表示が崩れる
1. CSSのモノクロ対応を確認
2. フォントサイズが適切か確認
3. レスポンシブデザインを確認

## リソース

### 参考資料
- [Go コーディング規約](https://go.dev/doc/effective_go)
- [天気API仕様](./API.md)
- [アーキテクチャ設計](./ARCHITECTURE.md)
- [開発環境セットアップ](./DEVELOPMENT.md)

### 外部リンク
- [OpenWeatherMap API](https://openweathermap.org/api)
- [天気API (Tsukumijima)](https://weather.tsukumijima.net/)
- [NHK ニュースRSS](https://www.nhk.or.jp/toppage/rss/index.html)