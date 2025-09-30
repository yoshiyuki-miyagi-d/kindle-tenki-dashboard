package main

import (
	"encoding/json"
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
}

type HourlyForecast struct {
	Time string `json:"time"`
	Temp int    `json:"temp"`
	Desc string `json:"desc"`
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
		return getSampleData(), nil
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

	return processWeatherData(currentData, forecastData), nil
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
	}
}

func getSampleData() *WeatherData {
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