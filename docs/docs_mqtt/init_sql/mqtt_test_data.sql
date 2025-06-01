-- MQTT测试数据初始化SQL脚本
-- 用于在DataGrip中直接执行，快速创建MQTT通话测试所需的测试数据

-- 设置SQL模式，避免一些警告
SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

-- 清理旧的测试数据（如果存在）
DELETE FROM call_records WHERE device_id IN (SELECT id FROM devices WHERE serial_number = 'DEV-MQTT-TEST-001');
DELETE FROM devices WHERE serial_number = 'DEV-MQTT-TEST-001';
DELETE FROM residents WHERE phone = '13800138000';
DELETE FROM households WHERE household_number = 'MQTT-101';
DELETE FROM buildings WHERE building_code = 'MQTT-TEST';

-- 1. 创建测试建筑
INSERT INTO buildings (building_name, building_code, address, status, created_at, updated_at)
VALUES ('MQTT测试小区', 'MQTT-TEST', 'MQTT测试地址', 'active', NOW(), NOW());

-- 获取刚插入的建筑ID
SET @building_id = LAST_INSERT_ID();

-- 2. 创建测试户号
INSERT INTO households (household_number, building_id, status, created_at, updated_at)
VALUES ('MQTT-101', @building_id, 'active', NOW(), NOW());

-- 获取刚插入的户号ID
SET @household_id = LAST_INSERT_ID();

-- 3. 创建多个测试住户（便于测试多人接听场景）
-- 住户1
INSERT INTO residents (name, phone, email, password, household_id, created_at, updated_at)
VALUES ('MQTT测试用户1', '13800138000', 'mqtt_test1@example.com', 'password123', @household_id, NOW(), NOW());

-- 获取住户1的ID
SET @resident1_id = LAST_INSERT_ID();

-- 住户2
INSERT INTO residents (name, phone, email, password, household_id, created_at, updated_at)
VALUES ('MQTT测试用户2', '13800138001', 'mqtt_test2@example.com', 'password123', @household_id, NOW(), NOW());

-- 获取住户2的ID
SET @resident2_id = LAST_INSERT_ID();

-- 4. 创建测试设备
INSERT INTO devices (name, serial_number, location, status, building_id, household_id, created_at, updated_at)
VALUES ('MQTT测试门禁', 'DEV-MQTT-TEST-001', 'MQTT测试小区大门', 'online', @building_id, @household_id, NOW(), NOW());

-- 获取设备ID
SET @device_id = LAST_INSERT_ID();

-- 5. 创建一些历史通话记录（可选）
INSERT INTO call_records (call_id, device_id, resident_id, call_status, timestamp, duration, created_at, updated_at)
VALUES 
('mqtt-test-call-1', @device_id, @resident1_id, 'answered', DATE_SUB(NOW(), INTERVAL 1 DAY), 45, NOW(), NOW()),
('mqtt-test-call-2', @device_id, @resident2_id, 'missed', DATE_SUB(NOW(), INTERVAL 2 DAY), 0, NOW(), NOW()),
('mqtt-test-call-3', @device_id, @resident1_id, 'timeout', DATE_SUB(NOW(), INTERVAL 3 DAY), 0, NOW(), NOW());

-- 输出创建的测试数据信息
SELECT 'MQTT测试数据创建成功！' AS '执行结果';

-- 输出测试信息
SELECT @building_id AS building_id, 'MQTT-TEST' AS building_code, 
       @household_id AS household_id, 'MQTT-101' AS household_number,
       @resident1_id AS resident1_id, '13800138000' AS resident1_phone,
       @resident2_id AS resident2_id, '13800138001' AS resident2_phone,
       @device_id AS device_id, 'DEV-MQTT-TEST-001' AS device_serial_number;

-- MQTT通话API示例说明
/*
MQTT通话流程测试API示例:

1. 发起通话:
   POST /api/mqtt/call
   请求体: {
     "device_id": "<device_id>",
     "household_number": "MQTT-101"
   }

2. 设备端控制:
   POST /api/mqtt/controller/device
   请求体: {
     "action": "hangup",
     "call_id": "<call_id>"
   }

3. 住户端控制:
   POST /api/mqtt/controller/resident
   请求体: {
     "action": "answered",
     "call_id": "<call_id>"
   }
*/ 