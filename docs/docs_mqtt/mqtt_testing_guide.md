# MQTT测试数据使用指南

为了方便测试MQTT通话功能，本项目提供了SQL初始化脚本来快速创建测试数据，无需手动逐个添加或使用Go脚本。

## 1. 数据初始化

### 使用DataGrip执行SQL脚本

1. 打开DataGrip连接到开发数据库
2. 打开文件 `initsql/mqtt_test_data.sql`
3. 执行整个脚本（或选择性执行部分内容）

脚本会自动创建:
- 测试小区（MQTT-TEST）
- 测试户号（MQTT-101）
- 两名测试住户（手机号：13800138000 和 13800138001）
- 测试设备（序列号：DEV-MQTT-TEST-001）
- 几条测试通话记录

**请记录这些ID值，用于后续API测试**

## 2. MQTT通话测试

### 启动MQTT服务

确保MQTT服务已经启动:

```bash
./start_mqtt.sh
```

### 使用MQTTX客户端

按照 `docs/mqtt_api_design.md` 文档设置三个MQTTX客户端:

1. **服务端**: 订阅 `mqtt_call/#`
2. **住户端**: 订阅 `mqtt_call/incoming`, `mqtt_call/controller/resident`, `mqtt_call/system`
3. **设备端**: 订阅 `mqtt_call/controller/device`, `mqtt_call/system`

### API测试流程

1. **发起通话**:
   ```
   POST /api/mqtt/call
   {
     "device_id": "5",
     "household_number": "MQTT-101"
   }
   ```

2. **住户接听通话**:
   ```
   POST /api/mqtt/controller/resident
   {
     "action": "answered",
     "call_id": "<从步骤1获得的call_id>"
   }
   ```

3. **设备挂断通话**:
   ```
   POST /api/mqtt/controller/device
   {
     "action": "hangup",
     "call_id": "<从步骤1获得的call_id>"
   }
   ```

4. **查询会话信息**:
   ```
   GET /api/mqtt/session?call_id=<从步骤1获得的call_id>
   ```

## 3. 清理测试数据

如需清理测试数据，可以执行以下SQL:

```sql
DELETE FROM call_records WHERE device_id IN (SELECT id FROM devices WHERE serial_number = 'DEV-MQTT-TEST-001');
DELETE FROM devices WHERE serial_number = 'DEV-MQTT-TEST-001';
DELETE FROM residents WHERE phone IN ('13800138000', '13800138001');
DELETE FROM households WHERE household_number = 'MQTT-101';
DELETE FROM buildings WHERE building_code = 'MQTT-TEST';
```

## 4. 故障排查

### MQTT连接问题

1. 确认MQTT服务已启动，并检查MQTT端口（1883）是否可访问
2. 检查MQTTX客户端配置是否正确

### API调用问题

1. 确认使用了正确的设备ID和户号
2. 检查call_id是否正确传递

如有其他问题，请查阅 `docs/mqtt_api_design.md` 获取更详细的API说明。 