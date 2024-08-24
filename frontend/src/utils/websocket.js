import { writable } from 'svelte/store';
import { currentMeasurementData } from '../stores/measurementStore';
import { measurementResults } from '../stores/measurementStore';
import { measurementStatus } from '../stores/measurementStore';
import { updateCurrentMeasurementData } from '../stores/measurementStore';

export const websocketStatus = writable('disconnected');

let ws;
let messageQueue = [];
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 3;

export function initializeWebSocket() {
  if (ws && (ws.readyState === WebSocket.CONNECTING || ws.readyState === WebSocket.OPEN)) {
    console.log('WebSocket is already connecting or connected, not reinitializing');
    return;
  }

  if (ws) {
    ws.close();
  }

  const wsUrl = 'ws://localhost:5177/ws';  // 确保这个URL是正确的
  console.log(`尝试连接到 ${wsUrl}`);
  
  ws = new WebSocket(wsUrl);

  ws.onopen = () => {
    console.log('WebSocket连接已建立');
    websocketStatus.set('connected');
    reconnectAttempts = 0;
    // 连接建立后，发送队列中的所有消息
    while (messageQueue.length > 0) {
      const message = messageQueue.shift();
      ws.send(JSON.stringify(message));
    }
  };

  ws.onmessage = (event) => {
    //console.log('收到消息:', event.data);
    try {
      const message = JSON.parse(event.data);
  
      if (message.channel !== undefined && message.device !== undefined && message.voltages !== undefined) {
        // This is the measurement data message
        updateCurrentMeasurementData(message);
      } else {
        switch (message.type) {
          case 'measurement_results':
            measurementResults.set(message.results);
            break;
          case 'measurement_status':
            measurementStatus.set(message.status);
            break;
          default:
            console.log('未知的消息类型:', message.type);
        }
      }
    } catch (error) {
      console.error('解析消息时出错:', error);
    }
  };

  ws.onclose = (event) => {
    console.log('WebSocket连接已关闭', event);
    websocketStatus.set('disconnected');
    if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
      console.log(`尝试重新连接 (${reconnectAttempts + 1}/${MAX_RECONNECT_ATTEMPTS})`);
      setTimeout(initializeWebSocket, 5000);
      reconnectAttempts++;
    } else {
      console.log('达到最大重连次数，停止尝试');
    }
  };

  ws.onerror = (error) => {
    console.error('WebSocket错误:', error);
    websocketStatus.set('error');
  };

  return ws;
}

export function sendMessage(message) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(message));
    console.log('已发送消息:', message);
  } else {
    console.log('WebSocket未连接，消息已加入队列:', message);
    messageQueue.push(message);
    if (!ws || ws.readyState === WebSocket.CLOSED) {
      console.log('尝试重新初始化WebSocket');
      initializeWebSocket();
    }
  }
}

// 移除 manualConnect 函数

// 在模块加载时自动初始化WebSocket
initializeWebSocket();