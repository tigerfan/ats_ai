# measurement_device.py
import time
import threading
from data_generator import DataGenerator

class MeasurementDevice(threading.Thread):
    def __init__(self, device_id, num_channels, sample_rate, measurement_duration, full_scale):
        threading.Thread.__init__(self)
        self.device_id = device_id
        self.channels = num_channels
        self.sample_rate = sample_rate
        self.measurement_duration = measurement_duration
        self.full_scale = full_scale
        self.data_generator = DataGenerator(self.sample_rate, self.measurement_duration, self.full_scale)

    def start_measurement(self, channel):
        # 延时 measurement_duration 时间
        time.sleep(self.measurement_duration)
       
        data = self.data_generator.generate_sample_data()
        return {str(channel): data}