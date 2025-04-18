package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2"
	"log"
	"net/http"
	"strings"
)

// Struct for point with named fields
type DatadogPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type DatadogMetric struct {
	Metric    string         `json:"metric"`
	Type      int            `json:"type"` // 0 = gauge
	Points    []DatadogPoint `json:"points"`
	Resources []Resource     `json:"resources,omitempty"`
	Tags      []string       `json:"tags,omitempty"`
}

type Resource struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func main() {
	// InfluxDB config
	influxUrl := "http://localhost:8086"
	influxToken := "cO5GJoY2gv2vF6n6iDIrkDWEircxNmShqvp0q1j7sKVJVvFPe5jx71R2u3sxLA0YaYMEIWcYqjlHn88AP5WQ8g=="
	org := "tui"
	bucket := "gatling"

	// Datadog config
	datadogAPIKey := "26ee9a57a7478f5bbf404345440ece6d"
	datadogURL := "https://api.datadoghq.eu/api/v2/series"

	client := influxdb2.NewClient(influxUrl, influxToken)
	defer client.Close()

	query := fmt.Sprintf(`
	from(bucket: "%s")
	|> range(start: -10m)
	|> filter(fn: (r) => r._measurement == "requests")
	`, bucket)

	queryAPI := client.QueryAPI(org)
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		log.Fatalf("Query error: %v", err)
	}

	var series []DatadogMetric

	for result.Next() {
		r := result.Record()

		measurement := "gatling." + r.Measurement()
		name := safeString(r.ValueByKey("name"))
		simulation := safeString(r.ValueByKey("simulation"))
		resultVal := safeString(r.ValueByKey("result"))
		errorMessage := escapeQuotes(safeString(r.ValueByKey("errorMessage")))
		timestamp := r.Time().Unix()

		// Assert value to float64
		value := parseFloat(r.Value())
		if value == nil {
			log.Printf("Failed to parse value: %v", r.Value())
			continue
		}

		metric := DatadogMetric{
			Metric: measurement,
			Type:   0,
			Points: []DatadogPoint{
				{Timestamp: timestamp, Value: *value},
			},
			Resources: []Resource{
				{Name: "https://tui-observability.datadoghq.eu/", Type: "host"},
			},
			Tags: []string{
				"name=" + name,
				"simulation=" + simulation,
				"result=" + resultVal,
			},
		}

		series = append(series, metric)

		// Log raw input line
		fmt.Printf("Line: %s,name=%s,simulation=%s,result=%s,errorMessage=%s,_value=%.2f %d\n",
			measurement, name, simulation, resultVal, errorMessage, *value, timestamp)
	}

	if result.Err() != nil {
		log.Fatalf("Query parse error: %v", result.Err())
	}

	payload := map[string]interface{}{
		"series": series,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("JSON marshal error: %v", err)
	}

	fmt.Printf("Payload to Datadog:\n%s\n", body)

	req, err := http.NewRequest("POST", datadogURL, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Request build error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("DD-API-KEY", datadogAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Send error: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Sent to Datadog. Status: %s", resp.Status)
}

// utils
func safeString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func escapeQuotes(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

func parseFloat(val interface{}) *float64 {
	switch v := val.(type) {
	case float64:
		return &v
	case int64:
		f := float64(v)
		return &f
	default:
		return nil
	}
}
