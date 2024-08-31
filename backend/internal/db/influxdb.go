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
	// 读取 config.json 文件
	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Printf("Failed to read config file: %v", err)
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 将 JSON 数据解码到 config 结构体中
	if err := json.Unmarshal(configFile, &config); err != nil {
		log.Printf("Failed to unmarshal config file: %v", err)
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	// 初始化 InfluxDB 客户端
	client = influxdb2.NewClient(config.InfluxDB.URL, config.InfluxDB.Token)
	log.Println("InfluxDB client initialized successfully")
	return nil
}

func WriteMeasurementHistory(startTime, endTime time.Time, status string, deviceCount, channelCount int) (int64, error) {
	writeAPI := client.WriteAPIBlocking(config.InfluxDB.Org, config.InfluxDB.Bucket)

	historyID := time.Now().UnixNano() // 使用当前时间的纳秒值作为唯一标识符

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

		// 非阻塞写入
		writeAPI.WritePoint(p)
	}

	// 确保数据被刷新到服务器并处理完毕
	writeAPI.Flush()

	log.Printf("Successfully wrote %d voltage measurements for device %d, channel %d", len(voltages), deviceID, channelID)
	return nil
}

func GetMeasurementHistory() ([]map[string]interface{}, error) {
	queryAPI := client.QueryAPI(config.InfluxDB.Org)

	query := fmt.Sprintf(`
        from(bucket:"%s")
            |> range(start: -30d)
            |> filter(fn: (r) => r._measurement == "measurement_history")
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
		history = append(history, map[string]interface{}{
			"start_time":    record.ValueByKey("start_time"),
			"end_time":      record.ValueByKey("end_time"),
			"status":        record.ValueByKey("status"),
			"device_count":  record.ValueByKey("device_count"),
			"channel_count": record.ValueByKey("channel_count"),
		})
	}

	if result.Err() != nil {
		log.Printf("Result error: %v", result.Err())
	}

	log.Printf("Returning measurement history: %v", history)
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
