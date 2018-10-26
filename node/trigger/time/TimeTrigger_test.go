package time

import (
	"github.com/cpucycle/astrotime"
	"github.com/kelvins/sunrisesunset"
	"testing"
	"time"
)

func TestAstroTime(t *testing.T) {
	var lat, long float64
	lat = 58.969976
	long = 5.733107
	sunriseTime := astrotime.NextSunrise(time.Now(), lat, -long)
	sunsetTime := astrotime.NextSunset(time.Now(), lat, -long)
	t.Log("1.Next Sunrise at ", sunriseTime.Format(TIME_FORMAT))
	t.Log("1.Next Sunset at ", sunsetTime.Format(TIME_FORMAT))

	p := sunrisesunset.Parameters{
		Latitude:  lat,
		Longitude: long,
		UtcOffset: 1.0,
		Date:      time.Now(),
	}

	// Calculate the sunrise and sunset times
	sunrise, sunset, _ := p.GetSunriseSunset()
	today := time.Now()
	sunrise = time.Date(today.Year(), today.Month(), today.Day(), sunrise.Hour(), sunrise.Minute(), 0, 0, time.Local)

	t.Log("2.Next Sunrise at ", sunrise.Format(TIME_FORMAT))
	t.Log("2.Next Sunset at ", sunset.Format(TIME_FORMAT))

}

func TestGetSunriseAndSunset(t *testing.T) {
	tmNode := Node{}
	tmNodeConfig := NodeConfig{Latitude: 58.969976, Longitude: 5.733107, TimeZone: 1}
	tmNode.config = tmNodeConfig
	sunrise1, sunset1, err := tmNode.getSunriseAndSunset(false)
	sunrise2, sunset2, err := tmNode.getSunriseAndSunset(true)
	if err != nil {
		t.Error(err)
	}

	t.Log("1. Sunrise at :", sunrise1.Format(TIME_FORMAT))
	t.Log("1. Sunset  at :", sunset1.Format(TIME_FORMAT))

	t.Log("2. Sunrise at :", sunrise2.Format(TIME_FORMAT))
	t.Log("2. Sunset  at :", sunset2.Format(TIME_FORMAT))
}

func TestNextEvent(t *testing.T) {
	tmNode := Node{}
	tmNodeConfig := NodeConfig{Latitude: 58.969976, Longitude: 5.733107, TimeZone: 1}
	tmNode.config = tmNodeConfig
	timeUntilEvent, eventType, err := tmNode.getTimeUntilNextEvent()
	if err != nil {
		t.Error(err)
	}
	t.Logf(" Next event is %s after %f hours ", eventType, timeUntilEvent.Hours())
	t.Log("Trigger will fire at ", time.Now().Add(timeUntilEvent).Format(TIME_FORMAT))

}
