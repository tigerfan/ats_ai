import { writable } from 'svelte/store';

export const selectedDevices = writable([]);
export const selectedChannels = writable([]);
export const measurementStatus = writable('stopped');
export const measurementResults = writable([]);
export const measurementHistory = writable([]);
export const currentMeasurementData = writable([]);
export const measurementProgress = writable(0);
export const progressStatus = writable('Waiting');
export const selectedHistoricalChannel = writable(null);

export function initializeStores() {
  const totalDevices = 12;
  const totalChannels = 18;

  // Initialize measurement results
  measurementResults.set(Array(totalDevices * totalChannels).fill().map((_, index) => ({
    device: Math.floor(index / totalChannels) + 1,
    channel: (index % totalChannels) + 1,
    value: 0,
    passed: false, // 初始状态设为未通过
    tested: false // 新增：表示是否已被测试
  })));

  // Initialize measurement history (keeping this as is for now)
  measurementHistory.set(Array(15).fill().map((_, i) => ({
    timestamp: Date.now() - i * 60000,
    deviceCount: Math.floor(Math.random() * 12) + 1,
    channelCount: Math.floor(Math.random() * 18) + 1,
    status: ['completed', 'failed', 'in progress'][Math.floor(Math.random() * 3)]
  })));

  // Initialize current measurement data
  currentMeasurementData.set({
    channel: null,
    device: null,
    voltages: []
  });

  // Initialize measurement progress
  measurementProgress.set(0);
}

// Call this function when the application starts
initializeStores();

// Function to simulate measurement process
export function simulateMeasurement() {
  progressStatus.set('开始测量');
  measurementProgress.set(0);

  const interval = setInterval(() => {
    measurementProgress.update(p => {
      if (p >= 100) {
        clearInterval(interval);
        progressStatus.set('测量完成');
        return 100;
      }
      progressStatus.set(`测量进行中 ${p}%`);
      return p + 1;
    });
  }, 100);
}

export function updateCurrentMeasurementData(data) {
  currentMeasurementData.set({
    channel: data.channel,
    device: data.device,
    voltages: data.voltages.map((value, index) => ({ time: index, value: Math.round((value / 65536) * 5000) })),
    passed: data.passed
  });

  // Update the measurementResults store
  measurementResults.update(results => {
    const index = results.findIndex(r => r.device === data.device && r.channel === data.channel);
    if (index !== -1) {
      results[index] = {
        ...results[index],
        value: Math.round((data.voltages[data.voltages.length - 1] / 65536) * 5000),
        passed: data.passed,
        tested: true // 标记为已测试
      };
    }
    return results;
  });
}