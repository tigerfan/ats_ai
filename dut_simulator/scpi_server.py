# scpi_server.py

import threading
import socket
import json
from measurement_device import MeasurementDevice

class SCPIServer:
    def __init__(self, host, port_base, num_devices, num_channels):
        self.host = host
        self.port_base = port_base
        self.num_devices = num_devices
        self.num_channels = num_channels
        self.devices = {i: MeasurementDevice(i, num_channels) for i in range(1, num_devices + 1)}
        for device in self.devices.values():
            device.start()

    def handle_client(self, client_socket, device_id):
        while True:
            try:
                data = client_socket.recv(8192).decode().strip()
                if not data:
                    break
                response = self.process_command(data, device_id)
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

    def process_command(self, command, device_id):
        print(f"Received command: {command}")
        command = command.upper()
        parts = command.split(':')

        if len(parts) < 2:
            return "ERROR: Invalid command"

        try:
            channel = int(parts[1])

            # 检查通道号的合法性
            if channel < 1 or channel > self.num_channels:
                return f"ERROR: Invalid channel number {channel}"

            device = self.devices[device_id]
            cmd = parts[0]

            if cmd == "MEAS":
                data = device.start_measurement(channel)
                return json.dumps(data)
            else:
                return f"ERROR: Unknown command {cmd}"
        except Exception as e:
            print(f"Error processing command: {str(e)}")
            return f"ERROR: {str(e)}"

    def run_device_server(self, device_id):
        port = self.port_base + device_id - 1
        server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        server_socket.bind((self.host, port))
        server_socket.listen(5)
        print(f"SCPI Server for device {device_id} running on {self.host}:{port}")

        while True:
            client_socket, addr = server_socket.accept()
            print(f"Accepted connection from {addr} for device {device_id}")
            client_handler = threading.Thread(target=self.handle_client, args=(client_socket, device_id))
            client_handler.start()

    def run(self):
        for device_id in range(1, self.num_devices + 1):
            threading.Thread(target=self.run_device_server, args=(device_id,)).start()
