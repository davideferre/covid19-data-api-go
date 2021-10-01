package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/joho/godotenv"
)

var HTTP_LISTEN string
var INFLUX_ADDRESS string

func main() {
	/*
		loading .env files in this order, if a variable is not set in `.env`,
		it's read from `.env.default`
	*/
	errEnv := godotenv.Load(".env")
	errEnvDefault := godotenv.Load(".env.default")
	if errEnvDefault != nil && errEnv != nil {
		log.Fatal("Error loading .env.default or .env file")
	}
	HTTP_LISTEN = os.Getenv("HTTP_LISTEN")
	INFLUX_ADDRESS = os.Getenv("INFLUX_ADDRESS")

	router := gin.Default()
	router.GET("/api/v2/data/nations", nations)
	router.Run(HTTP_LISTEN)
}

func nations(ctx *gin.Context) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: INFLUX_ADDRESS,
	})
	if err != nil {
		fmt.Println("Error creating InfluxDB Client: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	defer c.Close()

	q := client.NewQuery("SELECT * FROM nation ORDER BY time ASC", "dati", "")
	if response, err := c.Query(q); err == nil && response.Error() == nil {
		results := make(map[string][]map[string]string)
		for _, s := range response.Results[0].Series {
			for _, v := range s.Values {
				m := make(map[string]string)
				for k := range v {
					if v[k] != nil {
						m[s.Columns[k]] = fmt.Sprintf("%s", v[k])
					} else {
						m[s.Columns[k]] = "0"
					}
				}
				time := m["time"]
				results[time] = append(results[time], m)
			}
		}
		ctx.JSON(http.StatusOK, results)
	}
}
