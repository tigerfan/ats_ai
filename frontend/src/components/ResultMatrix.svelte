<script>
  import { measurementResults } from '../stores/measurementStore';
  import { selectedDevices } from '../stores/measurementStore';
  import { selectedChannels } from '../stores/measurementStore';
  import { measurementStatus } from '../stores/measurementStore';
  import { selectResult } from '../stores/measurementStore';
  import { createEventDispatcher } from 'svelte';
  
  export let results = [];
  
  const dispatch = createEventDispatcher();
  
  $: {
    results = $measurementResults;
  }

  function handleCellClick(result) {
    selectResult(result);
  }

  function getStatus(result) {
    if (!$selectedDevices.includes(result.device) || !$selectedChannels.includes(result.channel)) {
      return 'unselected';
    }
    return result.tested ? (result.passed ? 'passed' : 'failed') : 'untested';
  }

  function getColor(status) {
    switch (status) {
      case 'passed': return 'green';
      case 'failed': return 'red';
      case 'untested': return 'gray';
      case 'unselected': return 'lightgray';
    }
  }

  function getStatusText(status) {
    switch (status) {
      case 'passed': return '正常';
      case 'failed': return '异常';
      case 'untested': return '未测';
      case 'unselected': return '未选';
    }
  }
</script>

<div class="result-matrix">
  <h3>判读结果</h3>
  <div class="matrix-container">
    {#each results as result, i}
      {@const status = getStatus(result)}
      <div 
        class="matrix-cell" 
        class:disabled={$measurementStatus === 'running'}
        style="background-color: {getColor(status)};"
        title="设备: {result.device}, 通道: {result.channel}, 状态: {getStatusText(status)}"
        on:click={() => handleCellClick(result)}
      ></div>
    {/each}
  </div>
</div>

<style>
  .result-matrix {
    margin-top: 20px;
  }

  .matrix-container {
    display: grid;
    grid-template-columns: repeat(18, 1fr);
    gap: 2px;
  }

  .matrix-cell {
    width: 35px;
    height: 20px;
    border: 1px solid #ccc;
  }

  .matrix-cell.disabled {
    cursor: not-allowed;
    opacity: 0.8;
  }
</style>