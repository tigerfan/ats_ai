package db

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var (
	client influxdb2.Client
	config struct {
		InfluxDB struct {
			URL    string `json:"url"`
			Token  string `json:"token"`
			Org    string `json:"org"`
			Bucket string `json:"bucket"`
		} `json:"influxdb"`
	}
)

func InitDB() error {
	// Read config.json file
	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal JSON data into config struct
	if err := json.Unmarshal(configFile, &config); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	// Initialize InfluxDB client
	client = influxdb2.NewClient(config.InfluxDB.URL, config.InfluxDB.Token)
	return nil
}

func WriteMeasurementData(deviceID, channelID int, voltages []float64) error {
	writeAPI := client.WriteAPIBlocking(config.InfluxDB.Org, config.InfluxDB.Bucket)

	for i, voltage := range voltages {
		p := influxdb2.NewPoint(
			"voltage",
			map[string]string{
				"device_id":  fmt.Sprintf("%d", deviceID),
				"channel_id": fmt.Sprintf("%d", channelID),
			},
			map[string]interface{}{
				"value": voltage,
			},
			time.Now().Add(time.Duration(i)*time.Millisecond),
		)

		if err := writeAPI.WritePoint(context.Background(), p); err != nil {
			return err
		}
	}

	return nil
}

func GetHistoricalData(deviceID, channelID int) ([]float64, error) {
	queryAPI := client.QueryAPI(config.InfluxDB.Org)

	query := fmt.Sprintf(`
		from(bucket:"%s")
			|> range(start: -1h)
			|> filter(fn: (r) => r._measurement == "voltage" and r.device_id == "%d" and r.channel_id == "%d")
			|> sort(columns: ["_time"])
			|> limit(n: 1000)
	`, config.InfluxDB.Bucket, deviceID, channelID)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	var voltages []float64
	for result.Next() {
		voltages = append(voltages, result.Record().Values()["_value"].(float64))
	}

	return voltages, nil
}
