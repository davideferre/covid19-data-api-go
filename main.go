package main

import (
	"encoding/json"
	"fmt"

	client "github.com/influxdata/influxdb1-client/v2"
)

func main() {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://localhost:8086",
	})
	if err != nil {
		fmt.Println("Error creating InfluxDB Client: ", err.Error())
	}
	defer c.Close()

	q := client.NewQuery("SELECT * FROM nation ORDER BY time ASC", "dati", "")
	if response, err := c.Query(q); err == nil && response.Error() == nil {
		results := make(map[int][]map[string]string)
		for _, s := range response.Results[0].Series {
			// bs, _ := json.Marshal(s)
			// fmt.Println(string(bs))
			for j, v := range s.Values {
				m := make(map[string]string)
				for k := range v {
					if v[k] != nil {
						m[s.Columns[k]] = fmt.Sprintf("%s", v[k])
					} else {
						m[s.Columns[k]] = "0"
					}
				}
				results[j] = append(results[j], m)
			}
		}
		bs, _ := json.Marshal(results)
		fmt.Println(string(bs))
	}
}
