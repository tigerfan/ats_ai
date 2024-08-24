package models

type Measurement struct {
	DeviceID  int       `json:"device_id"`
	ChannelID int       `json:"channel_id"`
	Voltages  []float64 `json:"voltages"`
	Timestamp int64     `json:"timestamp"`
	Status    string    `json:"status"`
}

type MeasurementResult struct {
	DeviceID  int   `json:"device_id"`
	ChannelID int   `json:"channel_id"`
	Success   bool  `json:"success"`
	Timestamp int64 `json:"timestamp"`
}

type MeasurementHistory struct {
	Timestamp    int64  `json:"timestamp"`
	DeviceCount  int    `json:"device_count"`
	ChannelCount int    `json:"channel_count"`
	Status       string `json:"status"`
}
