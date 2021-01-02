package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Remo struct {
	Name         string `json:"name"`
	NewestEvents struct {
		Te struct {
			Val float64 `json:"val"`
		}
	} `json:"newest_events"`
}

type OpenWeather struct {
	Main struct {
		// Temperature in Celcius
		Temp     float64 `json:"temp"`
		TempMax  float64 `json:"temp_max"`
		TempMin  float64 `json:"temp_min"`
		Humidity float64 `json:"humidity"`
	}
}

type Response events.APIGatewayProxyResponse

func Handler() (Response, error) {
	fmt.Println("Temperature recording test.")
	fmt.Println("Room temperature is ", getTemp())
	return Response{StatusCode: 200}, nil
}

func getTemp() float64 {
	url := "https://api.nature.global/1/devices"
	req, _ := http.NewRequest("GET", url, nil)
	token := fmt.Sprintf("Bearer %s", os.Getenv("REMO_TOKEN"))
	req.Header.Set("Authorization", token)

	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)

	var remos []Remo
	err = json.Unmarshal(body, &remos)
	if err != nil {
		log.Fatal(err)
	}

	return remos[0].NewestEvents.Te.Val
}

func main() {
	lambda.Start(Handler)
}
