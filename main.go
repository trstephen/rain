package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gizak/termui"
	"github.com/shawntoffel/darksky"
)

type location struct {
	latitude, longitude darksky.Measurement
}

type configuration struct {
	canned bool
	city   string
}

type forecastData struct {
	precipProbabilities []int
	summary             string
}

var (
	config configuration

	cityLocations = map[string]location{
		"seattle":  location{47.6038, -122.3301},
		"victoria": location{48.4263, -123.3538},
	}

	cannedData = forecastData{
		precipProbabilities: []int{
			0, 0, 0, 2, 2, 4, 8, 9, 13, 13,
			13, 13, 20, 20, 20, 45, 49, 55, 57, 61,
			63, 77, 57, 58, 60, 45, 45, 46, 43, 41,
			40, 39, 35, 33, 33, 31, 33, 37, 34, 35,
			36, 38, 36, 35, 31, 32, 29, 25, 24, 20,
			18, 9, 4, 0, 0, 2, 3, 0, 4, 0,
		},
		summary: "Heavy rain starting in 15 min.",
	}
)

func main() {
	flag.StringVar(&config.city, "city", "seattle", "City for the forecast")
	flag.BoolVar(&config.canned, "canned", false, "Use canned data")
	flag.Parse()

	if err := termui.Init(); err != nil {
		panic(err)
	}
	defer termui.Close()

	var data forecastData
	if config.canned {
		data = cannedData
	} else {
		data = getDarkskyForecast()
	}

	widgetWidth := 60
	headerHeight := 1
	cityPar := termui.NewPar(config.city)
	cityPar.Height = headerHeight
	cityPar.Width = widgetWidth
	cityPar.Border = false

	spl0 := termui.NewSparkline()
	spl0.Data = data.precipProbabilities
	spl0.LineColor = termui.ColorBlue
	spl0.Height = 4

	precipSLHeight := 7
	spls0 := termui.NewSparklines(spl0)
	spls0.Y = headerHeight
	spls0.Height = precipSLHeight
	spls0.Width = widgetWidth
	spls0.BorderLabel = "Precipitation probability"

	summaryHeight := 1
	summaryPar := termui.NewPar(data.summary)
	summaryPar.Height = summaryHeight
	summaryPar.Width = widgetWidth
	summaryPar.Border = false
	summaryPar.Y = headerHeight + precipSLHeight

	termui.Render(cityPar, spls0, summaryPar)

	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
	})

	termui.Loop()
}

func getDarkskyForecast() forecastData {
	client := darksky.New(os.Getenv("DARKSKY_API_TOKEN"))
	request := darksky.ForecastRequest{}
	request.Options = darksky.ForecastRequestOptions{
		Exclude: "hourly,daily,flags",
		Units:   "ca",
	}

	if l, found := cityLocations[config.city]; found {
		request.Latitude = l.latitude
		request.Longitude = l.longitude
	} else {
		panic("I don't know the lat/long for " + config.city)
	}

	forecast, err := client.Forecast(request)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Forecast for", config.city)
	fmt.Printf("%+v\n", forecast)

	if forecast.Alerts != nil {
		for _, alert := range forecast.Alerts {
			fmt.Printf("(%s) %s\n", alert.Severity, alert.Title)
		}
	}

	if forecast.Minutely == nil {
		fmt.Printf("\nNo minutely data for %s :(\n", config.city)
		return forecastData{
			precipProbabilities: nil,
			summary:             "No minutely data :(",
		}
	}

	precipProbabilities := make([]int, len(forecast.Minutely.Data))
	for i, datum := range forecast.Minutely.Data {
		forecastInstant := time.Unix(int64(datum.Time), 0)
		precipProbability := int(datum.PrecipProbability * 100)
		precipProbabilities[i] = precipProbability
		fmt.Printf("%s | %3d%%, %f\n", forecastInstant.Format("15:04"), precipProbability, datum.PrecipIntensity)
	}

	fmt.Println(forecast.Minutely.Summary)

	return forecastData{
		precipProbabilities: precipProbabilities,
		summary:             forecast.Minutely.Summary,
	}
}
