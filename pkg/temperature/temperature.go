package temperature

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func GetMoscowTemperature() (string, error) {

	resp, err := http.Get("https://yandex.ru/pogoda/?lat=55.85489273&lon=37.47623444")
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("response status code ", resp.StatusCode)
		return "", fmt.Errorf("%s", "error response")
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}

	result := doc.Find("div.link__feelings.fact__feelings").Find(
		"div.link__condition.day-anchor").Text()
	result += "\nТекущая температура: " + doc.Find("div.temp.fact__temp.fact__temp_size_s").Find(
		"span.temp__value.temp__value_with-unit").Text() + "°"
	result += "\nОщущается как: " + doc.Find("div.link__feelings.fact__feelings").Find(
		"span.temp__value.temp__value_with-unit").Text() + "°"

	return result, nil
}
