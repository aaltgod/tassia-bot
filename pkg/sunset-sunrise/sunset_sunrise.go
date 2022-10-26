package sunsetsunrise

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetSunIntervals() (string, error) {

	resp, err := http.Get("https://yandex.ru/pogoda/?lat=55.85489273&lon=37.47623444")
	if err != nil {
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

	var result strings.Builder

	card := doc.Find("div.sun-card.sun-card_theme_light.card")
	cardDiagram := card.Find("div.sun-card__diagram")
	cardInfo := card.Find("div.sun-card__info")

	sunrise := cardDiagram.Find("div.sun-card__sunrise-sunset-info.sun-card__sunrise-sunset-info_value_rise-time").After("span")
	sunset := cardDiagram.Find("div.sun-card__sunrise-sunset-info.sun-card__sunrise-sunset-info_value_set-time").After("span")
	dayLength := cardDiagram.Find("div.sun-card__day-duration-value")

	result.WriteString("Восход: ")
	result.WriteString(sunrise.Text())
	result.WriteRune('\n')
	result.WriteString("Закат: ")
	result.WriteString(sunset.Text())
	result.WriteRune('\n')
	result.WriteString("Долгота дня: ")
	result.WriteString(dayLength.Text())
	result.WriteRune('\n')

	additionalInfoItems := cardInfo.Find("div.sun-card__info-item").Map(func(i int, s *goquery.Selection) string {
		return s.Text()
	})

	for _, item := range additionalInfoItems {
		result.WriteRune('\n')
		result.WriteString(item)
	}

	return result.String(), nil
}
