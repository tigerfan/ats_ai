# data_generator.py

import numpy as np
import random

class DataGenerator:
    def __init__(self, sample_rate, measurement_duration, full_scale):
        self.sample_rate = sample_rate
        self.measurement_duration = measurement_duration
        self.full_scale = full_scale

    def generate_sample_data(self):
        """生成模拟的样本数据"""
        num_samples = int(self.sample_rate * self.measurement_duration)

        # 生成低-高-低的电压序列
        low_voltage = np.random.uniform(0.005, 0.015)  # 5-15mV

        # 1% 的概率生成错误数据
        if random.random() < 0.01:
            high_voltage = np.random.uniform(2.3, 2.7)  # 错误情况：2.3-2.7V
        else:
            high_voltage = np.random.uniform(3.8, 4.2)  # 正常情况：3.8-4.2V

        low_samples = int(num_samples * 0.1)  # 10% of samples for initial low state
        high_samples = int(num_samples * 0.8)  # 80% of samples for high state
        final_low_samples = num_samples - low_samples - high_samples  # Remaining samples for final low state

        voltage_data = np.concatenate([
            np.random.normal(low_voltage, 0.001, low_samples),
            np.random.normal(high_voltage, 0.05, high_samples),
            np.random.normal(low_voltage, 0.001, final_low_samples)
        ])

        voltage_data = np.clip(voltage_data, 0, self.full_scale)
        int_data = np.round(voltage_data / self.full_scale * 65535).astype(np.uint16)
        return int_data.tolist()[:num_samples]