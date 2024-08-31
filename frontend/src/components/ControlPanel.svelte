<script>
  import { measurementStatus } from '../stores/measurementStore';
  import { simulateMeasurement } from '../stores/measurementStore';
  import { selectedDevices } from '../stores/measurementStore';
  import { selectedChannels } from '../stores/measurementStore';
  import { ininitializeMatrix } from '../stores/measurementStore';
  import { measurementProgress } from '../stores/measurementStore';
  import { sendMessage } from '../utils/websocket';
  
  function startMeasurement() {
    ininitializeMatrix();

    const devices = $selectedDevices;
    const channels = $selectedChannels;

    if (devices.length === 0 || channels.length === 0) {
      alert('设备或通道未选择');
      return;
    }
    
    $measurementStatus = 'running';
    
    const message = {
      action: 'start',
      devices: devices,
      channels: channels,
    };
    sendMessage(message);
  }
  
  function pauseMeasurement() {
    $measurementStatus = 'paused';
    sendMessage({ action: 'pause' });
  }
  
  function resumeMeasurement() {
    $measurementStatus = 'running';
    sendMessage({ action: 'resume' });
  }

  function stopMeasurement() {
    $measurementStatus = 'stopped';
    sendMessage({ action: 'stop' });
  }

  function check1() {
    $measurementStatus = 'stopped';
    sendMessage({ action: 'getMeasurementHistory' });
  }
  
  function check2() {
    $measurementStatus = 'stopped';
    sendMessage({ action: 'getHistoricalData' });
  }
</script>
 
 <div class="control-panel">
  <button on:click={startMeasurement} disabled={$measurementStatus !== 'stopped'}>开始</button>
  {#if $measurementStatus === 'paused'}
    <button on:click={resumeMeasurement}>继续</button>
  {:else}
    <button on:click={pauseMeasurement} disabled={$measurementStatus !== 'running'}>暂停</button>
  {/if}
  <button on:click={stopMeasurement} disabled={$measurementStatus === 'stopped'}>
    停止
  </button>
  <button on:click={check1}>测量历史</button>
  <button on:click={check2}>历史数据</button>
</div>

<style>
  .control-panel {
    margin-top: 20px;
    display: flex;
    gap: 10px;
  }

  button {
    padding: 10px 30px;
    font-size: 16px;
    cursor: pointer;
  }

  button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>