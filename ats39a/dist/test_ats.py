# test_ats.py

import socket
import json

def send_command(host, port, command):
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as client_socket:
        client_socket.connect((host, port))
        client_socket.sendall(command.encode())
        response = client_socket.recv(4096).decode().strip()
        return response

def main():
    SCPI_HOST = 'localhost'  # 根据实际情况修改
    SCPI_PORT = 5025         # 根据实际情况修改

    # 测试命令
    commands = [
        "1:MEAS:1",
        "2:MEAS:2",
        "3:MEAS:3",
        "12:MEAS:18",
        "13:MEAS:1",  # 无效设备号
        "1:MEAS:19",  # 无效通道号
        "1:UNKNOWN:1"  # 无效命令
    ]

    for command in commands:
        print(f"Sending command: {command}")
        response = send_command(SCPI_HOST, SCPI_PORT, command)
        try:
            response_json = json.loads(response)
            print(f"Response: {json.dumps(response_json, indent=4)}")
        except json.JSONDecodeError:
            print(f"Response: {response}")

if __name__ == "__main__":
    main()
