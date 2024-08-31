```markdown
# ATS (Automated Test System) 项目

## 项目概述

ATS项目是一个自动化测试系统，用于同时测量多台设备的多个通道。系统包括前端用户界面、后端服务和设备模拟器。

### 主要特性

- 支持同时测试12台设备，每台设备18个通道
- 实时数据采集和显示
- 历史数据查询和可视化
- 模拟设备支持，便于系统测试和调试

## 系统架构

1. 前端：使用Svelte框架开发，通过WebSocket与后端通信
2. 后端：使用Go语言开发，采用Gin框架处理HTTP请求，WebSocket进行实时通信
3. 数据存储：使用InfluxDB存储测量数据
4. 设备模拟器：使用Python开发，模拟SCPI协议的设备行为

## 安装说明

### 前端

1. 进入frontend目录
2. 安装依赖：
   ```
   npm install
   ```
3. 构建项目：
   ```
   npm run build
   ```

### 后端

1. 确保已安装Go 1.16或更高版本
2. 进入backend目录
3. 安装依赖：
   ```
   go mod tidy
   ```
4. 编译项目：
   ```
   go build -o test.exe ./cmd/main.go
   ```

### 设备模拟器

1. 确保已安装Python 3.7或更高版本
2. 进入device_simulator目录
3. 安装依赖：
   ```
   pip install -r requirements.txt
   ```
4. 构建：
   ```
   pyinstall --onefile dut.py
   ```

### 数据库

1. 安装InfluxDB 2.0或更高版本
2. 创建一个新的bucket用于存储ATS数据

## 使用方法

1. 启动InfluxDB服务

2. 启动后端服务：
   ```
   test
   ```

3. 启动设备模拟器：
   ```
   dut 
   ```

4. 启动前端开发服务器（用于开发环境）：
   ```
   npm run dev
   ```
   或者使用构建后的文件部署到Web服务器

5. 在浏览器中访问前端应用（默认地址：http://localhost:5000）

## 配置

- 后端配置：`config.json` 文件

## 开发

### 前端开发

- 修改 `frontend/src` 目录下的文件
- 使用 `npm run dev` 启动开发服务器，支持热重载

### 后端开发

- 修改 `backend` 目录下的Go文件
- 使用 `go run cmd/main.go` 启动开发模式的服务器

### 设备模拟器开发

- 修改 `device_simulator/ats_device_simulator.py` 文件
- 直接运行 Python 脚本进行测试

## 贡献

欢迎提交问题报告和拉取请求。对于重大更改，请先开issue讨论您想要更改的内容。

## 许可证

[MIT](https://choosealicense.com/licenses/mit/)
```
