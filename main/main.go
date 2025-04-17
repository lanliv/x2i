package main

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"log"
)

func main() {
	influxUrl := "http://localhost:8086"
	influxToken := "cO5GJoY2gv2vF6n6iDIrkDWEircxNmShqvp0q1j7sKVJVvFPe5jx71R2u3sxLA0YaYMEIWcYqjlHn88AP5WQ8g=="
	org := "tui"
	bucket := "gatling"

	client := influxdb2.NewClient(influxUrl, influxToken)
	defer client.Close()

	query := fmt.Sprintf(`
	from(bucket: "%s")
	|> range(start: -1d) 
	|> filter(fn: (r) => r._measurement == "requests")
	`, bucket)

	queryAPI := client.QueryAPI(org)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		log.Fatalf("query error: %v", err)
	}

	for result.Next() {
		r := result.Record()
		fmt.Printf("%v, %v, %v, %v, %v, %v\n",
			r.ValueByKey("_time"),
			r.ValueByKey("name"),
			r.ValueByKey("_value"),
			r.ValueByKey("result"),
			r.ValueByKey("simulation"),
			r.ValueByKey("errorMessage"),
		)
	}

	if result.Err() != nil {
		log.Fatalf("query parsing error: %v", result.Err())
	}
}
