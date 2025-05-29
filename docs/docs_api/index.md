# ILock HTTP Service API 文档

## 目录

- [认证接口](01_auth_api.md)
- [管理员接口](02_admin_api.md)
- [设备接口](03_device_api.md)
- [居民接口](04_resident_api.md)
- [物业员工接口](05_staff_api.md)
- [通话记录接口](06_call_record_api.md)
- [紧急情况接口](07_emergency_api.md)
- [楼号接口](08_building_api.md)
- [户号接口](09_household_api.md)
- [音视频通话接口](10_rtc_api.md)
- [健康检查接口](11_health_api.md)

## 简介

本文档提供了 ILock HTTP Service 的 API 接口说明，包括认证、管理员、设备、居民、物业员工、通话记录、紧急情况、楼号、户号、音视频通话和健康检查等模块的接口。

## 认证说明

除了登录接口外，所有 API 接口都需要在请求头中包含有效的 JWT 令牌进行认证：

```
Authorization: Bearer <your_token>
```

## 响应格式

所有 API 响应都遵循以下格式：

```json
{
	"code": 0, // 0 表示成功，非 0 表示错误
	"message": "成功", // 响应消息
	"data": {} // 响应数据，可能是对象或数组
}
```

## 错误码说明

- 0: 成功
- 400: 请求参数错误
- 401: 未授权或令牌无效
- 403: 权限不足
- 404: 资源不存在
- 500: 服务器内部错误
