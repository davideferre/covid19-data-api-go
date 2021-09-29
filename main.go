package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	client "github.com/influxdata/influxdb1-client/v2"
)

func main() {
	router := gin.Default()
	router.GET("/api/v2/data/nations", nations)
	router.Run(":8080")
}

func nations(ctx *gin.Context) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://localhost:8086",
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
