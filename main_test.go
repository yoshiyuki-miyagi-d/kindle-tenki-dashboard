package main

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"testing"
	"time"
)

// parseTemperature のテスト
func TestParseTemperature(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		hasError bool
	}{
		{
			name:     "正常な数値",
			input:    "25",
			expected: 25,
			hasError: false,
		},
		{
			name:     "負の数値",
			input:    "-5",
			expected: -5,
			hasError: false,
		},
		{
			name:     "ゼロ",
			input:    "0",
			expected: 0,
			hasError: false,
		},
		{
			name:     "空文字列",
			input:    "",
			expected: 0,
			hasError: true,
		},
		{
			name:     "null文字列",
			input:    "null",
			expected: 0,
			hasError: true,
		},
		{
			name:     "不正な文字列",
			input:    "abc",
			expected: 0,
			hasError: true,
		},
		{
			name:     "小数点付き",
			input:    "25.5",
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTemperature(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("期待: エラー, 実際: nil")
				}
			} else {
				if err != nil {
					t.Errorf("期待: エラーなし, 実際: %v", err)
				}
				if result != tt.expected {
					t.Errorf("期待: %d, 実際: %d", tt.expected, result)
				}
			}
		})
	}
}

// getEnv のテスト
func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		expected     string
		setEnv       bool
	}{
		{
			name:         "環境変数が設定されている場合",
			envKey:       "TEST_KEY_1",
			envValue:     "test_value",
			defaultValue: "default",
			expected:     "test_value",
			setEnv:       true,
		},
		{
			name:         "環境変数が設定されていない場合",
			envKey:       "TEST_KEY_2",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
			setEnv:       false,
		},
		{
			name:         "環境変数が空文字列の場合",
			envKey:       "TEST_KEY_3",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
			setEnv:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト前にクリーンアップ
			os.Unsetenv(tt.envKey)

			if tt.setEnv {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnv(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("期待: %s, 実際: %s", tt.expected, result)
			}
		})
	}
}

// processWeatherData のテスト
func TestProcessWeatherData(t *testing.T) {
	tests := []struct {
		name     string
		response TsukumijimaWeatherResponse
		validate func(*testing.T, *WeatherData)
	}{
		{
			name: "正常なデータ処理",
			response: TsukumijimaWeatherResponse{
				PublicTime: "2025-10-02T11:00:00+09:00",
				Location: struct {
					Area       string `json:"area"`
					Prefecture string `json:"prefecture"`
					District   string `json:"district"`
					City       string `json:"city"`
				}{
					City: "東京",
				},
				Forecasts: []struct {
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
				}{
					{
						Date:      "2025-10-02",
						DateLabel: "今日",
						Telop:     "晴れ",
						Temperature: struct {
							Min struct {
								Celsius string `json:"celsius"`
							} `json:"min"`
							Max struct {
								Celsius string `json:"celsius"`
							} `json:"max"`
						}{
							Min: struct {
								Celsius string `json:"celsius"`
							}{Celsius: "18"},
							Max: struct {
								Celsius string `json:"celsius"`
							}{Celsius: "28"},
						},
					},
					{
						Date:      "2025-10-03",
						DateLabel: "明日",
						Telop:     "曇り",
						Temperature: struct {
							Min struct {
								Celsius string `json:"celsius"`
							} `json:"min"`
							Max struct {
								Celsius string `json:"celsius"`
							} `json:"max"`
						}{
							Min: struct {
								Celsius string `json:"celsius"`
							}{Celsius: "19"},
							Max: struct {
								Celsius string `json:"celsius"`
							}{Celsius: "25"},
						},
					},
				},
			},
			validate: func(t *testing.T, data *WeatherData) {
				if data.Location != "東京" {
					t.Errorf("Location: 期待=東京, 実際=%s", data.Location)
				}
				if data.Temperature != 28 {
					t.Errorf("Temperature: 期待=28, 実際=%d", data.Temperature)
				}
				if data.FeelsLike != 28 {
					t.Errorf("FeelsLike: 期待=28, 実際=%d", data.FeelsLike)
				}
				if data.Description != "晴れ" {
					t.Errorf("Description: 期待=晴れ, 実際=%s", data.Description)
				}
				if len(data.HourlyForecast) == 0 {
					t.Error("HourlyForecast が空です")
				}
			},
		},
		{
			name: "気温データがnullの場合",
			response: TsukumijimaWeatherResponse{
				Location: struct {
					Area       string `json:"area"`
					Prefecture string `json:"prefecture"`
					District   string `json:"district"`
					City       string `json:"city"`
				}{
					City: "大阪",
				},
				Forecasts: []struct {
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
				}{
					{
						Date:      "2025-10-02",
						DateLabel: "今日",
						Telop:     "曇り",
						Temperature: struct {
							Min struct {
								Celsius string `json:"celsius"`
							} `json:"min"`
							Max struct {
								Celsius string `json:"celsius"`
							} `json:"max"`
						}{
							Min: struct {
								Celsius string `json:"celsius"`
							}{Celsius: ""},
							Max: struct {
								Celsius string `json:"celsius"`
							}{Celsius: ""},
						},
					},
					{
						Date:      "2025-10-03",
						DateLabel: "明日",
						Telop:     "晴れ",
						Temperature: struct {
							Min struct {
								Celsius string `json:"celsius"`
							} `json:"min"`
							Max struct {
								Celsius string `json:"celsius"`
							} `json:"max"`
						}{
							Min: struct {
								Celsius string `json:"celsius"`
							}{Celsius: "20"},
							Max: struct {
								Celsius string `json:"celsius"`
							}{Celsius: "30"},
						},
					},
				},
			},
			validate: func(t *testing.T, data *WeatherData) {
				if data.Location != "大阪" {
					t.Errorf("Location: 期待=大阪, 実際=%s", data.Location)
				}
				// 今日のデータがnullの場合、明日の気温を使用
				if data.Temperature != 30 {
					t.Errorf("Temperature: 期待=30, 実際=%d", data.Temperature)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processWeatherData(tt.response)

			if result == nil {
				t.Fatal("processWeatherData が nil を返しました")
			}

			tt.validate(t, result)
		})
	}
}

// getSampleData のテスト
func TestGetSampleData(t *testing.T) {
	data, err := getSampleData()

	if err != nil {
		t.Fatalf("getSampleData がエラーを返しました: %v", err)
	}

	if data == nil {
		t.Fatal("getSampleData が nil を返しました")
	}

	if data.Location != "東京" {
		t.Errorf("Location: 期待=東京, 実際=%s", data.Location)
	}

	if data.Temperature != 22 {
		t.Errorf("Temperature: 期待=22, 実際=%d", data.Temperature)
	}

	if len(data.HourlyForecast) == 0 {
		t.Error("HourlyForecast が空です")
	}

	if len(data.News) == 0 {
		t.Error("News が空です")
	}

	// UpdateTime が正しいフォーマットかチェック
	_, err = time.Parse("2006/01/02 15:04", data.UpdateTime)
	if err != nil {
		t.Errorf("UpdateTime のフォーマットが不正: %s, エラー: %v", data.UpdateTime, err)
	}
}

// getSampleNews のテスト
func TestGetSampleNews(t *testing.T) {
	news := getSampleNews()

	if len(news) == 0 {
		t.Fatal("getSampleNews が空の配列を返しました")
	}

	for i, item := range news {
		if item.Title == "" {
			t.Errorf("News[%d].Title が空です", i)
		}
		if item.Link == "" {
			t.Errorf("News[%d].Link が空です", i)
		}
		if item.Description == "" {
			t.Errorf("News[%d].Description が空です", i)
		}
		if item.PubDate == "" {
			t.Errorf("News[%d].PubDate が空です", i)
		}
	}
}

// fetchWeatherData のモックテスト (統合テスト)
func TestFetchWeatherDataIntegration(t *testing.T) {
	t.Run("正常なAPIレスポンスの処理", func(t *testing.T) {
		mockResponse := `{
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
					"temperature": {
						"min": {"celsius": "18"},
						"max": {"celsius": "28"}
					}
				},
				{
					"date": "2025-10-03",
					"dateLabel": "明日",
					"telop": "曇り",
					"temperature": {
						"min": {"celsius": "19"},
						"max": {"celsius": "25"}
					}
				}
			]
		}`

		var weatherResponse TsukumijimaWeatherResponse
		err := json.Unmarshal([]byte(mockResponse), &weatherResponse)
		if err != nil {
			t.Fatalf("モックデータのUnmarshalに失敗: %v", err)
		}

		data := processWeatherData(weatherResponse)
		if data == nil {
			t.Fatal("data が nil です")
		}
		if data.Location != "東京" {
			t.Errorf("Location: 期待=東京, 実際=%s", data.Location)
		}
		if data.Temperature != 28 {
			t.Errorf("Temperature: 期待=28, 実際=%d", data.Temperature)
		}
	})

	t.Run("APIエラー時のフォールバック", func(t *testing.T) {
		// getSampleData() が正常に動作することを確認
		data, err := getSampleData()
		if err != nil {
			t.Fatalf("getSampleData がエラーを返しました: %v", err)
		}
		if data == nil {
			t.Fatal("data が nil です")
		}
		if data.Location != "東京" {
			t.Errorf("サンプルデータの Location が期待と異なります")
		}
	})
}

// fetchNewsData のモックテスト (統合テスト)
func TestFetchNewsDataIntegration(t *testing.T) {
	t.Run("正常なRSSレスポンスの処理", func(t *testing.T) {
		mockResponse := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>NHKニュース</title>
    <link>https://www.nhk.or.jp/news/</link>
    <description>NHKニュース</description>
    <item>
      <title>テストニュース1</title>
      <link>http://example.com/news1</link>
      <description>テスト説明1</description>
      <pubDate>Wed, 02 Oct 2025 12:00:00 +0900</pubDate>
    </item>
    <item>
      <title>テストニュース2</title>
      <link>http://example.com/news2</link>
      <description>テスト説明2</description>
      <pubDate>Wed, 02 Oct 2025 11:00:00 +0900</pubDate>
    </item>
  </channel>
</rss>`

		var rss NHKNewsRSS
		err := xml.Unmarshal([]byte(mockResponse), &rss)
		if err != nil {
			t.Fatalf("モックデータのUnmarshalに失敗: %v", err)
		}

		if len(rss.Channel.Items) != 2 {
			t.Errorf("ニュース数: 期待=2, 実際=%d", len(rss.Channel.Items))
		}

		// 日付フォーマットのテスト
		item := rss.Channel.Items[0]
		pubTime, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", item.PubDate)
		if err != nil {
			t.Errorf("日付のパースに失敗: %v", err)
		}
		formattedDate := pubTime.Format("01/02 15:04")
		if formattedDate != "10/02 12:00" {
			t.Errorf("フォーマット後の日付: 期待=10/02 12:00, 実際=%s", formattedDate)
		}
	})

	t.Run("不正なXMLのエラーハンドリング", func(t *testing.T) {
		invalidXML := `<invalid xml`

		var rss NHKNewsRSS
		err := xml.Unmarshal([]byte(invalidXML), &rss)
		if err == nil {
			t.Error("エラーが期待されましたが nil でした")
		}
	})

	t.Run("サンプルニュースのフォールバック", func(t *testing.T) {
		// getSampleNews() が正常に動作することを確認
		news := getSampleNews()
		if len(news) == 0 {
			t.Fatal("サンプルニュースが空です")
		}
		if news[0].Title == "" {
			t.Error("サンプルニュースのTitleが空です")
		}
	})
}