package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/joho/godotenv"
)

var HTTP_LISTEN string
var INFLUX_ADDRESS string
var INFLUX_DB_NAME string
var RELEASE_MODE bool

var influxClient client.Client

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
	INFLUX_DB_NAME = os.Getenv("INFLUX_DB_NAME")
	RELEASE_MODE, _ = strconv.ParseBool(os.Getenv("RELEASE_MODE"))

	/*
		initializing influx connection
	*/
	var err error
	influxClient, err = client.NewHTTPClient(client.HTTPConfig{
		Addr: INFLUX_ADDRESS,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	defer influxClient.Close()

	if RELEASE_MODE {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(cors.Default())
	router.GET("/api/v2/data/nations", nations)
	router.Run(HTTP_LISTEN)
}

func nations(ctx *gin.Context) {
	get_results("SELECT * FROM nation", "nations", ctx)
}

func get_results(query string, results_name string, ctx *gin.Context) {
	q := client.NewQuery(query, INFLUX_DB_NAME, "")
	response, err := influxClient.Query(q)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
	} else if response.Error() != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Response error"})
	} else {
		results := []gin.H{}
		for _, s := range response.Results[0].Series {
			for _, v := range s.Values {
				m := make(gin.H)
				for k := range v {
					if v[k] != nil {
						m[s.Columns[k]] = fmt.Sprintf("%s", v[k])
					} else {
						m[s.Columns[k]] = "0"
					}
					if s.Columns[k] == "time" {
						m["id"] = fmt.Sprintf("%s", v[k])
					}
				}
				results = append(results, m)
			}
		}
		ctx.JSON(http.StatusOK, gin.H{results_name: results})
	}
}
