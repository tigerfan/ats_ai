package db

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var (
	client influxdb2.Client
)

func InitDB() error {
	client = influxdb2.NewClient("http://localhost:8086", "your-token")
	return nil
}

func WriteMeasurementData(deviceID, channelID int, voltages []float64) error {
	writeAPI := client.WriteAPIBlocking("your-org", "your-bucket")

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
	queryAPI := client.QueryAPI("your-org")

	query := fmt.Sprintf(`
		from(bucket:"your-bucket")
			|> range(start: -1h)
			|> filter(fn: (r) => r._measurement == "voltage" and r.device_id == "%d" and r.channel_id == "%d")
			|> sort(columns: ["_time"])
			|> limit(n: 1000)
	`, deviceID, channelID)

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
