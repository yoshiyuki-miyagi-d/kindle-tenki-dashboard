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
	"strconv"
	"time"
)

type WeatherData struct {
	Location       string           `json:"location"`
	Temperature    int              `json:"temperature"`
	FeelsLike      int              `json:"feelsLike"`
	Description    string           `json:"description"`
	UpdateTime     string           `json:"updateTime"`
	HourlyForecast []HourlyForecast `json:"hourlyForecast"`
	News           []NewsItem       `json:"news"`
}

type HourlyForecast struct {
	Time        string `json:"time"`
	Temp        int    `json:"temp"`
	Desc        string `json:"desc"`
	ChartHeight int    `json:"chartHeight"` // グラフ表示用の高さ(%)
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

type TsukumijimaWeatherResponse struct {
	PublicTime          string `json:"publicTime"`
	PublicTimeFormatted string `json:"publicTimeFormatted"`
	PublishingOffice    string `json:"publishingOffice"`
	Title               string `json:"title"`
	Forecasts           []struct {
		Date      string `json:"date"`
		DateLabel string `json:"dateLabel"`
		Telop     string `json:"telop"`
		Detail    struct {
			Weather string `json:"weather"`
			Wind    string `json:"wind"`
			Wave    string `json:"wave"`
		} `json:"detail"`
		Temperature struct {
			Min struct {
				Celsius string `json:"celsius"`
			} `json:"min"`
			Max struct {
				Celsius string `json:"celsius"`
			} `json:"max"`
		} `json:"temperature"`
		ChanceOfRain struct {
			T00_06 string `json:"T00_06"`
			T06_12 string `json:"T06_12"`
			T12_18 string `json:"T12_18"`
			T18_24 string `json:"T18_24"`
		} `json:"chanceOfRain"`
		Image struct {
			Title string `json:"title"`
			URL   string `json:"url"`
		} `json:"image"`
	} `json:"forecasts"`
	Location struct {
		Area       string `json:"area"`
		Prefecture string `json:"prefecture"`
		District   string `json:"district"`
		City       string `json:"city"`
	} `json:"location"`
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func fetchWeatherData() (*WeatherData, error) {
	cityCode := getEnv("CITY_CODE", "130010") // 東京のデフォルト
	weatherURL := fmt.Sprintf("https://weather.tsukumijima.net/api/forecast/city/%s", cityCode)

	// 天気データを取得
	resp, err := http.Get(weatherURL)
	if err != nil {
		log.Printf("⚠️  天気APIの取得に失敗しました: %v", err)
		log.Println("   サンプルデータを使用します")
		return getSampleData()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️  天気API Error: %d", resp.StatusCode)
		log.Println("   サンプルデータを使用します")
		return getSampleData()
	}

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

	weatherData := processWeatherData(weatherResponse)

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

func processWeatherData(response TsukumijimaWeatherResponse) *WeatherData {
	now := time.Now()

	// 今日の天気情報（最初の予報データを使用）
	var todayForecast = response.Forecasts[0]

	// 温度の処理（文字列から数値に変換）
	// 今日のデータがnullの場合は明日のデータを使用
	temperature := 0
	feelsLike := 0
	if todayForecast.Temperature.Max.Celsius != "" {
		if temp, err := parseTemperature(todayForecast.Temperature.Max.Celsius); err == nil {
			temperature = temp
			feelsLike = temp // 体感温度は最高気温で代用
		}
	} else if len(response.Forecasts) >= 2 && response.Forecasts[1].Temperature.Max.Celsius != "" {
		// 今日のデータがない場合は明日の最高気温を使用
		if temp, err := parseTemperature(response.Forecasts[1].Temperature.Max.Celsius); err == nil {
			temperature = temp
			feelsLike = temp
		}
	}

	// 時間別予報を生成（現在時刻以降の予報のみ表示）
	var hourlyForecast []HourlyForecast
	currentHour := now.Hour()

	if len(response.Forecasts) >= 2 {
		tomorrowForecast := response.Forecasts[1]
		var tomorrowMinTemp, tomorrowMaxTemp int
		if tomorrowForecast.Temperature.Min.Celsius != "" {
			if minTemp, err := parseTemperature(tomorrowForecast.Temperature.Min.Celsius); err == nil {
				tomorrowMinTemp = minTemp
			}
		}
		if tomorrowForecast.Temperature.Max.Celsius != "" {
			if maxTemp, err := parseTemperature(tomorrowForecast.Temperature.Max.Celsius); err == nil {
				tomorrowMaxTemp = maxTemp
			}
		}

		// 予報時刻のスロット（3時間ごと、48時間後まで）
		var forecastTimes []struct {
			hour  int
			label string
		}

		// 現在時刻から48時間後までの3時間ごとのスロットを生成
		for h := 0; h <= 72; h += 3 {
			hourInDay := h % 24
			forecastTimes = append(forecastTimes, struct {
				hour  int
				label string
			}{
				hour:  h,
				label: fmt.Sprintf("%02d:00", hourInDay),
			})
		}

		for _, ft := range forecastTimes {
			// 現在時刻以降の予報のみ追加
			if ft.hour > currentHour {
				var temp int
				var desc string

				// 24時以降は明日の予報
				if ft.hour >= 24 {
					// 明日の予報：時間帯によって気温を調整
					hourInDay := ft.hour % 24
					if hourInDay >= 0 && hourInDay < 6 {
						temp = tomorrowMinTemp
					} else if hourInDay >= 6 && hourInDay < 15 {
						temp = tomorrowMaxTemp
					} else if hourInDay >= 15 && hourInDay < 18 {
						temp = tomorrowMaxTemp - 2
					} else {
						temp = tomorrowMinTemp + 2
					}
					desc = tomorrowForecast.Telop
				} else {
					// 今日の予報
					// 時間帯によって気温を調整
					if ft.hour <= 15 {
						temp = temperature
					} else if ft.hour <= 18 {
						temp = temperature - 2
					} else {
						temp = temperature - 4
					}
					desc = todayForecast.Telop
				}

				hourlyForecast = append(hourlyForecast, HourlyForecast{
					Time: ft.label,
					Temp: temp,
					Desc: desc,
				})

				// 48時間後まで（最大15件）
				if len(hourlyForecast) >= 15 {
					break
				}
			}
		}
	}

	// グラフ表示用の高さを計算
	if len(hourlyForecast) > 0 {
		minTemp := hourlyForecast[0].Temp
		maxTemp := hourlyForecast[0].Temp
		for _, hf := range hourlyForecast {
			if hf.Temp < minTemp {
				minTemp = hf.Temp
			}
			if hf.Temp > maxTemp {
				maxTemp = hf.Temp
			}
		}

		// 最小気温を20%、最高気温を100%にマッピング
		tempRange := maxTemp - minTemp
		if tempRange == 0 {
			tempRange = 1 // ゼロ除算を防ぐ
		}

		for i := range hourlyForecast {
			heightPercent := 20 + ((hourlyForecast[i].Temp-minTemp)*80)/tempRange
			hourlyForecast[i].ChartHeight = heightPercent
		}
	}

	return &WeatherData{
		Location:       response.Location.City,
		Temperature:    temperature,
		FeelsLike:      feelsLike,
		Description:    todayForecast.Telop,
		UpdateTime:     now.Format("2006/01/02 15:04"),
		HourlyForecast: hourlyForecast,
		News:           []NewsItem{}, // 後で設定
	}
}

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

func getSampleData() (*WeatherData, error) {
	return &WeatherData{
		Location:    "東京",
		Temperature: 22,
		FeelsLike:   25,
		Description: "晴れ",
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

	// Go のhtml/template でパース（算術関数を追加）
	tmpl, err := template.New("index").Funcs(template.FuncMap{
		"mul": func(a, b int) int { return a * b },
		"sub": func(a, b int) int { return a - b },
	}).Parse(string(tmplContent))
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