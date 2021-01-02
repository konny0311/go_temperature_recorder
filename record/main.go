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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type weatherInfo interface {
	getInfo()
}

type Remo struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	NewestEvents struct {
		Te struct {
			Val       float64 `json:"val"`
			CreatedAt string  `json:"created_at"`
		}
	} `json:"newest_events"`
}

func (remo *Remo) getInfo() {
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

	var results []Remo
	err = json.Unmarshal(body, &results)
	if err != nil {
		log.Fatal(err)
	}

	remo.Id = results[0].Id
	remo.Name = results[0].Name
	remo.NewestEvents = results[0].NewestEvents
}

type OpenWeather struct {
	Main struct {
		// Temperature in Celcius
		Temp     float64 `json:"temp"`
		TempMax  float64 `json:"temp_max"`
		TempMin  float64 `json:"temp_min"`
		Humidity float64 `json:"humidity"`
		Pressure float64 `json:"pressure"`
	}
	Clouds struct {
		All float64 `json:"all"`
	}
}

func (ow *OpenWeather) getInfo() {
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=35.631562&lon=139.644311&units=metric&appid=%s", os.Getenv("OW_KEY"))
	req, _ := http.NewRequest("GET", url, nil)

	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)

	var result OpenWeather
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatal(err)
	}

	ow.Main = result.Main
	ow.Clouds = result.Clouds
}

type TempRecord struct {
	Remo      Remo
	OW        OpenWeather
	CreatedAt string
}

func (tr TempRecord) putItemDynamodb() error {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)

	av, err := dynamodbattribute.MarshalMap(tr)
	if err != nil {
		fmt.Println("Got error marshalling new movie item:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	tableName := "TemperatureRecord"
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}
	_, err = svc.PutItem(input)
	return err
}

type Response events.APIGatewayProxyResponse

func Handler() (Response, error) {
	fmt.Println("Temperature recording test.")
	tempRecord := TempRecord{}
	tempRecord.Remo.getInfo()
	fmt.Println("remo done")
	tempRecord.OW.getInfo()
	fmt.Println("ow done")
	tempRecord.CreatedAt = tempRecord.Remo.NewestEvents.Te.CreatedAt
	fmt.Println(tempRecord)
	fmt.Println("start putting to DB")
	err := tempRecord.putItemDynamodb()
	fmt.Println("finish putting DB")

	if err != nil {
		return Response{StatusCode: 500}, err
	}
	return Response{StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}
