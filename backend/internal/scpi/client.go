package scpi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
)

type Client struct {
	conn net.Conn
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Connect(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) SendCommand(command string) (string, error) {
	_, err := fmt.Fprintf(c.conn, command+"\n")
	if err != nil {
		return "", err
	}

	response, err := bufio.NewReader(c.conn).ReadString('\n')
	if err != nil {
		return "", err
	}

	return response, nil
}

func (c *Client) MeasureVoltage(device, channel int) ([]int, error) {
	command := fmt.Sprintf("%d:MEAS:%d", device, channel)
	response, err := c.SendCommand(command)
	if err != nil {
		return nil, err
	}

	var result map[string][]int
	err = json.Unmarshal([]byte(response), &result)
	if err != nil {
		return nil, err
	}

	channelStr := strconv.Itoa(channel)
	voltages, ok := result[channelStr]
	if !ok {
		return nil, fmt.Errorf("channel %d not found in response", channel)
	}

	return voltages, nil
}

func (c *Client) StartMeasurement(devices, channels []int) error {
	// In this implementation, we don't need a separate start command
	// as the measurement starts immediately when we send the MEAS command
	return nil
}

func (c *Client) StopMeasurement() error {
	// The SCPI server doesn't have a stop command in this implementation
	return nil
}

// SetSamplingRate and SetMeasurementDuration are not supported by the current SCPI server
// You may want to remove these methods or keep them as no-op for future compatibility

func (c *Client) SetSamplingRate(rate float64) error {
	return nil
}

func (c *Client) SetMeasurementDuration(duration float64) error {
	return nil
}
