package scpi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
)

type Client struct {
	connections map[int]net.Conn
	mu          sync.Mutex
}

func NewClient() *Client {
	return &Client{
		connections: make(map[int]net.Conn),
	}
}

func (c *Client) Connect(baseAddress string, basePort int, numDevices int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 1; i <= numDevices; i++ {
		address := fmt.Sprintf("%s:%d", baseAddress, basePort+i-1)
		conn, err := net.Dial("tcp", address)
		if err != nil {
			return fmt.Errorf("failed to connect to device %d: %w", i, err)
		}
		c.connections[i] = conn
	}
	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var lastErr error
	for _, conn := range c.connections {
		if err := conn.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (c *Client) SendCommand(device int, command string) (string, error) {
	c.mu.Lock()
	conn, ok := c.connections[device]
	c.mu.Unlock()

	if !ok {
		return "", fmt.Errorf("device %d not connected", device)
	}

	_, err := fmt.Fprintf(conn, command+"\n")
	if err != nil {
		return "", err
	}

	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", err
	}

	return response, nil
}

func (c *Client) MeasureVoltage(device, channel int) ([]int, error) {
	command := fmt.Sprintf("MEAS:%d", channel)
	response, err := c.SendCommand(device, command)
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
