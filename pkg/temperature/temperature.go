package temperature

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	client *http.Client

	apiKey string
}

func New(apiKey string) *Client {

	return &Client{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiKey: apiKey,
	}
}

func (c *Client) GetMoscowTemperature() (string, error) {
	resp, err := c.client.Get(
		"https://api.openweathermap.org/data/2.5/weather?lat=55.85488892&lon=37.47623062&units=metric&lang=ru&appid=" + c.apiKey)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("response status code ", resp.StatusCode)
		return "", fmt.Errorf("%s", "error response")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response Response

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	return fmt.Sprintf(`%s
Текущая температура: %d°
Ощущается как: %d°
Ветер: %d м/с
	`, strings.ToTitle(response.Weather[0].Description),
		int64(math.Round(response.Main.Temp)),
		int64(math.Round(response.Main.FeelsLike)),
		int64(math.Round(response.Wind.Speed)),
	), nil
}

type Response struct {
	Weather []Weather `json:"weather"`
	Main    Main      `json:"main"`
	Wind    Wind      `json:"wind"`
}

type Weather struct {
	Description string `json:"description"`
}
type Main struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
}

type Wind struct {
	Speed float64 `json:"speed"`
}
