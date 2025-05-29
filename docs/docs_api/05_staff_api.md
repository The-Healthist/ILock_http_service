# 物业员工接口

## 获取物业员工列表

- **路径**: `/api/staffs`
- **方法**: GET
- **描述**: 获取所有物业员工的列表，支持分页和搜索
- **参数**:
  - `page`: 页码，默认 1
  - `page_size`: 每页条数，默认 10
  - `search`: 搜索关键词
- **响应**: 物业员工列表

## 获取带设备信息的物业员工列表

- **路径**: `/api/staffs/with-devices`
- **方法**: GET
- **描述**: 获取所有物业员工的列表及其关联的设备信息
- **参数**: 同获取物业员工列表
- **响应**: 带设备信息的物业员工列表

## 获取物业员工详情

- **路径**: `/api/staffs/:id`
- **方法**: GET
- **描述**: 根据 ID 获取特定物业员工的详细信息
- **响应**: 物业员工详情

## 创建物业员工

- **路径**: `/api/staffs`
- **方法**: POST
- **描述**: 创建一个新的物业员工
- **参数**:
  ```json
  {
  	"name": "王物业",
  	"phone": "13700001234",
  	"property_name": "阳光花园小区",
  	"position": "物业经理",
  	"role": "manager",
  	"status": "active",
  	"remark": "负责A区日常管理工作",
  	"username": "wangwuye",
  	"password": "Property@123",
  	"device_ids": [1, 2, 3]
  }
  ```
- **响应**: 创建的物业员工信息

## 更新物业员工

- **路径**: `/api/staffs/:id`
- **方法**: PUT
- **描述**: 更新现有物业员工的信息
- **参数**:
  ```json
  {
  	"name": "李物业",
  	"phone": "13700005678",
  	"property_name": "幸福家园小区",
  	"position": "前台客服",
  	"role": "staff",
  	"status": "active",
  	"remark": "负责接待访客和处理居民投诉",
  	"username": "liwuye",
  	"password": "NewProperty@456",
  	"device_ids": [1, 3, 5]
  }
  ```
- **响应**: 更新后的物业员工信息

## 删除物业员工

- **路径**: `/api/staffs/:id`
- **方法**: DELETE
- **描述**: 删除指定 ID 的物业员工
- **响应**: 操作结果
