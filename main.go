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

	// SVG気温グラフのY座標定数
	ChartTopY    = 20 // 最高気温のY座標(上部)
	ChartBottomY = 75 // 最低気温のY座標(下部)
	ChartMiddleY = 47 // 全気温同一時の中央Y座標 (ChartTopY + ChartBottomY) / 2
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

// fetchAndDecodeXML はURLからXMLを取得しデコードする
func fetchAndDecodeXML[T any](url string) (T, error) {
	var zero T

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return zero, fmt.Errorf("RSSリクエストの作成に失敗しました(%s): %w", url, err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return zero, fmt.Errorf("RSSの取得に失敗しました(%s): %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return zero, fmt.Errorf("RSS API Error(%s): %d", url, resp.StatusCode)
	}

	var result T
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return zero, fmt.Errorf("RSSのパースに失敗しました(%s): %w", url, err)
	}

	return result, nil
}

// formatRSSDate は日付文字列を "01/02 15:04" 形式に変換する
func formatRSSDate(dateStr, layout string) string {
	pubTime, err := time.Parse(layout, dateStr)
	if err != nil {
		return dateStr
	}
	return pubTime.Format("01/02 15:04")
}

// rssResults はRSS並列取得の結果をまとめる構造体
type rssResults struct {
	news           []NewsItem
	economyNews    []NewsItem
	hatenaEntries  []HatenaEntry
	knowledgeHatena []HatenaEntry
	hasNewsError            bool
	hasEconomyNewsError     bool
	hasHatenaError          bool
	hasKnowledgeHatenaError bool
}

func fetchWeatherData() (*WeatherData, error) {
	cityCode := getEnv("CITY_CODE", "130010") // 東京のデフォルト
	weatherURL := fmt.Sprintf("https://weather.tsukumijima.net/api/forecast/city/%s", cityCode)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", weatherURL, nil)
	if err != nil {
		log.Printf("⚠️  天気APIリクエストの作成に失敗しました: %v", err)
		return weatherDataAllError(), nil
	}

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
	rss := fetchAllRSSData()

	weatherData.News = rss.news
	weatherData.EconomyNews = rss.economyNews
	weatherData.HatenaEntries = rss.hatenaEntries
	weatherData.KnowledgeHatenaEntries = rss.knowledgeHatena
	weatherData.HasNewsError = rss.hasNewsError
	weatherData.HasEconomyNewsError = rss.hasEconomyNewsError
	weatherData.HasHatenaError = rss.hasHatenaError
	weatherData.HasKnowledgeHatenaError = rss.hasKnowledgeHatenaError

	return weatherData, nil
}

// fetchAllRSSData は4つのRSS APIをgoroutineで並列取得し、エラー処理・重複除外まで行う
func fetchAllRSSData() rssResults {
	type newsResult struct {
		news []NewsItem
		err  error
	}
	type hatenaResult struct {
		entries []HatenaEntry
		err     error
	}

	newsCh := make(chan newsResult, 1)
	economyNewsCh := make(chan newsResult, 1)
	hatenaCh := make(chan hatenaResult, 1)
	knowledgeHatenaCh := make(chan hatenaResult, 1)

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

	newsRes := <-newsCh
	economyNewsRes := <-economyNewsCh
	hatenaRes := <-hatenaCh
	knowledgeHatenaRes := <-knowledgeHatenaCh

	var result rssResults

	if newsRes.err != nil {
		log.Printf("⚠️  ニュースデータの取得に失敗しました: %v", newsRes.err)
		result.hasNewsError = true
	} else {
		result.news = newsRes.news
	}

	if economyNewsRes.err != nil {
		log.Printf("⚠️  経済ニュースデータの取得に失敗しました: %v", economyNewsRes.err)
		result.hasEconomyNewsError = true
	} else {
		result.economyNews = filterDuplicateNews(economyNewsRes.news, result.news)
	}

	if hatenaRes.err != nil {
		log.Printf("⚠️  はてなブックマーク(総合)データの取得に失敗しました: %v", hatenaRes.err)
		result.hasHatenaError = true
	} else {
		result.hatenaEntries = hatenaRes.entries
	}

	if knowledgeHatenaRes.err != nil {
		log.Printf("⚠️  はてなブックマーク(学び)データの取得に失敗しました: %v", knowledgeHatenaRes.err)
		result.hasKnowledgeHatenaError = true
	} else {
		result.knowledgeHatena = filterDuplicateHatenaEntries(knowledgeHatenaRes.entries, result.hatenaEntries)
	}

	return result
}

func processWeatherData(response TsukumijimaWeatherResponse) *WeatherData {
	now := time.Now()
	todayForecast := response.Forecasts[0]

	temperature, minTemp, maxTemp, feelsLike, hasMinTemp := extractTemperatures(response.Forecasts)
	hourlyForecast := generateHourlyForecasts(response.Forecasts, now.Hour(), temperature)
	calculateChartHeights(hourlyForecast)
	dailyForecasts := generateDailyForecasts(response.Forecasts)

	return &WeatherData{
		Location:       response.Location.City,
		Temperature:    temperature,
		MinTemp:        minTemp,
		MaxTemp:        maxTemp,
		FeelsLike:      feelsLike,
		Description:    todayForecast.Telop,
		WeatherIcon:    getWeatherIcon(todayForecast.Telop),
		Wind:           todayForecast.Detail.Wind,
		ChanceOfRain:   []string{todayForecast.ChanceOfRain.T06_12, todayForecast.ChanceOfRain.T12_18, todayForecast.ChanceOfRain.T18_24},
		UpdateTime:     now.Format("2006/01/02 15:04"),
		HourlyForecast: hourlyForecast,
		News:           []NewsItem{},
		DailyForecasts: dailyForecasts,
		HasMinTemp:     hasMinTemp,
	}
}

// extractTemperatures は予報データから今日/明日の気温情報を抽出する
func extractTemperatures(forecasts []struct {
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
}) (temperature, minTemp, maxTemp, feelsLike int, hasMinTemp bool) {
	todayForecast := forecasts[0]

	if todayForecast.Temperature.Max.Celsius != "" {
		if temp, err := parseTemperature(todayForecast.Temperature.Max.Celsius); err == nil {
			temperature = temp
			maxTemp = temp
			feelsLike = temp
		}
	} else if len(forecasts) >= 2 && forecasts[1].Temperature.Max.Celsius != "" {
		if temp, err := parseTemperature(forecasts[1].Temperature.Max.Celsius); err == nil {
			temperature = temp
			maxTemp = temp
			feelsLike = temp
		}
	}

	if todayForecast.Temperature.Min.Celsius != "" {
		if temp, err := parseTemperature(todayForecast.Temperature.Min.Celsius); err == nil {
			minTemp = temp
			hasMinTemp = true
		}
	}

	return
}

// generateHourlyForecasts は現在時刻以降の3時間ごとの時間別予報を生成する
func generateHourlyForecasts(forecasts []struct {
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
}, currentHour int, todayMaxTemp int) []HourlyForecast {
	if len(forecasts) < 2 {
		return nil
	}

	todayForecast := forecasts[0]
	tomorrowForecast := forecasts[1]

	var tomorrowMinTemp, tomorrowMaxTemp int
	if tomorrowForecast.Temperature.Min.Celsius != "" {
		if temp, err := parseTemperature(tomorrowForecast.Temperature.Min.Celsius); err == nil {
			tomorrowMinTemp = temp
		}
	}
	if tomorrowForecast.Temperature.Max.Celsius != "" {
		if temp, err := parseTemperature(tomorrowForecast.Temperature.Max.Celsius); err == nil {
			tomorrowMaxTemp = temp
		}
	}

	// 予報時刻のスロット(3時間ごと、72時間分)
	type forecastSlot struct {
		hour  int
		label string
	}
	var forecastTimes []forecastSlot
	for h := 0; h <= 72; h += 3 {
		forecastTimes = append(forecastTimes, forecastSlot{
			hour:  h,
			label: fmt.Sprintf("%02d:00", h%24),
		})
	}

	var hourlyForecast []HourlyForecast
	for _, ft := range forecastTimes {
		if ft.hour <= currentHour {
			continue
		}

		var temp int
		var desc string
		var rainChance string

		if ft.hour >= 24 {
			// 明日の予報
			hourInDay := ft.hour % 24
			if hourInDay < 6 {
				temp = tomorrowMinTemp
				rainChance = tomorrowForecast.ChanceOfRain.T00_06
			} else if hourInDay < 12 {
				temp = tomorrowMaxTemp
				rainChance = tomorrowForecast.ChanceOfRain.T06_12
			} else if hourInDay < 18 {
				temp = tomorrowMaxTemp - 2
				rainChance = tomorrowForecast.ChanceOfRain.T12_18
			} else {
				temp = tomorrowMinTemp + 2
				rainChance = tomorrowForecast.ChanceOfRain.T18_24
			}
			desc = tomorrowForecast.Telop
		} else {
			// 今日の予報
			if ft.hour < 6 {
				temp = todayMaxTemp - 4
				rainChance = todayForecast.ChanceOfRain.T00_06
			} else if ft.hour < 12 {
				temp = todayMaxTemp
				rainChance = todayForecast.ChanceOfRain.T06_12
			} else if ft.hour < 18 {
				temp = todayMaxTemp
				rainChance = todayForecast.ChanceOfRain.T12_18
			} else {
				temp = todayMaxTemp - 2
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

		if len(hourlyForecast) >= MaxHourlyForecastItems {
			break
		}
	}

	return hourlyForecast
}

// calculateChartHeights はSVG座標系でのグラフ高さを計算する(スライスを直接変更)
func calculateChartHeights(hourlyForecast []HourlyForecast) {
	if len(hourlyForecast) == 0 {
		return
	}

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

	tempRange := maxTemp - minTemp
	if tempRange == 0 {
		for i := range hourlyForecast {
			hourlyForecast[i].ChartHeight = ChartMiddleY
		}
		return
	}

	chartYRange := ChartBottomY - ChartTopY
	for i := range hourlyForecast {
		hourlyForecast[i].ChartHeight = ChartBottomY - ((hourlyForecast[i].Temp-minTemp)*chartYRange)/tempRange
	}
}

// generateDailyForecasts は3日間の日別予報を生成する
func generateDailyForecasts(forecasts []struct {
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
}) []DailyForecast {
	dateLabels := []string{"今日", "明日", "明後日"}
	var dailyForecasts []DailyForecast

	for i := 0; i < 3 && i < len(forecasts); i++ {
		forecast := forecasts[i]

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

		rainChances := []string{
			forecast.ChanceOfRain.T00_06,
			forecast.ChanceOfRain.T06_12,
			forecast.ChanceOfRain.T12_18,
			forecast.ChanceOfRain.T18_24,
		}

		dailyForecasts = append(dailyForecasts, DailyForecast{
			Date:        dateLabels[i],
			WeatherIcon: getWeatherIcon(forecast.Telop),
			Description: forecast.Telop,
			MaxTemp:     dailyMaxTemp,
			MinTemp:     dailyMinTemp,
			RainChance:  maxRainChancePercent(rainChances),
		})
	}

	return dailyForecasts
}

// maxRainChancePercent は降水確率4スロットから最大値を返す
func maxRainChancePercent(chances []string) string {
	maxRainChance := "0%"
	maxPercent := 0
	for _, rc := range chances {
		if rc == "" || rc == "-" {
			continue
		}
		percentStr := rc
		if len(rc) > 0 && rc[len(rc)-1] == '%' {
			percentStr = rc[:len(rc)-1]
		}
		if currentPercent, err := strconv.Atoi(percentStr); err == nil && currentPercent > maxPercent {
			maxPercent = currentPercent
			maxRainChance = rc
		}
	}
	return maxRainChance
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
	rss, err := fetchAndDecodeXML[NHKNewsRSS]("https://www3.nhk.or.jp/rss/news/cat0.xml")
	if err != nil {
		return nil, err
	}
	return convertRSSItemsToNews(rss.Channel.Items, MaxNewsItems), nil
}

func fetchEconomyNewsData() ([]NewsItem, error) {
	rss, err := fetchAndDecodeXML[NHKNewsRSS]("https://www3.nhk.or.jp/rss/news/cat5.xml")
	if err != nil {
		return nil, err
	}
	return convertRSSItemsToNews(rss.Channel.Items, MaxEconomyNewsItems), nil
}

// convertRSSItemsToNews はNHKのRSSアイテムをNewsItemスライスに変換する
func convertRSSItemsToNews(items []RSSItem, maxItems int) []NewsItem {
	if len(items) < maxItems {
		maxItems = len(items)
	}
	news := make([]NewsItem, 0, maxItems)
	for i := 0; i < maxItems; i++ {
		item := items[i]
		news = append(news, NewsItem{
			Title:       item.Title,
			Link:        item.Link,
			Description: item.Description,
			PubDate:     formatRSSDate(item.PubDate, "Mon, 02 Jan 2006 15:04:05 -0700"),
		})
	}
	return news
}

// fetchHatenaBookmarks はてなブックマークの人気エントリーを取得する
func fetchHatenaBookmarks() ([]HatenaEntry, error) {
	rss, err := fetchAndDecodeXML[HatenaBookmarkRSS]("https://b.hatena.ne.jp/hotentry/all.rss")
	if err != nil {
		return nil, err
	}
	return convertHatenaRSSToEntries(rss.Items, MaxHatenaItems), nil
}

// fetchKnowledgeHatenaBookmarks はてなブックマークの学びカテゴリの人気エントリーを取得する
func fetchKnowledgeHatenaBookmarks() ([]HatenaEntry, error) {
	rss, err := fetchAndDecodeXML[HatenaBookmarkRSS]("https://b.hatena.ne.jp/hotentry/knowledge.rss")
	if err != nil {
		return nil, err
	}
	return convertHatenaRSSToEntries(rss.Items, MaxHatenaItems), nil
}

// convertHatenaRSSToEntries ははてなブックマークのRSSアイテムをHatenaEntryスライスに変換する
func convertHatenaRSSToEntries(items []HatenaRSSItem, maxItems int) []HatenaEntry {
	if len(items) < maxItems {
		maxItems = len(items)
	}
	entries := make([]HatenaEntry, 0, maxItems)
	for i := 0; i < maxItems; i++ {
		item := items[i]
		category := ""
		if len(item.Subjects) > 0 {
			category = item.Subjects[0]
		}
		entries = append(entries, HatenaEntry{
			Title:        item.Title,
			Link:         item.Link,
			BookmarkLink: hatenaBookmarkURL(item.Link),
			Description:  item.Description,
			PubDate:      formatRSSDate(item.Date, "2006-01-02T15:04:05Z"),
			Category:     category,
		})
	}
	return entries
}

// filterDuplicates はexcludeItemsに含まれるキーと重複しないアイテムを最大maxItems件返す
func filterDuplicates[T any](items, excludeItems []T, getKey func(T) string, maxItems int) []T {
	excludeKeys := make(map[string]bool, len(excludeItems))
	for _, item := range excludeItems {
		excludeKeys[getKey(item)] = true
	}

	var filtered []T
	for _, item := range items {
		if !excludeKeys[getKey(item)] {
			filtered = append(filtered, item)
			if len(filtered) >= maxItems {
				break
			}
		}
	}
	return filtered
}

func filterDuplicateNews(economyNews []NewsItem, mainNews []NewsItem) []NewsItem {
	return filterDuplicates(economyNews, mainNews, func(n NewsItem) string { return n.Title }, MaxNewsItems)
}

func filterDuplicateHatenaEntries(knowledgeEntries []HatenaEntry, generalEntries []HatenaEntry) []HatenaEntry {
	return filterDuplicates(knowledgeEntries, generalEntries, func(e HatenaEntry) string { return e.Title }, MaxHatenaItems)
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