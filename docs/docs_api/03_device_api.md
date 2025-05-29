# 设备接口

## 获取设备列表

- **路径**: `/api/devices`
- **方法**: GET
- **描述**: 获取所有设备的列表，支持按楼号筛选
- **参数**:
  - `building_id`: 楼号 ID，用于筛选特定楼号下的设备
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "成功",
  	"data": [
  		{
  			"id": 1,
  			"name": "门禁1号",
  			"serial_number": "SN12345678",
  			"location": "小区北门入口",
  			"status": "online",
  			"building_id": 1,
  			"building": {
  				"id": 1,
  				"building_name": "1号楼"
  			}
  		}
  	]
  }
  ```

## 获取设备详情

- **路径**: `/api/devices/:id`
- **方法**: GET
- **描述**: 根据 ID 获取设备信息
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "成功",
  	"data": {
  		"id": 1,
  		"name": "门禁1号",
  		"serial_number": "SN12345678",
  		"location": "小区北门入口",
  		"status": "online",
  		"building_id": 1,
  		"building": {
  			"id": 1,
  			"building_name": "1号楼"
  		},
  		"households": [
  			{
  				"id": 1,
  				"household_number": "1-1-101"
  			}
  		],
  		"staff": [
  			{
  				"id": 1,
  				"name": "王物业"
  			}
  		]
  	}
  }
  ```

## 创建设备

- **路径**: `/api/devices`
- **方法**: POST
- **描述**: 创建一个新的门禁设备
- **参数**:
  ```json
  {
  	"name": "门禁1号",
  	"serial_number": "SN12345678",
  	"status": "online",
  	"location": "小区北门入口",
  	"building_id": 1,
  	"staff_ids": [1, 2, 3] //可选
  }
  ```
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "成功",
  	"data": {
  		"id": 1,
  		"name": "门禁1号",
  		"serial_number": "SN12345678",
  		"location": "小区北门入口",
  		"status": "online",
  		"building_id": 1
  	}
  }
  ```

## 更新设备

- **路径**: `/api/devices/:id`
- **方法**: PUT
- **描述**: 根据 ID 更新设备信息，支持更新关联
- **参数**:
  ```json
  {
  	"name": "门禁1号",
  	"serial_number": "SN12345678",
  	"status": "online",
  	"location": "小区北门入口",
  	"building_id": 1,
  	"household_ids": [1, 2],
  	"staff_ids": [1, 2, 3] //可选
  }
  ```
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "成功",
  	"data": {
  		"id": 1,
  		"name": "门禁1号",
  		"serial_number": "SN12345678",
  		"location": "小区北门入口",
  		"status": "online",
  		"building_id": 1
  	}
  }
  ```

## 删除设备

- **路径**: `/api/devices/:id`
- **方法**: DELETE
- **描述**: 根据 ID 删除设备
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "成功",
  	"data": null
  }
  ```

## 获取设备状态

- **路径**: `/api/devices/:id/status`
- **方法**: GET
- **描述**: 获取设备的当前状态信息，包括在线状态、最后更新时间等
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "成功",
  	"data": {
  		"id": 1,
  		"name": "门禁1号",
  		"serial_number": "SN12345678",
  		"status": "online",
  		"location": "小区北门入口",
  		"last_online": "2023-01-01T00:00:00Z"
  	}
  }
  ```

## 设备健康检测

- **路径**: `/api/device/status`
- **方法**: POST
- **描述**: 设备用于报告在线状态的简单健康检测接口
- **参数**:
  ```json
  {
  	"device_id": "1"
  }
  ```
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "设备状态更新成功",
  	"data": {
  		"device_id": "1",
  		"status": "online",
  		"timestamp": "2023-01-01T00:00:00Z"
  	}
  }
  ```

## 关联设备与楼号

- **路径**: `/api/devices/:id/building`
- **方法**: POST
- **描述**: 将指定设备关联到楼号
- **参数**:
  ```json
  {
  	"building_id": 1
  }
  ```
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "设备与楼号关联成功",
  	"data": {
  		"id": 1,
  		"name": "门禁1号",
  		"building_id": 1
  	}
  }
  ```

## 获取设备关联的户号

- **路径**: `/api/devices/:id/households`
- **方法**: GET
- **描述**: 获取指定设备关联的所有户号
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "成功",
  	"data": [
  		{
  			"id": 1,
  			"household_number": "1-1-101",
  			"building_id": 1
  		}
  	]
  }
  ```

## 关联设备与户号

- **路径**: `/api/devices/:id/households`
- **方法**: POST
- **描述**: 将指定设备关联到户号
- **参数**:
  ```json
  {
  	"household_id": 1
  }
  ```
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "设备与户号关联成功",
  	"data": null
  }
  ```

## 解除设备与户号的关联

- **路径**: `/api/devices/:id/households`
- **方法**: DELETE
- **描述**: 解除指定设备与其当前关联的户号的关联
- **响应**:
  ```json
  {
  	"code": 0,
  	"message": "设备与户号关联已解除",
  	"data": null
  }
  ```
