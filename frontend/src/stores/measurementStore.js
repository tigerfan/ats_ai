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
export const selectedResult = writable(null);
export const allChannelData = writable({}); // 新增：存储所有通道的数据

export function ininitializeMatrix() {
  const totalDevices = 12;
  const totalChannels = 18;

  measurementResults.set(Array(totalDevices * totalChannels).fill().map((_, index) => ({
    device: Math.floor(index / totalChannels) + 1,
    channel: (index % totalChannels) + 1,
    value: 0,
    passed: false,
    tested: false
  })));
}

export function initializeStores() {
  ininitializeMatrix();

  measurementHistory.set(Array(15).fill().map((_, i) => ({
    timestamp: Date.now() - i * 60000,
    deviceCount: Math.floor(Math.random() * 12) + 1,
    channelCount: Math.floor(Math.random() * 18) + 1,
    status: ['completed', 'failed', 'in progress'][Math.floor(Math.random() * 3)]
  })));

  currentMeasurementData.set({
    channel: null,
    device: null,
    voltages: []
  });

  measurementProgress.set(0);
  allChannelData.set({}); // 初始化所有通道数据存储
}

initializeStores();

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

  measurementResults.update(results => {
    const index = results.findIndex(r => r.device === data.device && r.channel === data.channel);
    if (index !== -1) {
      results[index] = {
        ...results[index],
        value: Math.round((data.voltages[data.voltages.length - 1] / 65536) * 5000),
        passed: data.passed,
        tested: true
      };
    }
    return results;
  });

  // 更新所有通道数据存储
  allChannelData.update(channelData => {
    const key = `${data.device}-${data.channel}`;
    channelData[key] = {
      device: data.device,
      channel: data.channel,
      voltages: data.voltages.map((value, index) => ({ time: index, value: Math.round((value / 65536) * 5000) })),
      passed: data.passed
    };
    return channelData;
  });
}

export function selectResult(result) {
  selectedResult.set(result);
  allChannelData.update(channelData => {
    const key = `${result.device}-${result.channel}`;
    const data = channelData[key] || {
      channel: result.channel,
      device: result.device,
      voltages: [],
      passed: result.passed
    };
    currentMeasurementData.set(data);
    return channelData;
  });
}