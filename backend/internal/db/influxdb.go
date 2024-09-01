package db

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Printf("Failed to read config file: %v", err)
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(configFile, &config); err != nil {
		log.Printf("Failed to unmarshal config file: %v", err)
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	client = influxdb2.NewClient(config.InfluxDB.URL, config.InfluxDB.Token)
	log.Println("InfluxDB client initialized successfully")
	return nil
}

func WriteMeasurementHistory(startTime, endTime time.Time, status string, deviceCount, channelCount int) (int64, error) {
	writeAPI := client.WriteAPIBlocking(config.InfluxDB.Org, config.InfluxDB.Bucket)

	historyID := time.Now().UnixNano()

	p := influxdb2.NewPoint(
		"measurement_history",
		map[string]string{
			"history_id": fmt.Sprintf("%d", historyID),
		},
		map[string]interface{}{
			"start_time":    startTime.Unix(),
			"end_time":      endTime.Unix(),
			"status":        status,
			"device_count":  deviceCount,
			"channel_count": channelCount,
		},
		time.Now(),
	)

	err := writeAPI.WritePoint(context.Background(), p)
	if err != nil {
		log.Printf("Failed to write measurement history: %v", err)
		return 0, err
	}
	log.Printf("Writing measurement history: start_time=%v, end_time=%v, status=%s, device_count=%d, channel_count=%d, history_id=%d",
		startTime.Unix(), endTime.Unix(), status, deviceCount, channelCount, historyID)
	log.Println("Successfully wrote measurement history")
	return historyID, nil
}

func WriteMeasurementData(historyID int64, deviceID, channelID int, voltages []float64) error {
	writeAPI := client.WriteAPI(config.InfluxDB.Org, config.InfluxDB.Bucket)

	for i, voltage := range voltages {
		p := influxdb2.NewPoint(
			"measurement_data",
			map[string]string{
				"history_id": fmt.Sprintf("%d", historyID),
				"device_id":  fmt.Sprintf("%d", deviceID),
				"channel_id": fmt.Sprintf("%d", channelID),
			},
			map[string]interface{}{
				"value": voltage,
			},
			time.Now().Add(time.Duration(i)*time.Millisecond),
		)

		writeAPI.WritePoint(p)
	}

	writeAPI.Flush()

	log.Printf("Successfully wrote %d voltage measurements for device %d, channel %d, history_id %d", len(voltages), deviceID, channelID, historyID)
	return nil
}

func GetMeasurementHistory() ([]map[string]interface{}, error) {
	queryAPI := client.QueryAPI(config.InfluxDB.Org)

	query := fmt.Sprintf(`
        from(bucket:"%s")
            |> range(start: -1h)
            |> filter(fn: (r) => r._measurement == "measurement_history")
            |> pivot(rowKey:["_time", "history_id"], columnKey: ["_field"], valueColumn: "_value")
            |> sort(columns: ["_time"], desc: true)
            |> limit(n: 10)
    `, config.InfluxDB.Bucket)

	log.Printf("Executing query: %s", query)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		return nil, err
	}
	defer result.Close()

	var history []map[string]interface{}
	for result.Next() {
		record := result.Record()
		log.Printf("Raw record: %+v", record.Values())
		historyEntry := map[string]interface{}{
			"start_time":    record.ValueByKey("start_time"),
			"end_time":      record.ValueByKey("end_time"),
			"status":        record.ValueByKey("status"),
			"device_count":  record.ValueByKey("device_count"),
			"channel_count": record.ValueByKey("channel_count"),
			"history_id":    record.ValueByKey("history_id"),
			"timestamp":     record.Time().Unix(),
		}
		history = append(history, historyEntry)

		log.Printf("Record: %+v (Time: %s)", historyEntry, time.Unix(record.Time().Unix(), 0).Format(time.RFC3339))
	}

	if result.Err() != nil {
		log.Printf("Result error: %v", result.Err())
	}

	log.Printf("Returning %d measurement history records", len(history))
	for i, entry := range history {
		log.Printf("Record %d: %+v (Time: %s)", i+1, entry, time.Unix(entry["timestamp"].(int64), 0).Format(time.RFC3339))
	}

	return history, nil
}

func GetHistoricalData(historyID int64) ([]map[string]interface{}, error) {
	queryAPI := client.QueryAPI(config.InfluxDB.Org)

	query := fmt.Sprintf(`
        from(bucket:"%s")
            |> range(start: -1h)
            |> filter(fn: (r) => r._measurement == "measurement_data" and r.history_id == "%d")
            |> sort(columns: ["_time"])
            |> limit(n: 1000)
    `, config.InfluxDB.Bucket, historyID)

	log.Printf("Executing query: %s", query)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		return nil, err
	}
	defer result.Close()

	var voltages []map[string]interface{}
	for result.Next() {
		record := result.Record()
		voltages = append(voltages, map[string]interface{}{
			"value":   record.Value(),
			"time":    record.Time().Unix(),
			"device":  record.ValueByKey("device_id"),
			"channel": record.ValueByKey("channel_id"),
			"history": record.ValueByKey("history_id"),
		})
	}

	if result.Err() != nil {
		log.Printf("Result error: %v", result.Err())
	}

	log.Printf("Successfully retrieved %d records for history ID %d", len(voltages), historyID)
	return voltages, nil
}
