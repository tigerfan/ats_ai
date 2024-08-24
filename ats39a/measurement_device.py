# measurement_device.py

import threading
from data_generator import DataGenerator

class MeasurementDevice(threading.Thread):
    def __init__(self, device_id, num_channels, sample_rate=1000, measurement_duration=1, full_scale=5):
        threading.Thread.__init__(self)
        self.device_id = device_id
        self.channels = num_channels
        self.sample_rate = sample_rate
        self.measurement_duration = measurement_duration
        self.full_scale = full_scale
        self.data_generator = DataGenerator(self.sample_rate, self.measurement_duration, self.full_scale)

    def start_measurement(self, channel):
        # 立即生成并返回测量数据
        data = self.data_generator.generate_sample_data()
        return {str(channel): data}