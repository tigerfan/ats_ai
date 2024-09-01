<script>
  export let history = [];

  function formatDate(timestamp) {
    const date = new Date(timestamp * 1000);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit', second:'2-digit'});
  }
</script>

<style>
  .history-container {
    border: 1px solid #e0e0e0;
    border-radius: 4px;
    overflow: hidden;
    font-size: 0.9em;
    max-height: 200px;
    overflow-y: auto;
  }
  .history-item {
    padding: 6px 10px;
    border-bottom: 1px solid #e0e0e0;
    display: flex;
    align-items: center;
  }
  .history-item:last-child {
    border-bottom: none;
  }
  .history-item:nth-child(even) {
    background-color: #f8f8f8;
  }
  .history-item:nth-child(odd) {
    background-color: #ffffff;
  }
  .timestamp {
    flex: 0 0 180px;
    font-weight: bold;
  }
  .details {
    flex: 1;
    display: flex;
    justify-content: space-between;
    color: #555;
  }
</style>

<div class="history-container">
  {#each history as record}
    <div class="history-item">
      <span class="timestamp">{formatDate(record.timestamp)}</span>
      <div class="details">
        <span>设备数: {record.device_count}</span>
        <span>通道数: {record.channel_count}</span>
        <span>状态: {record.status}</span>
      </div>
    </div>
  {/each}
</div>