package glance

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	_ "time/tzdata"
)

var weatherWidgetTemplate = mustParseTemplate("weather.html", "widget-base.html")

// QWeather API configuration
// API Key should be set via QWEATHER_API_KEY environment variable
const (
	qweatherGeoAPI    = "https://geoapi.qweather.com/v2/city/lookup"
	qweatherNowAPI    = "https://devapi.qweather.com/v7/weather/now"
	qweatherHourlyAPI = "https://devapi.qweather.com/v7/weather/24h"
)

func getQWeatherAPIKey() string {
	if key := os.Getenv("QWEATHER_API_KEY"); key != "" {
		return key
	}
	// Fallback to default key for backward compatibility
	return "4d4123daeba9471c81cde83b411a05c4"
}

type weatherWidget struct {
	widgetBase   `yaml:",inline"`
	Location     string       `yaml:"location"`
	ShowAreaName bool         `yaml:"show-area-name"`
	HideLocation bool         `yaml:"hide-location"`
	HourFormat   string       `yaml:"hour-format"`
	Units        string       `yaml:"units"`
	Place        *qweatherLocation `yaml:"-"`
	Weather      *weather           `yaml:"-"`
	TimeLabels   [12]string         `yaml:"-"`
}

// Chinese time labels
var timeLabels12h = [12]string{"凌晨2点", "凌晨4点", "早上6点", "上午8点", "上午10点", "中午12点", "下午2点", "下午4点", "傍晚6点", "晚上8点", "晚上10点", "凌晨12点"}
var timeLabels24h = [12]string{"02:00", "04:00", "06:00", "08:00", "10:00", "12:00", "14:00", "16:00", "18:00", "20:00", "22:00", "00:00"}

func (widget *weatherWidget) initialize() error {
	// Changed title to Chinese
	widget.withTitle("天气").withCacheOnTheHour()

	if widget.Location == "" {
		return fmt.Errorf("location is required")
	}

	// Default to 24h format for Chinese users
	if widget.HourFormat == "" || widget.HourFormat == "24h" {
		widget.TimeLabels = timeLabels24h
	} else if widget.HourFormat == "12h" {
		widget.TimeLabels = timeLabels12h
	} else {
		return errors.New("hour-format must be either 12h or 24h")
	}

	// QWeather only supports metric (Celsius)
	widget.Units = "metric"

	return nil
}

func (widget *weatherWidget) update(ctx context.Context) {
	if widget.Place == nil {
		place, err := fetchQWeatherLocation(widget.Location)
		if err != nil {
			widget.withError(err).scheduleEarlyUpdate()
			return
		}

		widget.Place = place
	}

	weather, err := fetchQWeatherForLocation(widget.Place)

	if !widget.canContinueUpdateAfterHandlingErr(err) {
		return
	}

	widget.Weather = weather
}

func (widget *weatherWidget) Render() template.HTML {
	return widget.renderTemplate(widget, weatherWidgetTemplate)
}

type weather struct {
	Temperature         int
	ApparentTemperature int
	WeatherCode         int
	CurrentColumn       int
	SunriseColumn       int
	SunsetColumn        int
	Columns             []weatherColumn
}

func (w *weather) WeatherCodeAsString() string {
	if weatherCode, ok := weatherCodeTable[w.WeatherCode]; ok {
		return weatherCode
	}

	return ""
}

// QWeather API response structures
type qweatherLocationResponse struct {
	Code     string `json:"code"`
	Location []struct {
		Name      string `json:"name"`
		ID        string `json:"id"`
		Lat       string `json:"lat"`
		Lon       string `json:"lon"`
		Adm2      string `json:"adm2"`  // 区县
		Adm1      string `json:"adm1"`  // 省市
		Country   string `json:"country"`
		Timezone  string `json:"tz"`
	} `json:"location"`
}

type qweatherLocation struct {
	ID        string
	Name      string
	Adm2      string
	Adm1      string
	Latitude  float64
	Longitude float64
	Timezone  string
}

type qweatherNowResponse struct {
	Code       string `json:"code"`
	Now        struct {
		Temp        string `json:"temp"`
		FeelsLike   string `json:"feelsLike"`
		Icon        string `json:"icon"`
		Text        string `json:"text"`
	} `json:"now"`
}

type qweatherHourlyResponse struct {
	Code     string `json:"code"`
	Hourly   []struct {
		Temp string `json:"temp"`
		Icon string `json:"icon"`
	} `json:"hourly"`
}

type weatherColumn struct {
	Temperature      int
	Scale            float64
	HasPrecipitation bool
}

// fetchQWeatherLocation looks up location by name using QWeather API
func fetchQWeatherLocation(location string) (*qweatherLocation, error) {
	requestUrl := fmt.Sprintf("%s?location=%s&key=%s", qweatherGeoAPI, url.QueryEscape(location), qweatherAPIKey)
	request, _ := http.NewRequest("GET", requestUrl, nil)
	responseJson, err := decodeJsonFromRequest[qweatherLocationResponse](defaultHTTPClient, request)
	if err != nil {
		return nil, fmt.Errorf("fetching location data: %v", err)
	}

	if responseJson.Code != "200" || len(responseJson.Location) == 0 {
		return nil, fmt.Errorf("no locations found for %s", location)
	}

	loc := responseJson.Location[0]
	lat := 0.0
	lon := 0.0
	fmt.Sscanf(loc.Lat, "%f", &lat)
	fmt.Sscanf(loc.Lon, "%f", &lon)

	return &qweatherLocation{
		ID:        loc.ID,
		Name:      loc.Name,
		Adm2:      loc.Adm2,
		Adm1:      loc.Adm1,
		Latitude:  lat,
		Longitude: lon,
		Timezone:  loc.Timezone,
	}, nil
}

// fetchQWeatherForLocation fetches weather data from QWeather API
func fetchQWeatherForLocation(loc *qweatherLocation) (*weather, error) {
	// Fetch current weather
	nowUrl := fmt.Sprintf("%s?location=%s&key=%s", qweatherNowAPI, loc.ID, qweatherAPIKey)
	nowRequest, _ := http.NewRequest("GET", nowUrl, nil)
	nowResponse, err := decodeJsonFromRequest[qweatherNowResponse](defaultHTTPClient, nowRequest)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errNoContent, err)
	}

	if nowResponse.Code != "200" {
		return nil, fmt.Errorf("%w: API error code %s", errNoContent, nowResponse.Code)
	}

	// Fetch hourly forecast
	hourlyUrl := fmt.Sprintf("%s?location=%s&key=%s", qweatherHourlyAPI, loc.ID, getQWeatherAPIKey())
	hourlyRequest, _ := http.NewRequest("GET", hourlyUrl, nil)
	hourlyResponse, err := decodeJsonFromRequest[qweatherHourlyResponse](defaultHTTPClient, hourlyRequest)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errNoContent, err)
	}

	// Parse current temperature
	currentTemp := 0
	fmt.Sscanf(nowResponse.Now.Temp, "%d", &currentTemp)
	
	apparentTemp := 0
	fmt.Sscanf(nowResponse.Now.FeelsLike, "%d", &apparentTemp)

	// Parse hourly data for the chart (next 12 hours, every 2 hours)
	columns := make([]weatherColumn, 0, 12)
	currentHour := time.Now().Hour()
	
	for i := 0; i < 12 && i < len(hourlyResponse.Hourly); i += 2 {
		hourIndex := i
		if hourIndex >= len(hourlyResponse.Hourly) {
			break
		}
		
		temp := 0
		fmt.Sscanf(hourlyResponse.Hourly[hourIndex].Temp, "%d", &temp)
		
		columns = append(columns, weatherColumn{
			Temperature:      temp,
			Scale:            0, // Will be calculated later
			HasPrecipitation: false, // QWeather doesn't provide precipitation probability in free tier
		})
	}

	// Calculate scale for visualization
	if len(columns) > 0 {
		minTemp, maxTemp := columns[0].Temperature, columns[0].Temperature
		for _, col := range columns {
			if col.Temperature < minTemp {
				minTemp = col.Temperature
			}
			if col.Temperature > maxTemp {
				maxTemp = col.Temperature
			}
		}
		
		tempRange := maxTemp - minTemp
		if tempRange == 0 {
			tempRange = 1
		}
		
		for i := range columns {
			columns[i].Scale = float64(columns[i].Temperature-minTemp) / float64(tempRange)
		}
	}

	// Map QWeather icon to weather code
	weatherCode := mapQWeatherIconToCode(nowResponse.Now.Icon)

	return &weather{
		Temperature:         currentTemp,
		ApparentTemperature: apparentTemp,
		WeatherCode:         weatherCode,
		CurrentColumn:       currentHour / 2,
		SunriseColumn:       3,  // Approximate sunrise (6am)
		SunsetColumn:        9,  // Approximate sunset (6pm)
		Columns:             columns,
	}, nil
}

// mapQWeatherIconToCode maps QWeather icon codes to internal weather codes
func mapQWeatherIconToCode(icon string) int {
	// QWeather icon codes: https://dev.qweather.com/docs/resource/icons/
	switch icon {
	case "100", "150": // Sunny / Clear night
		return 0
	case "101", "151": // Mostly clear
		return 1
	case "102", "152": // Partly cloudy
		return 2
	case "103", "153": // Cloudy
		return 3
	case "104", "154": // Overcast
		return 3
	case "300", "301", "302", "303", "304": // Rain showers
		return 80
	case "305", "306", "307", "308", "309", "310", "311", "312", "313", "314", "315": // Rain
		return 61
	case "400", "401", "402", "403", "404", "405", "406", "407": // Snow
		return 71
	case "500", "501", "502", "503", "504", "505", "506", "507", "508": // Fog/Haze
		return 45
	default:
		return 0
	}
}

// Chinese weather condition descriptions
var weatherCodeTable = map[int]string{
	0:  "晴朗",
	1:  "晴间多云",
	2:  "多云",
	3:  "阴天",
	45: "雾",
	48: "雾凇",
	51: "毛毛雨",
	53: "小雨",
	55: "中雨",
	56: "冻雨",
	57: "冻雨",
	61: "小雨",
	63: "中雨",
	65: "大雨",
	66: "冻雨",
	67: "冻雨",
	71: "小雪",
	73: "中雪",
	75: "大雪",
	77: "雪粒",
	80: "阵雨",
	81: "中雨",
	82: "暴雨",
	85: "阵雪",
	86: "阵雪",
	95: "雷雨",
	96: "雷雨",
	99: "雷雨",
}
