<script>
  import { selectedDevices } from './stores/measurementStore.js';
  import { selectedChannels } from './stores/measurementStore.js';
  import { currentMeasurementData } from './stores/measurementStore.js';
  import { measurementStatus } from './stores/measurementStore.js';
  import { measurementHistory } from './stores/measurementStore.js';
  import { measurementResults } from './stores/measurementStore.js';
  import { measurementProgress } from './stores/measurementStore.js';
  import { progressStatus } from './stores/measurementStore.js';
  import { simulateMeasurement } from './stores/measurementStore.js';
  import ControlPanel from './components/ControlPanel.svelte';
  import CurveDisplay from './components/CurveDisplay.svelte';
  import ResultMatrix from './components/ResultMatrix.svelte';
  import MeasurementHistory from './components/MeasurementHistory.svelte';
  import { websocketStatus, initializeWebSocket } from './utils/websocket.js';

  initializeWebSocket();

  let selectedDevice = null;
  let selectedChannel = null;

  function handleSelectChannel(event) {
    if ($measurementStatus !== 'running') {
      selectedDevice = event.detail.device;
      selectedChannel = event.detail.channel;
    }
  }

  $: chartData = $measurementStatus === 'running' 
    ? $currentMeasurementData.voltages 
    : (selectedDevice === $currentMeasurementData.device && selectedChannel === $currentMeasurementData.channel)
      ? $currentMeasurementData.voltages
      : [];

  $: chartOptions = {
    title: {
      text: $measurementStatus === 'running' 
        ? `当前测试：设备 ${$currentMeasurementData.device} 通道 ${$currentMeasurementData.channel}`
        : selectedDevice && selectedChannel
          ? `设备 ${selectedDevice} 通道 ${selectedChannel} 的数据`
          : '请选择一个通道'
    },
    xAxis: {
      type: 'category',
      data: chartData.map(d => d.time)
    },
    yAxis: {
      type: 'value',
      name: '电压 (mV)'
    },
    series: [{
      data: chartData.map(d => d.value),
      type: 'line',
      name: '电压'
    }]
  };  

  let devices = Array.from({length: 12}, (_, i) => i + 1);
  let channels = Array.from({length: 18}, (_, i) => i + 1);
  
  function toggleDevice(device) {
    selectedDevices.update(devices => {
      if (devices.includes(device)) {
        return devices.filter(d => d !== device);
      } else {
        return [...devices, device];
      }
    });
  }
  
  function toggleAllDevices() {
    selectedDevices.update(devices => {
      return devices.length === 12 ? [] : Array.from({length: 12}, (_, i) => i + 1);
    });
  }
  
  function toggleChannel(channel) {
    selectedChannels.update(channels => {
      if (channels.includes(channel)) {
        return channels.filter(c => c !== channel);
      } else {
        return [...channels, channel];
      }
    });
  }
  
  function toggleAllChannels() {
    selectedChannels.update(channels => {
      return channels.length === 18 ? [] : Array.from({length: 18}, (_, i) => i + 1);
    });
  }
  
  let statusMessages = [];
  $: {
    if ($measurementStatus !== 'stopped') {
      statusMessages = [...statusMessages, `${new Date().toLocaleTimeString()}: ${$measurementStatus}`].slice(-5);
    } else {
      statusMessages = [];
    }
  }

  let indicatorTexts = ["network", "websocket", "scpi", "influxdb", "online"];
  $: indicatorStatus = [
    true,
    $websocketStatus === 'connected',
    false,
    false,
    true
  ];

</script>

<main>
  <h1>自动测试系统</h1>

  <div class="layout-container">
    <div class="content">
      <div class="left-panel">
        <div class="selection-panel">
          <div class="device-selection">
            <h3>选设备</h3>
            <div class="button-grid">
              {#each devices as device}
                <button 
                  class:selected={$selectedDevices.includes(device)}
                  on:click={() => toggleDevice(device)}
                >
                  {device}
                </button>
              {/each}
              <button class="all-button" on:click={toggleAllDevices}>ALL</button>
            </div>
          </div>

          <div class="channel-selection">
            <h3>选通道</h3>
            <div class="button-grid">
              {#each channels as channel}
                <button 
                  class:selected={$selectedChannels.includes(channel)}
                  on:click={() => toggleChannel(channel)}
                >
                  {channel}
                </button>
              {/each}
              <button class="all-button" on:click={toggleAllChannels}>ALL</button>
            </div>
          </div>
        </div>

        <ControlPanel />
        <div class="spacer"></div>
        <h3>历史记录</h3>
        <table>
          <tbody>
            <tr>
              <th>时间-</th>
              <th>设备数-</th>
              <th>通道数-</th>
              <th>测试结果</th>
            </tr>
          </tbody>
        </table>  
        <MeasurementHistory history={$measurementHistory.slice(0, 8)}/>
        <div class="spacer"></div>

        <div class="status-bar">
          <div class="status-messages">
            {#each statusMessages as message}
              <p>{message}</p>
            {/each}
          </div>
          <div class="progress-bar">
            <div class="progress" style="width: {$measurementProgress}%"></div>
          </div>
          <h3>当前进度: <span>{$progressStatus}</span></h3>
        </div>
      </div>

      <div class="right-panel">
        <ResultMatrix on:selectChannel={handleSelectChannel}/>
        <div class="spacer"></div>
        <CurveDisplay options={chartOptions}/>
        <div class="spacer"></div>
        <div class="spacer"></div>
        <div class="status-indicators">
          {#each Array(5) as _, i}
            <div class="indicator-container">
              <div class="indicator" class:active={indicatorStatus[i]}></div>
              <span class="indicator-text">{indicatorTexts[i]}</span>
            </div>
          {/each}
        </div>
      </div>
    </div>
  </div>
</main>

<style>
body {
  margin: 0;
  padding: 0;
  overflow: hidden;
}

main {
  max-width: 1600px;
  margin: 0 auto;
  padding: 10px;
  font-family: Arial, sans-serif;
  height: 95vh;
  display: flex;
  flex-direction: column;
}

h1 {
  text-align: center;
  color: #333;
  margin: 0 0 10px 0;
  font-size: 1.5em;
}

.layout-container {
  display: flex;
  flex-direction: column;
  flex-grow: 1;
  overflow: hidden;
}

.content {
  display: flex;
  flex-grow: 1;
  overflow: hidden;
}

.left-panel, .right-panel {
  width: 48%;
  padding: 20px;
  overflow-y: hidden;
}

.selection-panel {
  margin-top: 5px;
}

.device-selection, .channel-selection {
  margin-bottom: 5px;
}

h3 {
  margin: 5px 0;
  font-size: 1em;
}

.button-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 5px;
}

button {
  padding: 5px;
  border: 1px solid #ddd;
  background-color: #f0f0f0;
  cursor: pointer;
  font-size: 0.8em;
}

button.selected {
  background-color: #566087;
  color: white;
}

.all-button {
  grid-column: span 2;
  background-color: #2196F3;
  color: white;
}

.status-bar {
  background-color: #f9f9f9;
  padding: 10px;
  border-radius: 5px;
  border: 2px solid #ddd;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  display: flex;
  flex-direction: column;
  height: 170px;
  margin-top: 5px;
}

.status-messages {
  flex-grow: 1;
  overflow-y: hidden;
  margin-bottom: 5px;
}

.status-messages p {
  margin: 2px 0;
  font-size: 0.9em;
}

.progress-bar {
  width: 100%;
  height: 10px;
  background-color: #e0e0e0;
  border-radius: 1px;
  overflow: auto;
  margin-bottom: 5px;
}

.progress {
  height: 100%;
  background-color: #4CAF50;
  transition: width 0.3s ease-in-out;
}

.status-bar h3 {
  margin: 5px 0 0 0;
  color: #333;
  display: flex;
  justify-content: center;
  align-items: center;
}

.status-bar span {
  font-weight: bold;
  color: #4CAF50;
  margin-left: 10px;
  padding: 2px 8px;
  background-color: #e8f5e9;
  border-radius: 3px;
  border: 1px solid #4CAF50;
}

.spacer {
  height: 20px; /* 你可以根据需要调整高度 */
}

.status-indicators {
  display: flex;
  justify-content: space-between;
  margin: 20px 100px;
}

.indicator-container {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.indicator {
  width: 20px;
  height: 20px;
  background-color: gray;
  border-radius: 50%;
}

.indicator.active {
  background-color: green;
}

.indicator-text {
  margin-top: 5px;
  font-size: 20px;
}

</style>