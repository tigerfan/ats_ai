# dut.py

import json
from scpi_server import SCPIServer

# 读取配置文件
with open('config.json', 'r') as config_file:
    config = json.load(config_file)

# 测量配置
NUM_DEVICES = config['measurement']['devices']
NUM_CHANNELS = config['measurement']['channels']
SAMPLE_RATE = config['measurement']['sample_rate']
MEASUREMENT_DURATION = config['measurement']['measurement_duration']
FULL_SCALE = config['measurement']['full_scale']

# SCPI 服务器配置
SCPI_HOST = config['scpi_server']['host']
SCPI_PORT_BASE = config['scpi_server']['port']

def main():
    server = SCPIServer(SCPI_HOST, SCPI_PORT_BASE, NUM_DEVICES, NUM_CHANNELS, SAMPLE_RATE, MEASUREMENT_DURATION, FULL_SCALE)
    server.run()

if __name__ == "__main__": 
    main()