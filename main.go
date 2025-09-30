package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type WeatherData struct {
	Location        string           `json:"location"`
	Temperature     int              `json:"temperature"`
	FeelsLike       int              `json:"feelsLike"`
	Description     string           `json:"description"`
	Humidity        int              `json:"humidity"`
	Pressure        int              `json:"pressure"`
	WindSpeed       float64          `json:"windSpeed"`
	UpdateTime      string           `json:"updateTime"`
	HourlyForecast  []HourlyForecast `json:"hourlyForecast"`
	News            []NewsItem       `json:"news"`
}

type HourlyForecast struct {
	Time string `json:"time"`
	Temp int    `json:"temp"`
	Desc string `json:"desc"`
}

type NewsItem struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
	PubDate     string `json:"pubDate"`
}

type NHKNewsRSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Title       string     `xml:"title"`
		Description string     `xml:"description"`
		Link        string     `xml:"link"`
		Items       []RSSItem  `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type OpenWeatherMapCurrent struct {
	Name string `json:"name"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
		Pressure  int     `json:"pressure"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
}

type OpenWeatherMapForecast struct {
	List []struct {
		Dt   int64 `json:"dt"`
		Main struct {
			Temp float64 `json:"temp"`
		} `json:"main"`
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`
	} `json:"list"`
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func fetchWeatherData() (*WeatherData, error) {
	apiKey := getEnv("OPENWEATHER_API_KEY", "YOUR_API_KEY")
	city := getEnv("CITY", "Tokyo")
	countryCode := getEnv("COUNTRY_CODE", "JP")

	if apiKey == "YOUR_API_KEY" {
		log.Println("⚠️  OPENWEATHER_API_KEY環境変数が設定されていません")
		log.Println("   デモ用のサンプルデータを使用します")
		return getSampleData()
	}

	currentURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s,%s&appid=%s&units=metric&lang=ja", city, countryCode, apiKey)
	forecastURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?q=%s,%s&appid=%s&units=metric&lang=ja", city, countryCode, apiKey)

	// 現在の天気データを取得
	currentResp, err := http.Get(currentURL)
	if err != nil {
		return nil, fmt.Errorf("現在の天気データの取得に失敗しました: %w", err)
	}
	defer currentResp.Body.Close()

	if currentResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Error: %d", currentResp.StatusCode)
	}

	currentBody, err := io.ReadAll(currentResp.Body)
	if err != nil {
		return nil, fmt.Errorf("現在の天気データの読み込みに失敗しました: %w", err)
	}

	var currentData OpenWeatherMapCurrent
	if err := json.Unmarshal(currentBody, &currentData); err != nil {
		return nil, fmt.Errorf("現在の天気データのパースに失敗しました: %w", err)
	}

	// 予報データを取得
	forecastResp, err := http.Get(forecastURL)
	if err != nil {
		return nil, fmt.Errorf("予報データの取得に失敗しました: %w", err)
	}
	defer forecastResp.Body.Close()

	if forecastResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Error: %d", forecastResp.StatusCode)
	}

	forecastBody, err := io.ReadAll(forecastResp.Body)
	if err != nil {
		return nil, fmt.Errorf("予報データの読み込みに失敗しました: %w", err)
	}

	var forecastData OpenWeatherMapForecast
	if err := json.Unmarshal(forecastBody, &forecastData); err != nil {
		return nil, fmt.Errorf("予報データのパースに失敗しました: %w", err)
	}

	weatherData := processWeatherData(currentData, forecastData)

	// ニュースデータを取得して追加
	news, err := fetchNewsData()
	if err != nil {
		log.Printf("⚠️  ニュースデータの取得に失敗しました: %v", err)
		log.Println("   サンプルのニュースデータを使用します")
		weatherData.News = getSampleNews()
	} else {
		weatherData.News = news
	}

	return weatherData, nil
}

func processWeatherData(current OpenWeatherMapCurrent, forecast OpenWeatherMapForecast) *WeatherData {
	now := time.Now()

	// 時間別予報（次の4つの時間帯）
	var hourlyForecast []HourlyForecast
	maxItems := 4
	if len(forecast.List) < maxItems {
		maxItems = len(forecast.List)
	}

	for i := 0; i < maxItems; i++ {
		item := forecast.List[i]
		dt := time.Unix(item.Dt, 0)
		hourlyForecast = append(hourlyForecast, HourlyForecast{
			Time: dt.Format("15:04"),
			Temp: int(item.Main.Temp + 0.5), // 四捨五入
			Desc: item.Weather[0].Description,
		})
	}

	return &WeatherData{
		Location:       current.Name,
		Temperature:    int(current.Main.Temp + 0.5), // 四捨五入
		FeelsLike:      int(current.Main.FeelsLike + 0.5),
		Description:    current.Weather[0].Description,
		Humidity:       current.Main.Humidity,
		Pressure:       current.Main.Pressure,
		WindSpeed:      current.Wind.Speed,
		UpdateTime:     now.Format("2006/01/02 15:04"),
		HourlyForecast: hourlyForecast,
		News:           []NewsItem{}, // 後で設定
	}
}

func getSampleData() (*WeatherData, error) {
	return &WeatherData{
		Location:    "東京",
		Temperature: 22,
		FeelsLike:   25,
		Description: "晴れ",
		Humidity:    60,
		Pressure:    1013,
		WindSpeed:   2.5,
		UpdateTime:  time.Now().Format("2006/01/02 15:04"),
		HourlyForecast: []HourlyForecast{
			{Time: "12:00", Temp: 23, Desc: "晴れ"},
			{Time: "15:00", Temp: 25, Desc: "晴れ"},
			{Time: "18:00", Temp: 21, Desc: "曇り"},
			{Time: "21:00", Temp: 19, Desc: "曇り"},
		},
		News: getSampleNews(),
	}, nil
}

func fetchNewsData() ([]NewsItem, error) {
	url := "https://www3.nhk.or.jp/rss/news/cat0.xml"

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ニュースRSSの取得に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ニュースRSS API Error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ニュースRSSの読み込みに失敗しました: %w", err)
	}

	var rss NHKNewsRSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, fmt.Errorf("ニュースRSSのパースに失敗しました: %w", err)
	}

	var news []NewsItem
	maxItems := 5 // トップ5件のニュースを表示
	if len(rss.Channel.Items) < maxItems {
		maxItems = len(rss.Channel.Items)
	}

	for i := 0; i < maxItems; i++ {
		item := rss.Channel.Items[i]
		// 日付をパースして表示用にフォーマット
		pubTime, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", item.PubDate)
		var formattedDate string
		if err != nil {
			formattedDate = item.PubDate
		} else {
			formattedDate = pubTime.Format("01/02 15:04")
		}

		news = append(news, NewsItem{
			Title:       item.Title,
			Link:        item.Link,
			Description: item.Description,
			PubDate:     formattedDate,
		})
	}

	return news, nil
}

func getSampleNews() []NewsItem {
	return []NewsItem{
		{
			Title:       "新浪氏の処遇 経済同友会が協議 審査会は\"辞任勧告が相当\"",
			Link:        "http://www3.nhk.or.jp/news/html/20250930/k10014936121000.html",
			Description: "経済同友会は、サプリメントをめぐる警察の捜査を受けて活動を自粛している、新浪剛史代表幹事の処遇について30日、理事会を開いて協議しています。",
			PubDate:     "09/30 12:19",
		},
		{
			Title:       "10月 値上げの食品 半年ぶり3000品目超 7割が「酒類・飲料」",
			Link:        "http://www3.nhk.or.jp/news/html/20250930/k10014935951000.html",
			Description: "10月に値上げされる食品は3000品目を超え、ことし4月以来、半年ぶりの高い水準になることが民間の調査でわかりました。",
			PubDate:     "09/30 11:26",
		},
		{
			Title:       "首都高発注の道路清掃入札で談合か 4社に立ち入り検査 公取委",
			Link:        "http://www3.nhk.or.jp/news/html/20250930/k10014936281000.html",
			Description: "首都高速道路が発注した道路清掃の入札をめぐり、東京や神奈川にある4社が、事前に落札する会社を調整する談合を繰り返した疑いがあるとして、公正取引委員会が、30日午前、立ち入り検査に入りました。",
			PubDate:     "09/30 11:46",
		},
	}
}

func generateHTML(data *WeatherData) error {
	// テンプレートファイルを読み込み
	templatePath := filepath.Join("src", "templates", "index.html")
	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("テンプレートファイルの読み込みに失敗しました: %w", err)
	}

	// Go のhtml/template でパース
	tmpl, err := template.New("index").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("テンプレートのパースに失敗しました: %w", err)
	}

	// docsディレクトリを作成
	docsDir := "docs"
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("docsディレクトリの作成に失敗しました: %w", err)
	}

	// HTMLファイルを生成
	outputPath := filepath.Join(docsDir, "index.html")
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("出力ファイルの作成に失敗しました: %w", err)
	}
	defer outputFile.Close()

	if err := tmpl.Execute(outputFile, data); err != nil {
		return fmt.Errorf("テンプレートの実行に失敗しました: %w", err)
	}

	// CSSファイルをコピー
	if err := copyCSS(); err != nil {
		return fmt.Errorf("CSSファイルのコピーに失敗しました: %w", err)
	}

	log.Printf("HTMLファイルとCSSファイルが生成されました")
	log.Printf("出力先: %s", outputPath)

	return nil
}

func copyCSS() error {
	srcPath := filepath.Join("src", "styles", "kindle.css")
	destDir := filepath.Join("docs", "styles")
	destPath := filepath.Join(destDir, "kindle.css")

	// stylesディレクトリを作成
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("stylesディレクトリの作成に失敗しました: %w", err)
	}

	// CSSファイルを読み込み
	cssContent, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("CSSファイルの読み込みに失敗しました: %w", err)
	}

	// CSSファイルを書き込み
	if err := os.WriteFile(destPath, cssContent, 0644); err != nil {
		return fmt.Errorf("CSSファイルの書き込みに失敗しました: %w", err)
	}

	return nil
}

func main() {
	log.Println("天気データを取得中...")

	data, err := fetchWeatherData()
	if err != nil {
		log.Fatalf("❌ 天気データの取得に失敗しました: %v", err)
	}

	if err := generateHTML(data); err != nil {
		log.Fatalf("❌ HTMLファイルの生成に失敗しました: %v", err)
	}

	log.Println("✅ ビルドが完了しました")
}