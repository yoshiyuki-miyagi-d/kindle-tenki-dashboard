# 開発環境セットアップガイド

このドキュメントは、ローカル開発環境のセットアップ方法とデバッグ手順をまとめたものです。

## 必要な環境

### 必須

| ツール | バージョン | 用途 |
|--------|-----------|------|
| Go | 1.21以上 | アプリケーションのビルドと実行 |
| Git | 2.x以上 | バージョン管理 |

### 推奨

| ツール | バージョン | 用途 |
|--------|-----------|------|
| Python | 3.x | ローカルHTTPサーバー (開発用) |
| jq | 最新版 | JSONデータの確認 (デバッグ用) |
| xmllint | 最新版 | XMLデータの確認 (デバッグ用) |

## セットアップ手順

### 1. リポジトリのクローン

```bash
git clone https://github.com/your-username/kindle-tenki-dashbaord.git
cd kindle-tenki-dashbaord
```

### 2. Go の環境確認

```bash
# バージョン確認
go version

# 1.21以上であることを確認
# 出力例: go version go1.21.0 darwin/arm64
```

#### Go のインストール (必要な場合)

**macOS:**
```bash
brew install go
```

**NixOS:**
```nix
# configuration.nix に追加
environment.systemPackages = with pkgs; [
  go_1_21
];
```

**Linux (その他):**
```bash
# 公式サイトからダウンロード
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### 3. 環境変数の設定

`.env` ファイルを作成 (オプション):
```bash
cp .env.example .env
```

`.env` の内容を編集:
```bash
# 都市コードを設定 (デフォルト: 東京 130010)
CITY_CODE=130010
```

**主要な都市コード:**
- 札幌: `016010`
- 東京: `130010`
- 横浜: `140010`
- 名古屋: `230010`
- 大阪: `270000`
- 京都: `260010`
- 福岡: `400010`
- 那覇: `471010`

### 4. ビルドと実行

```bash
# アプリケーションをビルド・実行
go run main.go
```

成功すると以下のような出力が表示される:
```
2025/10/02 12:00:00 天気データを取得中...
2025/10/02 12:00:05 HTMLファイルとCSSファイルが生成されました
2025/10/02 12:00:05 出力先: docs/index.html
2025/10/02 12:00:05 ✅ ビルドが完了しました
```

### 5. ローカルサーバーの起動

生成されたHTMLを確認:

#### Python 3.x の場合
```bash
python -m http.server 8000 --directory docs
```

#### Python 2.x の場合
```bash
cd docs
python -m SimpleHTTPServer 8000
```

ブラウザで http://localhost:8000 にアクセス。

### 6. 動作確認

- 天気情報が正しく表示されているか
- 気温グラフが表示されているか
- ニュースフィードが表示されているか
- レスポンシブデザインが機能しているか

## ディレクトリ構造

```
kindle-tenki-dashbaord/
├── .github/
│   └── workflows/
│       └── update-weather.yml    # GitHub Actions設定
├── docs/                         # 生成されるファイル (GitHub Pages公開ディレクトリ)
│   ├── index.html               # 生成されたHTML
│   └── styles/
│       └── kindle.css           # コピーされたCSS
├── src/                          # ソースファイル
│   ├── templates/
│   │   └── index.html           # HTMLテンプレート
│   └── styles/
│       └── kindle.css           # CSSソースファイル
├── main.go                       # メインアプリケーション
├── go.mod                        # Go依存関係管理
├── .env.example                  # 環境変数のサンプル
├── .gitignore                    # Git除外設定
├── README.md                     # プロジェクト概要
├── ARCHITECTURE.md               # アーキテクチャ設計書
├── CONTRIBUTING.md               # 開発ガイドライン
├── API.md                        # API仕様書
└── DEVELOPMENT.md                # このファイル
```

## 開発ワークフロー

### 機能追加の流れ

1. **ブランチ作成**
   ```bash
   git checkout -b feature/new-feature-name
   ```

2. **コーディング**
   - `main.go` を編集
   - `src/templates/index.html` を編集
   - `src/styles/kindle.css` を編集

3. **ローカルテスト**
   ```bash
   go run main.go
   python -m http.server 8000 --directory docs
   ```

4. **コミット**
   ```bash
   git add .
   git commit -m "新機能を実装した"
   ```

5. **プッシュとPR作成**
   ```bash
   git push origin feature/new-feature-name
   # GitHubでPRを作成
   ```

### バグ修正の流れ

1. **Issueを確認**
   - 問題の再現手順を確認
   - 期待される動作を確認

2. **ブランチ作成**
   ```bash
   git checkout -b fix/bug-description
   ```

3. **デバッグ**
   - ログ出力を追加
   - エラーの原因を特定

4. **修正とテスト**
   ```bash
   go run main.go
   # 修正が反映されていることを確認
   ```

5. **コミットとPR**
   ```bash
   git commit -m "バグを修正した"
   git push origin fix/bug-description
   ```

## デバッグ方法

### 1. ログ出力の追加

```go
log.Println("デバッグ: 変数の値 =", variable)
log.Printf("デバッグ: %+v", structVariable)
```

### 2. APIレスポンスの確認

#### 天気API
```bash
curl -s https://weather.tsukumijima.net/api/forecast/city/130010 | jq
```

#### ニュースRSS
```bash
curl -s https://www3.nhk.or.jp/rss/news/cat0.xml | xmllint --format -
```

### 3. 生成されたHTMLの確認

```bash
cat docs/index.html
```

### 4. ブラウザの開発者ツール
- F12 または Cmd+Option+I で開く
- Consoleでエラーを確認
- Networkタブでリソース読み込みを確認
- Elementsタブでレイアウトを確認

### 5. Kindle実機でのデバッグ

1. **GitHub Pagesにデプロイ**
   ```bash
   git push origin main
   # GitHub Actionsが自動実行
   ```

2. **Kindleでアクセス**
   - 実験的ブラウザを起動
   - `https://your-username.github.io/kindle-tenki-dashbaord/` にアクセス

3. **表示確認**
   - レイアウトの崩れ
   - フォントサイズ
   - 読みやすさ

### 6. レスポンシブデザインのテスト

ブラウザの開発者ツールで:
- デバイスツールバーを有効化 (Cmd+Shift+M)
- カスタムサイズ: 758x1024 (Kindle Paperwhite)
- モノクロシミュレーション (Rendering > Emulate CSS media feature > prefers-color-scheme: monochrome)

## よくある問題と解決方法

### 問題1: `go: command not found`

**原因:** Go がインストールされていない、またはPATHが通っていない

**解決方法:**
```bash
# Goをインストール
brew install go  # macOS
# または公式サイトからダウンロード

# PATHを確認
echo $PATH | grep go

# PATHに追加 (必要な場合)
export PATH=$PATH:/usr/local/go/bin
```

### 問題2: API呼び出しが失敗する

**原因:** ネットワーク接続の問題、APIサーバーのダウン、レート制限

**解決方法:**
```bash
# ネットワーク接続を確認
curl -I https://weather.tsukumijima.net

# サンプルデータでテスト
# main.go が自動的にフォールバックする
go run main.go
```

### 問題3: HTMLが生成されない

**原因:** テンプレートファイルが見つからない、パース エラー

**解決方法:**
```bash
# テンプレートファイルの存在確認
ls -l src/templates/index.html

# エラーメッセージを確認
go run main.go 2>&1 | grep -i error
```

### 問題4: CSSが適用されない

**原因:** CSSファイルのパスが間違っている、ファイルがコピーされていない

**解決方法:**
```bash
# CSSファイルの存在確認
ls -l docs/styles/kindle.css

# HTMLのlinkタグを確認
grep stylesheet docs/index.html

# 手動でコピー (テスト用)
cp src/styles/kindle.css docs/styles/
```

### 問題5: Kindleで表示が崩れる

**原因:** Kindleブラウザの制限、CSSの互換性問題

**解決方法:**
1. **シンプルなCSSを使用**
   - Flexbox、Gridは避ける
   - Float、Inline-blockを使用

2. **フォントサイズを調整**
   - 最小14px以上を推奨

3. **カラーをモノクロに**
   - 黒 (#000000) と白 (#FFFFFF) のみ使用

## パフォーマンス最適化

### ビルド時間の短縮

#### Go のビルドキャッシュを活用
```bash
# ビルドキャッシュの確認
go env GOCACHE

# キャッシュのクリア (必要な場合)
go clean -cache
```

### APIレスポンスの高速化

#### タイムアウトの設定
```go
client := &http.Client{
    Timeout: 10 * time.Second,
}
```

#### 並列処理
```go
// 天気とニュースを並列取得
var wg sync.WaitGroup
var weatherData *WeatherData
var newsData []NewsItem

wg.Add(2)

go func() {
    defer wg.Done()
    weatherData, _ = fetchWeatherData()
}()

go func() {
    defer wg.Done()
    newsData, _ = fetchNewsData()
}()

wg.Wait()
```

## GitHub Actions (CI/CD)

### ローカルでのワークフローテスト

[act](https://github.com/nektos/act) を使用:
```bash
# actをインストール
brew install act  # macOS

# ワークフローを実行
act -j build
```

### ワークフローのデバッグ

```yaml
# .github/workflows/update-weather.yml
- name: デバッグ情報を出力
  run: |
    echo "Go version:"
    go version
    echo "Environment:"
    env | sort
```

## テスト

### 手動テスト

#### 基本動作テスト
```bash
# 1. ビルド
go run main.go

# 2. HTMLの生成確認
test -f docs/index.html && echo "OK" || echo "NG"

# 3. CSSのコピー確認
test -f docs/styles/kindle.css && echo "OK" || echo "NG"

# 4. HTMLの内容確認
grep -q "天気" docs/index.html && echo "OK" || echo "NG"
```

#### エラーハンドリングのテスト
```bash
# 1. 不正な都市コードでテスト
CITY_CODE=999999 go run main.go
# サンプルデータにフォールバックすることを確認

# 2. ネットワーク切断状態でテスト
# (Wi-Fiをオフにして実行)
go run main.go
# サンプルデータにフォールバックすることを確認
```

### ユニットテスト (今後の実装)

```go
// main_test.go
package main

import "testing"

func TestParseTemperature(t *testing.T) {
    tests := []struct {
        input    string
        expected int
        hasError bool
    }{
        {"25", 25, false},
        {"", 0, true},
        {"null", 0, true},
        {"abc", 0, true},
    }

    for _, test := range tests {
        result, err := parseTemperature(test.input)
        if test.hasError && err == nil {
            t.Errorf("期待: エラー, 実際: nil")
        }
        if !test.hasError && result != test.expected {
            t.Errorf("期待: %d, 実際: %d", test.expected, result)
        }
    }
}
```

実行:
```bash
go test -v
```

## エディタ設定

### VS Code

#### 推奨拡張機能
- **Go** (golang.go) - Go言語サポート
- **HTML CSS Support** - HTML/CSSサポート
- **GitLens** - Git統合

#### settings.json
```json
{
  "go.formatTool": "gofmt",
  "go.lintTool": "golangci-lint",
  "editor.formatOnSave": true,
  "[go]": {
    "editor.tabSize": 4,
    "editor.insertSpaces": false
  },
  "[html]": {
    "editor.tabSize": 2,
    "editor.insertSpaces": true
  },
  "[css]": {
    "editor.tabSize": 2,
    "editor.insertSpaces": true
  }
}
```

### Vim/Neovim

#### プラグイン
```vim
" vim-go
Plug 'fatih/vim-go'

" 自動フォーマット
autocmd BufWritePre *.go :GoFmt
```

## リソース

### 公式ドキュメント
- [Go公式ドキュメント](https://go.dev/doc/)
- [html/templateパッケージ](https://pkg.go.dev/html/template)
- [encoding/jsonパッケージ](https://pkg.go.dev/encoding/json)
- [encoding/xmlパッケージ](https://pkg.go.dev/encoding/xml)

### プロジェクト内ドキュメント
- [README.md](./README.md) - プロジェクト概要
- [ARCHITECTURE.md](./ARCHITECTURE.md) - アーキテクチャ設計
- [CONTRIBUTING.md](./CONTRIBUTING.md) - 開発ガイドライン
- [API.md](./API.md) - API仕様書

### 外部リンク
- [天気API](https://weather.tsukumijima.net/)
- [NHK ニュースRSS](https://www.nhk.or.jp/toppage/rss/index.html)
- [GitHub Actions ドキュメント](https://docs.github.com/en/actions)
- [Kindle Paperwhite仕様](https://www.amazon.co.jp/kindle-paperwhite)

## サポート

質問や問題がある場合:
1. [Issues](https://github.com/your-username/kindle-tenki-dashbaord/issues) を検索
2. 既存のIssueがなければ新規作成
3. 以下の情報を含める:
   - Go のバージョン
   - OS とバージョン
   - エラーメッセージ
   - 再現手順