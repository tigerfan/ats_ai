# ats.py

import json
from scpi_server import SCPIServer

# 读取配置文件
with open('config.json', 'r') as config_file:
    config = json.load(config_file)

# 测量配置
NUM_DEVICES = config['measurement']['devices']
NUM_CHANNELS = config['measurement']['channels']

# SCPI 服务器配置
SCPI_HOST = config['scpi_server']['host']
SCPI_PORT = config['scpi_server']['port']

def main():
    server = SCPIServer(SCPI_HOST, SCPI_PORT, NUM_DEVICES, NUM_CHANNELS)
    server.run()

if __name__ == "__main__":
    main()