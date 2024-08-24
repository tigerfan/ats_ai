# scpi_server.py

import threading
import socket
import json
from measurement_device import MeasurementDevice

class SCPIServer:
    def __init__(self, host, port, num_devices, num_channels):
        self.host = host
        self.port = port
        self.num_devices = num_devices
        self.num_channels = num_channels
        self.devices = {i: MeasurementDevice(i, num_channels) for i in range(1, num_devices + 1)}
        for device in self.devices.values():
            device.start()

    def handle_client(self, client_socket):
        while True:
            try:
                data = client_socket.recv(1024).decode().strip()
                if not data:
                    break
                response = self.process_command(data)
                response_bytes = (response + '\n').encode()
                total_sent = 0
                while total_sent < len(response_bytes):
                    sent = client_socket.send(response_bytes[total_sent:])
                    if sent == 0:
                        raise RuntimeError("socket connection broken")
                    total_sent += sent
            except Exception as e:
                print(f"Error handling client: {e}")
                break
        client_socket.close()

    def process_command(self, command):
        print(f"Received command: {command}")
        command = command.upper()
        parts = command.split(':')

        if len(parts) < 3:
            return "ERROR: Invalid command"

        try:
            device_id = int(parts[0])
            channel = int(parts[2])

            # 检查设备号和通道号的合法性
            if device_id < 1 or device_id > 12:
                return f"ERROR: Invalid device ID {device_id}"
            if channel < 1 or channel > 18:
                return f"ERROR: Invalid channel number {channel}"

            device = self.devices[device_id]
            cmd = parts[1]

            if cmd == "MEAS":
                data = device.start_measurement(channel)
                return json.dumps(data)
            else:
                return f"ERROR: Unknown command {cmd}"
        except Exception as e:
            print(f"Error processing command: {str(e)}")
            return f"ERROR: {str(e)}"

    def run(self):
        server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        server_socket.bind((self.host, self.port))
        server_socket.listen(5)
        print(f"SCPI Server running on {self.host}:{self.port}")

        while True:
            client_socket, addr = server_socket.accept()
            print(f"Accepted connection from {addr}")
            client_handler = threading.Thread(target=self.handle_client, args=(client_socket,))
            client_handler.start()