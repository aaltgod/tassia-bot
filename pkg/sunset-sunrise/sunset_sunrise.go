package sunsetsunrise

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetSunIntervals() (string, error) {

    resp, err := http.Get("https://voshod-solnca.ru/sun/%D0%BC%D0%BE%D1%81%D0%BA%D0%B2%D0%B0")
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

    span := doc.Find("div.position-relative.sun-today-table.with-adwords").
        Find("ul.today-list.list-group.list-unstyled.px-0").
        Find("li.today-list__item.align-items-center.d-flex").
        Find("p.today-list__item-container.w-100.mb-0.d-flex.flex-wrap").
        Find("span.today-list__item-value")
    sunrise := span.Eq(1)
    sunset := span.Eq(3)
    dayLength := span.Eq(9)

    result.WriteString("Восход: ")
    result.WriteString(sunrise.Text())
    result.WriteRune('\n')
    result.WriteString("Закат: ")
    result.WriteString(sunset.Text())
    result.WriteRune('\n')
    result.WriteString("Долгота дня: ")
    result.WriteString(dayLength.Text())

	return result.String(), nil
}
