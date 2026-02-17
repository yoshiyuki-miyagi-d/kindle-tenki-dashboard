package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 定数定義
const (
	MaxHourlyForecastItems = 20 // 時間別予報の最大表示数
	MaxNewsItems           = 5  // 主要ニュースの最大表示数
	MaxEconomyNewsItems    = 10 // 経済ニュースの最大取得数(重複除外前)
	MaxHatenaItems         = 5  // はてなブックマークの最大表示数
	HTTPClientTimeout      = 10 * time.Second
)

// グローバルHTTPクライアント (Keep-Alive接続を再利用)
var httpClient = &http.Client{
	Timeout: HTTPClientTimeout,
	Transport: &http.Transport{
		MaxIdleConns:          100,              // アイドル接続数を増やす
		MaxIdleConnsPerHost:   10,               // ホストごとのアイドル接続数
		IdleConnTimeout:       90 * time.Second, // アイドル接続のタイムアウト
		TLSHandshakeTimeout:   5 * time.Second,  // TLSハンドシェイクのタイムアウト
		ExpectContinueTimeout: 1 * time.Second,  // 100-continueレスポンスの待機時間
		ResponseHeaderTimeout: 5 * time.Second,  // レスポンスヘッダー待機のタイムアウト
		DisableKeepAlives:     false,            // Keep-Aliveを有効化
		DisableCompression:    false,            // 圧縮を有効化
		ForceAttemptHTTP2:     true,             // HTTP/2を強制的に試行
	},
}

type WeatherData struct {
	Location            string           `json:"location"`
	Temperature         int              `json:"temperature"`
	MinTemp             int              `json:"minTemp"`
	MaxTemp             int              `json:"maxTemp"`
	FeelsLike           int              `json:"feelsLike"`
	Description         string           `json:"description"`
	WeatherIcon         string           `json:"weatherIcon"`    // 天気アイコン(絵文字)
	Wind                string           `json:"wind"`
	ChanceOfRain        []string         `json:"chanceOfRain"` // 6時間ごとの降水確率
	UpdateTime          string           `json:"updateTime"`
	HourlyForecast         []HourlyForecast `json:"hourlyForecast"`
	News                   []NewsItem       `json:"news"`
	EconomyNews            []NewsItem       `json:"economyNews"`           // 経済ニュース
	HatenaEntries          []HatenaEntry    `json:"hatenaEntries"`         // はてなブックマーク(総合)
	KnowledgeHatenaEntries []HatenaEntry    `json:"knowledgeHatenaEntries"` // はてなブックマーク(学び)
	DailyForecasts         []DailyForecast  `json:"dailyForecasts"`        // 3日間の予報
	HasMinTemp             bool             `json:"hasMinTemp"`             // 最低気温データが有効かどうか
	HasWeatherError         bool `json:"hasWeatherError"`         // 天気API失敗
	HasNewsError            bool `json:"hasNewsError"`            // ニュース(主要)API失敗
	HasEconomyNewsError     bool `json:"hasEconomyNewsError"`     // ニュース(経済)API失敗
	HasHatenaError          bool `json:"hasHatenaError"`          // はてなブックマーク(総合)API失敗
	HasKnowledgeHatenaError bool `json:"hasKnowledgeHatenaError"` // はてなブックマーク(学び)API失敗
}

type DailyForecast struct {
	Date        string `json:"date"`        // 日付ラベル(今日/明日/明後日)
	WeatherIcon string `json:"weatherIcon"` // 天気アイコン(絵文字)
	Description string `json:"description"` // 天気概況
	MaxTemp     int    `json:"maxTemp"`     // 最高気温
	MinTemp     int    `json:"minTemp"`     // 最低気温
	RainChance  string `json:"rainChance"`  // 降水確率(最大値)
}

type HourlyForecast struct {
	Time        string `json:"time"`
	Temp        int    `json:"temp"`
	Desc        string `json:"desc"`
	WeatherIcon string `json:"weatherIcon"` // 天気アイコン(絵文字)
	RainChance  string `json:"rainChance"`  // 降水確率
	ChartHeight int    `json:"chartHeight"` // グラフ表示用の高さ(%)
}

type NewsItem struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
	PubDate     string `json:"pubDate"`
}

type HatenaEntry struct {
	Title        string `json:"title"`
	Link         string `json:"link"`
	BookmarkLink string `json:"bookmarkLink"` // はてなブックマークページURL (https://b.hatena.ne.jp/entry/s/...)
	Description  string `json:"description"`
	PubDate      string `json:"pubDate"`
	Category     string `json:"category"` // カテゴリ(学び、テクノロジーなど)
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

type HatenaBookmarkRSS struct {
	XMLName xml.Name `xml:"RDF"`
	Items   []HatenaRSSItem `xml:"item"`
}

type HatenaRSSItem struct {
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	Date        string   `xml:"date"`
	Subjects    []string `xml:"subject"`
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

// getWeatherIcon は天気の説明文から絵文字アイコンを返す
func getWeatherIcon(description string) string {
	// 天気の説明文に基づいて絵文字を返す
	switch {
	case containsAny(description, []string{"晴", "快晴"}):
		return "☀️"
	case containsAny(description, []string{"曇", "くもり"}):
		return "☁️"
	case containsAny(description, []string{"雨", "雨天", "大雨", "豪雨"}):
		return "☔"
	case containsAny(description, []string{"雪", "大雪"}):
		return "⛄"
	case containsAny(description, []string{"雷", "雷雨"}):
		return "⚡"
	case containsAny(description, []string{"霧"}):
		return "🌫️"
	case containsAny(description, []string{"晴れ時々曇り", "晴れのち曇り", "晴時々曇"}):
		return "🌤️"
	case containsAny(description, []string{"曇り時々晴れ", "曇りのち晴れ", "曇時々晴"}):
		return "⛅"
	case containsAny(description, []string{"曇り時々雨", "曇りのち雨", "曇時々雨"}):
		return "🌧️"
	default:
		return "🌡️"
	}
}

// containsAny は文字列に指定されたいずれかの部分文字列が含まれるかチェックする
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// hatenaBookmarkURL は記事URLからはてなブックマークページのURLを生成する
// 例: "https://example.com/article" → "https://b.hatena.ne.jp/entry/s/example.com/article"
func hatenaBookmarkURL(articleURL string) string {
	trimmed := strings.TrimPrefix(articleURL, "https://")
	trimmed = strings.TrimPrefix(trimmed, "http://")
	return "https://b.hatena.ne.jp/entry/s/" + trimmed
}

func fetchWeatherData() (*WeatherData, error) {
	cityCode := getEnv("CITY_CODE", "130010") // 東京のデフォルト
	weatherURL := fmt.Sprintf("https://weather.tsukumijima.net/api/forecast/city/%s", cityCode)

	// コンテキストを作成 (5秒のタイムアウト)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// コンテキスト付きリクエストを作成
	req, err := http.NewRequestWithContext(ctx, "GET", weatherURL, nil)
	if err != nil {
		log.Printf("⚠️  天気APIリクエストの作成に失敗しました: %v", err)
		return weatherDataAllError(), nil
	}

	// リクエストを実行
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("⚠️  天気APIの取得に失敗しました: %v", err)
		return weatherDataAllError(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️  天気API Error: %d", resp.StatusCode)
		return weatherDataAllError(), nil
	}

	var weatherResponse TsukumijimaWeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResponse); err != nil {
		log.Printf("⚠️  天気データのパースに失敗しました: %v", err)
		return weatherDataAllError(), nil
	}

	weatherData := processWeatherData(weatherResponse)

	// === API呼び出しを並列化 ===
	// 結果を受け取るための構造体
	type newsResult struct {
		news []NewsItem
		err  error
	}
	type hatenaResult struct {
		entries []HatenaEntry
		err     error
	}

	// チャネルを作成
	newsCh := make(chan newsResult, 1)
	economyNewsCh := make(chan newsResult, 1)
	hatenaCh := make(chan hatenaResult, 1)
	knowledgeHatenaCh := make(chan hatenaResult, 1)

	// 4つのAPIを並列で呼び出し
	go func() {
		news, err := fetchNewsData()
		newsCh <- newsResult{news: news, err: err}
	}()

	go func() {
		economyNews, err := fetchEconomyNewsData()
		economyNewsCh <- newsResult{news: economyNews, err: err}
	}()

	go func() {
		hatena, err := fetchHatenaBookmarks()
		hatenaCh <- hatenaResult{entries: hatena, err: err}
	}()

	go func() {
		knowledgeHatena, err := fetchKnowledgeHatenaBookmarks()
		knowledgeHatenaCh <- hatenaResult{entries: knowledgeHatena, err: err}
	}()

	// 結果を受け取る
	newsRes := <-newsCh
	economyNewsRes := <-economyNewsCh
	hatenaRes := <-hatenaCh
	knowledgeHatenaRes := <-knowledgeHatenaCh

	// ニュースデータの処理
	if newsRes.err != nil {
		log.Printf("⚠️  ニュースデータの取得に失敗しました: %v", newsRes.err)
		weatherData.HasNewsError = true
	} else {
		weatherData.News = newsRes.news
	}

	// 経済ニュースデータの処理
	if economyNewsRes.err != nil {
		log.Printf("⚠️  経済ニュースデータの取得に失敗しました: %v", economyNewsRes.err)
		weatherData.HasEconomyNewsError = true
	} else {
		// 主要ニュースと重複する記事を経済ニュースから除外
		weatherData.EconomyNews = filterDuplicateNews(economyNewsRes.news, weatherData.News)
	}

	// はてなブックマーク(総合)データの処理
	if hatenaRes.err != nil {
		log.Printf("⚠️  はてなブックマーク(総合)データの取得に失敗しました: %v", hatenaRes.err)
		weatherData.HasHatenaError = true
	} else {
		weatherData.HatenaEntries = hatenaRes.entries
	}

	// はてなブックマーク(学び)データの処理
	if knowledgeHatenaRes.err != nil {
		log.Printf("⚠️  はてなブックマーク(学び)データの取得に失敗しました: %v", knowledgeHatenaRes.err)
		weatherData.HasKnowledgeHatenaError = true
	} else {
		// 総合はてなブックマークと重複するエントリーを学びから除外
		weatherData.KnowledgeHatenaEntries = filterDuplicateHatenaEntries(knowledgeHatenaRes.entries, weatherData.HatenaEntries)
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
	minTemp := 0
	maxTemp := 0
	feelsLike := 0
	hasMinTemp := false

	if todayForecast.Temperature.Max.Celsius != "" {
		if temp, err := parseTemperature(todayForecast.Temperature.Max.Celsius); err == nil {
			temperature = temp
			maxTemp = temp
			feelsLike = temp // 体感温度は最高気温で代用
		}
	} else if len(response.Forecasts) >= 2 && response.Forecasts[1].Temperature.Max.Celsius != "" {
		// 今日のデータがない場合は明日の最高気温を使用
		if temp, err := parseTemperature(response.Forecasts[1].Temperature.Max.Celsius); err == nil {
			temperature = temp
			maxTemp = temp
			feelsLike = temp
		}
	}

	if todayForecast.Temperature.Min.Celsius != "" {
		if temp, err := parseTemperature(todayForecast.Temperature.Min.Celsius); err == nil {
			minTemp = temp
			hasMinTemp = true // 最低気温データが有効
		}
	}

	// 風の情報
	wind := todayForecast.Detail.Wind

	// 降水確率（6時間ごと）
	chanceOfRain := []string{
		todayForecast.ChanceOfRain.T06_12,
		todayForecast.ChanceOfRain.T12_18,
		todayForecast.ChanceOfRain.T18_24,
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
				var rainChance string

				// 24時以降は明日の予報
				if ft.hour >= 24 {
					// 明日の予報：時間帯によって気温を調整
					hourInDay := ft.hour % 24
					if hourInDay >= 0 && hourInDay < 6 {
						temp = tomorrowMinTemp
						rainChance = tomorrowForecast.ChanceOfRain.T00_06
					} else if hourInDay >= 6 && hourInDay < 12 {
						temp = tomorrowMaxTemp
						rainChance = tomorrowForecast.ChanceOfRain.T06_12
					} else if hourInDay >= 12 && hourInDay < 18 {
						temp = tomorrowMaxTemp - 2
						rainChance = tomorrowForecast.ChanceOfRain.T12_18
					} else {
						temp = tomorrowMinTemp + 2
						rainChance = tomorrowForecast.ChanceOfRain.T18_24
					}
					desc = tomorrowForecast.Telop
				} else {
					// 今日の予報
					hourInDay := ft.hour
					// 時間帯によって気温と降水確率を調整
					if hourInDay >= 0 && hourInDay < 6 {
						temp = temperature - 4
						rainChance = todayForecast.ChanceOfRain.T00_06
					} else if hourInDay >= 6 && hourInDay < 12 {
						temp = temperature
						rainChance = todayForecast.ChanceOfRain.T06_12
					} else if hourInDay >= 12 && hourInDay < 18 {
						temp = temperature
						rainChance = todayForecast.ChanceOfRain.T12_18
					} else {
						temp = temperature - 2
						rainChance = todayForecast.ChanceOfRain.T18_24
					}
					desc = todayForecast.Telop
				}

				hourlyForecast = append(hourlyForecast, HourlyForecast{
					Time:        ft.label,
					Temp:        temp,
					Desc:        desc,
					WeatherIcon: getWeatherIcon(desc),
					RainChance:  rainChance,
				})

				// 48時間後まで（最大件数）
				if len(hourlyForecast) >= MaxHourlyForecastItems {
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

		// SVGのY座標系に合わせて計算 (上が小さい値、下が大きい値)
		// 最高気温を上部(y=20)、最低気温を下部(y=75)に配置
		tempRange := maxTemp - minTemp
		if tempRange == 0 {
			// 全て同じ気温の場合は中央に配置
			for i := range hourlyForecast {
				hourlyForecast[i].ChartHeight = 47 // (75 + 20) / 2
			}
		} else {
			for i := range hourlyForecast {
				// 最低気温 → heightPercent=75(下部), 最高気温 → heightPercent=20(上部)
				// Y座標は上が小さいので、温度が高いほど小さいY値にする
				heightPercent := 75 - ((hourlyForecast[i].Temp-minTemp)*55)/tempRange
				hourlyForecast[i].ChartHeight = heightPercent
			}
		}
	}

	// 3日間の予報を生成
	var dailyForecasts []DailyForecast
	dateLabels := []string{"今日", "明日", "明後日"}
	for i := 0; i < 3 && i < len(response.Forecasts); i++ {
		forecast := response.Forecasts[i]

		// 最高気温と最低気温を取得
		var dailyMaxTemp, dailyMinTemp int
		if forecast.Temperature.Max.Celsius != "" {
			if temp, err := parseTemperature(forecast.Temperature.Max.Celsius); err == nil {
				dailyMaxTemp = temp
			}
		}
		if forecast.Temperature.Min.Celsius != "" {
			if temp, err := parseTemperature(forecast.Temperature.Min.Celsius); err == nil {
				dailyMinTemp = temp
			}
		}

		// 降水確率の最大値を取得
		rainChances := []string{
			forecast.ChanceOfRain.T00_06,
			forecast.ChanceOfRain.T06_12,
			forecast.ChanceOfRain.T12_18,
			forecast.ChanceOfRain.T18_24,
		}
		maxRainChance := "0%"
		maxPercent := 0
		for _, rc := range rainChances {
			if rc != "" && rc != "-" {
				// %を除去して数値として比較
				percentStr := rc
				if len(rc) > 0 && rc[len(rc)-1] == '%' {
					percentStr = rc[:len(rc)-1]
				}
				currentPercent, err := strconv.Atoi(percentStr)
				if err == nil && currentPercent > maxPercent {
					maxPercent = currentPercent
					maxRainChance = rc
				}
			}
		}

		dailyForecasts = append(dailyForecasts, DailyForecast{
			Date:        dateLabels[i],
			WeatherIcon: getWeatherIcon(forecast.Telop),
			Description: forecast.Telop,
			MaxTemp:     dailyMaxTemp,
			MinTemp:     dailyMinTemp,
			RainChance:  maxRainChance,
		})
	}

	return &WeatherData{
		Location:       response.Location.City,
		Temperature:    temperature,
		MinTemp:        minTemp,
		MaxTemp:        maxTemp,
		FeelsLike:      feelsLike,
		Description:    todayForecast.Telop,
		WeatherIcon:    getWeatherIcon(todayForecast.Telop),
		Wind:           wind,
		ChanceOfRain:   chanceOfRain,
		UpdateTime:     now.Format("2006/01/02 15:04"),
		HourlyForecast: hourlyForecast,
		News:           []NewsItem{}, // 後で設定
		DailyForecasts: dailyForecasts,
		HasMinTemp:     hasMinTemp,
	}
}

func parseTemperature(tempStr string) (int, error) {
	if tempStr == "" {
		return 0, fmt.Errorf("temperature is empty")
	}
	if tempStr == "null" {
		return 0, fmt.Errorf("temperature is null")
	}
	temp, err := strconv.Atoi(tempStr)
	if err != nil {
		return 0, fmt.Errorf("invalid temperature format %q: %w", tempStr, err)
	}
	return temp, nil
}

// weatherDataAllError は全API失敗時のWeatherDataを返す
func weatherDataAllError() *WeatherData {
	return &WeatherData{
		UpdateTime:              time.Now().Format("2006/01/02 15:04"),
		HasWeatherError:         true,
		HasNewsError:            true,
		HasEconomyNewsError:     true,
		HasHatenaError:          true,
		HasKnowledgeHatenaError: true,
	}
}

func fetchNewsData() ([]NewsItem, error) {
	url := "https://www3.nhk.or.jp/rss/news/cat0.xml"

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ニュースRSS API Error: %d", resp.StatusCode)
	}

	var rss NHKNewsRSS
	if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
		return nil, fmt.Errorf("ニュースRSSのパースに失敗しました: %w", err)
	}

	var news []NewsItem
	maxItems := MaxNewsItems
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

func fetchEconomyNewsData() ([]NewsItem, error) {
	url := "https://www3.nhk.or.jp/rss/news/cat5.xml" // 経済ニュースRSS

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("経済ニュースRSSリクエストの作成に失敗しました: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("経済ニュースRSSの取得に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("経済ニュースRSS API Error: %d", resp.StatusCode)
	}

	var rss NHKNewsRSS
	if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
		return nil, fmt.Errorf("経済ニュースRSSのパースに失敗しました: %w", err)
	}

	var news []NewsItem
	maxItems := MaxEconomyNewsItems
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

// fetchHatenaBookmarks はてなブックマークの人気エントリーを取得する
func fetchHatenaBookmarks() ([]HatenaEntry, error) {
	url := "https://b.hatena.ne.jp/hotentry/all.rss"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("はてなブックマークRSSリクエストの作成に失敗しました: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("はてなブックマークRSSの取得に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("はてなブックマークRSS API Error: %d", resp.StatusCode)
	}

	var rss HatenaBookmarkRSS
	if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
		return nil, fmt.Errorf("はてなブックマークRSSのパースに失敗しました: %w", err)
	}

	var entries []HatenaEntry
	maxItems := MaxHatenaItems
	if len(rss.Items) < maxItems {
		maxItems = len(rss.Items)
	}

	for i := 0; i < maxItems; i++ {
		item := rss.Items[i]
		// 日付をパースして表示用にフォーマット
		// はてなブックマークの日付形式: 2025-10-30T16:24:16Z
		pubTime, err := time.Parse("2006-01-02T15:04:05Z", item.Date)
		var formattedDate string
		if err != nil {
			formattedDate = item.Date
		} else {
			formattedDate = pubTime.Format("01/02 15:04")
		}

		// カテゴリを取得 (最初のSubjectを使用)
		category := ""
		if len(item.Subjects) > 0 {
			category = item.Subjects[0]
		}

		entries = append(entries, HatenaEntry{
			Title:        item.Title,
			Link:         item.Link,
			BookmarkLink: hatenaBookmarkURL(item.Link),
			Description:  item.Description,
			PubDate:      formattedDate,
			Category:     category,
		})
	}

	return entries, nil
}

// fetchKnowledgeHatenaBookmarks はてなブックマークの学びカテゴリの人気エントリーを取得する
func fetchKnowledgeHatenaBookmarks() ([]HatenaEntry, error) {
	url := "https://b.hatena.ne.jp/hotentry/knowledge.rss"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("はてなブックマーク(学び)RSSリクエストの作成に失敗しました: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("はてなブックマーク(学び)RSSの取得に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("はてなブックマーク(学び)RSS API Error: %d", resp.StatusCode)
	}

	var rss HatenaBookmarkRSS
	if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
		return nil, fmt.Errorf("はてなブックマーク(学び)RSSのパースに失敗しました: %w", err)
	}

	var entries []HatenaEntry
	maxItems := MaxHatenaItems
	if len(rss.Items) < maxItems {
		maxItems = len(rss.Items)
	}

	for i := 0; i < maxItems; i++ {
		item := rss.Items[i]
		// 日付をパースして表示用にフォーマット
		pubTime, err := time.Parse("2006-01-02T15:04:05Z", item.Date)
		var formattedDate string
		if err != nil {
			formattedDate = item.Date
		} else {
			formattedDate = pubTime.Format("01/02 15:04")
		}

		// カテゴリを取得 (最初のSubjectを使用)
		category := ""
		if len(item.Subjects) > 0 {
			category = item.Subjects[0]
		}

		entries = append(entries, HatenaEntry{
			Title:        item.Title,
			Link:         item.Link,
			BookmarkLink: hatenaBookmarkURL(item.Link),
			Description:  item.Description,
			PubDate:      formattedDate,
			Category:     category,
		})
	}

	return entries, nil
}

func filterDuplicateNews(economyNews []NewsItem, mainNews []NewsItem) []NewsItem {
	// 主要ニュースのタイトルをマップに格納
	mainTitles := make(map[string]bool)
	for _, item := range mainNews {
		mainTitles[item.Title] = true
	}

	// 重複しない経済ニュースを抽出し、最大件数になるまで追加
	var filtered []NewsItem
	for _, item := range economyNews {
		if !mainTitles[item.Title] {
			filtered = append(filtered, item)
			if len(filtered) >= MaxNewsItems {
				break
			}
		}
	}

	return filtered
}

func filterDuplicateHatenaEntries(knowledgeEntries []HatenaEntry, generalEntries []HatenaEntry) []HatenaEntry {
	// 総合はてなブックマークのタイトルをマップに格納
	generalTitles := make(map[string]bool)
	for _, item := range generalEntries {
		generalTitles[item.Title] = true
	}

	// 重複しない学びはてなブックマークを抽出し、最大件数になるまで追加
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

	// distディレクトリを作成
	distDir := "dist"
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return fmt.Errorf("distディレクトリの作成に失敗しました: %w", err)
	}

	// HTMLファイルを生成
	outputPath := filepath.Join(distDir, "index.html")
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
	destDir := filepath.Join("dist", "styles")
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