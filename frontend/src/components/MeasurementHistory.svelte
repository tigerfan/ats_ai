<script>
  export let history = [];
</script>
  
<div class="measurement-history scrollable-container">
  <table>
    <tbody>
      {#each history as record}
        <tr>
          <td>{new Date(record.timestamp).toLocaleString()}</td>
          <td>{record.deviceCount}</td>
          <td>{record.channelCount}</td>
          <td>{record.status}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
  
<style>
.measurement-history {
  margin-bottom: 10px;
}

.scrollable-container {
  max-height: fit-content; /* 设置最大高度，可以根据需要调整 */
  overflow-y: scroll; /* 显示垂直滚动条 */
}

h3 {
  margin: 5px 0;
  font-size: 1em;
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8em;
}

th, td {
  border: 1px solid #ddd;
  padding: 4px;
  text-align: left;
}

th {
  background-color: #f2f2f2;
}
</style>
