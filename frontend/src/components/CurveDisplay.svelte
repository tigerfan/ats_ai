<script>
  import { onMount } from 'svelte';
  import { currentMeasurementData } from '../stores/measurementStore.js';
  import { selectedHistoricalChannel } from '../stores/measurementStore.js';
  import * as echarts from 'echarts';

  let chartCanvas;
  let chart;

  onMount(() => {
    chart = echarts.init(chartCanvas);

    const updateChart = () => {
      if (!$currentMeasurementData) return;

      const option = {
        tooltip: {
          trigger: 'axis'
        },
        legend: {
          data: [`设备${$currentMeasurementData.device}通道${$currentMeasurementData.channel}`]
        },
        xAxis: {
          type: 'category',
          name: '采样点',
          data: $currentMeasurementData.voltages.map(v => v.time)
        },
        yAxis: {
          type: 'value',
          name: '电压(mV)',
          max: 5000 // Fix the Y-axis at 5000mV
        },
        series: [{
          name: `设备${$currentMeasurementData.device}通道${$currentMeasurementData.channel}`,
          type: 'line',
          data: $currentMeasurementData.voltages.map(v => v.value),
          smooth: true,
          lineStyle: {
            color: 'rgba(0, 216, 230, 1)',
            width: 2
          },
          areaStyle: {
            color: 'rgba(0, 216, 230, 0.1)'
          },
          symbol: 'none'
        }]
      };

      chart.setOption(option);
    };

    updateChart();

    return () => {
      chart.dispose();
    };
  });

  $: if (chart && $currentMeasurementData) {
    const updateChart = () => {
      const option = {
        xAxis: {
          data: $currentMeasurementData.voltages.map(v => v.time)
        },
        series: [{
          name: `设备${$currentMeasurementData.device}通道${$currentMeasurementData.channel}`,
          data: $currentMeasurementData.voltages.map(v => v.value)
        }]
      };

      chart.setOption(option);
    };

    updateChart();
  }

  //$: console.log('曲线组件中的数据:', $currentMeasurementData);
</script>

<div class="curve-display">
  <div bind:this={chartCanvas} style="width: 100%; height: 300px;"></div>
</div>

<style>
.curve-display {
  width: 100%;
  height: 300px;
  margin-bottom: 20px;
}
</style>
