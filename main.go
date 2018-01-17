package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/shawntoffel/darksky"
)

type location struct {
	latitude, longitude darksky.Measurement
}

type configuration struct {
	city string
}

var (
	config configuration

	cityLocations = map[string]location{
		"seattle":  location{47.6038, -122.3301},
		"victoria": location{48.4263, -123.3538},
	}
)

func main() {
	flag.StringVar(&config.city, "city", "seattle", "City for the forecast")
	flag.Parse()

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

	if forecast.Minutely == nil {
		fmt.Printf("\nNo minutely data for %s :(\n", config.city)
	} else {
		for _, datum := range forecast.Minutely.Data {
			forecastInstant := time.Unix(int64(datum.Time), 0)
			fmt.Printf("%s | %3.0f%%, %f\n", forecastInstant.Format("15:04"), datum.PrecipProbability*100, datum.PrecipIntensity)
		}

		fmt.Println(forecast.Minutely.Summary)
	}

	if forecast.Alerts != nil {
		for _, alert := range forecast.Alerts {
			fmt.Printf("(%s) %s\n", alert.Severity, alert.Title)
		}
	}
}
