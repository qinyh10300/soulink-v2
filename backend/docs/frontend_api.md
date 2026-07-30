# Soulink Web 后端 API 文档

本文档基于当前后端实现整理，供前端联调使用。

## 1. 基础信息

- 本机直连基地址：`http://localhost:8080`
- Docker + Nginx 联调基地址：`http://localhost:18080`
- API 前缀：`/api/v1`
- 请求与响应格式：`application/json`
- 鉴权方式：`Authorization: Bearer <token>`
- 时间字段格式：RFC3339，例如 `2026-03-27T21:20:00+08:00`

## 2. 统一响应格式

成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

错误响应：

```json
{
  "code": 4031,
  "message": "匹配结果将在 2026-03-27T21:20:00+08:00 公布",
  "data": null
}
```

### 2.1 常见业务码

| code | 含义 |
| --- | --- |
| 0 | 成功 |
| 4001 | 参数错误 |
| 4002 | 验证码错误 |
| 4003 | 验证码已过期 |
| 4004 | 验证码发送过于频繁 |
| 4005 | 邮箱格式错误 |
| 4006 | 密码格式错误 |
| 4011 | 未登录 |
| 4012 | Token 无效或已过期 |
| 4013 | 邮箱或密码错误 |
| 4031 | 无权限 / 当前不允许该操作 / 结果尚未公布 |
| 4041 | 用户不存在 |
| 4042 | 问卷不存在 |
| 4043 | 当前暂无匹配结果 |
| 4044 | 联系方式不存在 |
| 4091 | 邮箱已注册 |
| 4092 | 数据冲突 |
| 5001 | 服务器内部错误 |

## 3. 健康检查

### 3.1 健康检查

- 方法：`GET`
- 路径：`/healthz`
- 鉴权：否

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "status": "ok"
  }
}
```

## 4. 认证模块

### 4.1 登录

- 方法：`POST`
- 路径：`/api/v1/auth/login`
- 鉴权：否

请求体：

```json
{
  "email": "test@example.com",
  "password": "Start1234"
}
```

响应 `data`：

```json
{
  "token": "jwt_token",
  "user": {
    "id": 1,
    "email": "test@example.com",
    "nickname": "小王",
    "department": "计算机学院",
    "grade": "本科三年级",
    "hasCompletedQuestionnaire": false
  }
}
```

### 4.2 发送注册验证码

- 方法：`POST`
- 路径：`/api/v1/auth/register/send-code`
- 鉴权：否

请求体：

```json
{
  "email": "test@example.com"
}
```

响应 `data`：

```json
{
  "expireSeconds": 300
}
```

### 4.3 校验注册验证码

- 方法：`POST`
- 路径：`/api/v1/auth/register/verify-code`
- 鉴权：否

请求体：

```json
{
  "email": "test@example.com",
  "code": "123456"
}
```

响应 `data`：

```json
{
  "verified": true
}
```

### 4.4 注册并自动登录

- 方法：`POST`
- 路径：`/api/v1/auth/register`
- 鉴权：否

请求体：

```json
{
  "email": "test@example.com",
  "code": "123456",
  "password": "Start1234",
  "nickname": "小王",
  "department": "计算机学院",
  "grade": "本科三年级"
}
```

响应 `data` 与登录接口一致。

### 4.5 发送忘记密码验证码

- 方法：`POST`
- 路径：`/api/v1/auth/password/forgot/send-code`
- 鉴权：否

请求体：

```json
{
  "email": "test@example.com"
}
```

响应 `data`：

```json
{
  "expireSeconds": 300
}
```

### 4.6 忘记密码重置

- 方法：`POST`
- 路径：`/api/v1/auth/password/forgot/reset`
- 鉴权：否

请求体：

```json
{
  "email": "test@example.com",
  "code": "123456",
  "newPassword": "NewPass1234"
}
```

### 4.7 退出登录

- 方法：`POST`
- 路径：`/api/v1/auth/logout`
- 鉴权：是

成功响应：`message = "logout success"`

### 4.8 获取当前登录用户信息

- 方法：`GET`
- 路径：`/api/v1/auth/me`
- 鉴权：是

响应 `data`：

```json
{
  "id": 1,
  "email": "test@example.com",
  "nickname": "小王",
  "department": "计算机学院",
  "grade": "本科三年级",
  "contactType": "wechat",
  "contactValue": "wx_123",
  "hasCompletedQuestionnaire": true
}
```

### 4.9 发送修改密码验证码

- 方法：`POST`
- 路径：`/api/v1/users/me/password/send-code`
- 鉴权：是

响应 `data`：

```json
{
  "email": "test@example.com",
  "expireSeconds": 300
}
```

### 4.10 修改密码

- 方法：`POST`
- 路径：`/api/v1/users/me/password`
- 鉴权：是

请求体：

```json
{
  "code": "123456",
  "newPassword": "NewPass1234"
}
```

成功后当前用户历史会话会失效，需要重新登录。

### 4.11 发送注销账号验证码

- 方法：`POST`
- 路径：`/api/v1/users/me/delete/send-code`
- 鉴权：是

响应 `data`：

```json
{
  "email": "test@example.com",
  "expireSeconds": 300
}
```

### 4.12 注销账号

- 方法：`DELETE`
- 路径：`/api/v1/users/me`
- 鉴权：是

请求体：

```json
{
  "code": "123456"
}
```

成功响应：`message = "account deleted"`

## 5. 基础配置模块

### 5.1 获取院系列表

- 方法：`GET`
- 路径：`/api/v1/meta/departments`
- 鉴权：否

### 5.2 获取年级列表

- 方法：`GET`
- 路径：`/api/v1/meta/grades`
- 鉴权：否

### 5.3 获取联系方式类型列表

- 方法：`GET`
- 路径：`/api/v1/meta/contact-types`
- 鉴权：否

返回示例：

```json
[
  { "label": "微信", "value": "wechat" },
  { "label": "QQ", "value": "qq" },
  { "label": "手机号", "value": "phone" },
  { "label": "邮箱", "value": "email" },
  { "label": "其他", "value": "other" }
]
```


### 5.4 获取当前已注册人数

- 方法：`GET`
- 路径：`/api/v1/meta/registered-user-count`
- 鉴权：否

响应 `data`：

```json
{
  "registeredUserCount": 123
}
```

说明：

- 该接口返回当前未注销用户人数
- 后端会优先读取 Redis 缓存，缓存未命中时回源数据库并自动回填缓存
## 6. 个人资料与联系方式模块

说明：

- `contactType/contactValue` 是兼容字段
- `contacts` 是推荐字段
- 当前实现下，用户最多只会保留一条主动保存的联系方式
- 如果用户没有主动保存过联系方式，接口会回退返回注册邮箱，且该条会带 `readonly=true`

### 6.1 获取个人资料

- 方法：`GET`
- 路径：`/api/v1/users/me/profile`
- 鉴权：是

响应 `data`：

```json
{
  "id": 1,
  "email": "test@example.com",
  "nickname": "小王",
  "department": "计算机学院",
  "grade": "本科三年级",
  "contactType": "wechat",
  "contactValue": "wx_123",
  "contacts": [
    {
      "type": "wechat",
      "label": "微信",
      "value": "wx_123",
      "updatedAt": "2026-03-25T12:00:00Z"
    }
  ],
  "createdAt": "2026-03-25T11:30:00Z"
}
```

默认邮箱回退示例：

```json
{
  "contacts": [
    {
      "type": "email",
      "label": "邮箱",
      "value": "test@example.com",
      "updatedAt": "2026-03-25T11:30:00Z",
      "readonly": true
    }
  ]
}
```

### 6.2 更新个人资料

- 方法：`PUT`
- 路径：`/api/v1/users/me/profile`
- 鉴权：是

请求体：

```json
{
  "nickname": "小王",
  "department": "计算机学院",
  "grade": "本科三年级",
  "contactType": "wechat",
  "contactValue": "wx_123"
}
```

说明：

- 如果同时传入 `contactType/contactValue`，会把用户当前所有自定义联系方式替换成这一条

### 6.3 获取联系方式列表

- 方法：`GET`
- 路径：`/api/v1/users/me/contacts`
- 鉴权：是

响应 `data`：

```json
{
  "contacts": [
    {
      "type": "wechat",
      "label": "微信",
      "value": "wx_123",
      "updatedAt": "2026-03-25T12:00:00Z"
    }
  ]
}
```

### 6.4 保存联系方式

- 方法：`PUT`
- 路径：`/api/v1/users/me/contacts/{contactType}`
- 鉴权：是

路径参数：

- `contactType`：如 `wechat`、`qq`、`phone`、`email`、`other`

请求体：

```json
{
  "value": "wx_123"
}
```

响应 `data`：

```json
{
  "contacts": [
    {
      "type": "wechat",
      "label": "微信",
      "value": "wx_123",
      "updatedAt": "2026-03-25T12:00:00Z"
    }
  ]
}
```

说明：

- 每次保存前，后端会先删除之前所有自定义联系方式，再保存这条新的联系方式

### 6.5 删除联系方式

- 方法：`DELETE`
- 路径：`/api/v1/users/me/contacts/{contactType}`
- 鉴权：是

成功响应 `data`：

```json
{
  "contacts": [
    {
      "type": "email",
      "label": "邮箱",
      "value": "test@example.com",
      "updatedAt": "2026-03-25T11:30:00Z",
      "readonly": true
    }
  ]
}
```

## 7. 问卷模块

说明：

- 当前只支持“当前问卷”
- 问卷答案按当前问卷版本保存
- 非开放时间可以查看问卷，但保存答案会被拒绝
- 重新提交同一题时，会覆盖原答案
- 用户首次完成当前版本问卷后，会分配 `resultRevealAt`

### 7.1 获取当前问卷定义和已保存答案

- 方法：`GET`
- 路径：`/api/v1/questionnaires/current`
- 鉴权：是

响应 `data` 字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| questionnaireId | number | 问卷 ID |
| questionnaireKey | string | 问卷标识 |
| title | string | 问卷标题 |
| version | string | 问卷版本 |
| collectionStatus | string | `always_open` / `scheduled` / `open` / `closed` |
| acceptingResponses | boolean | 当前是否允许提交 |
| collectionStartsAt | string | 开放开始时间，可为空 |
| collectionEndsAt | string | 开放结束时间，可为空 |
| collectionTimezone | string | 时间窗口时区，可为空 |
| resultRevealAt | string | 当前用户这份问卷的匹配结果预计公布时间，可为空 |
| totalCount | number | 总题数 |
| answeredCount | number | 已答题数 |
| progressPercent | number | 进度百分比 |
| hasCompletedQuestionnaire | boolean | 当前版本必答题是否完成 |
| questions | Question[] | 问卷题目定义 |
| answers | Answer[] | 当前用户已保存答案 |

`Question` 字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | string | 题号，如 `1`、`15a` |
| order | number | 排序 |
| section | string | 分组标题 |
| title | string | 题目标题 |
| type | string | `single_choice` / `multiple_choice` / `number` / `range` / `scale` / `text` |
| widget | string | 前端建议控件类型 |
| required | boolean | 是否必答 |
| options | Option[] | 选项题可见 |
| min | number | 最小值，可为空 |
| max | number | 最大值，可为空 |
| leftLabel | string | 量表左锚点 |
| rightLabel | string | 量表右锚点 |
| minSelections | number | 多选最少项，可为空 |
| maxSelections | number | 多选最多项，可为空 |
| maxLength | number | 文本最大长度，可为空 |
| placeholder | string | 占位提示，可为空 |
| allowUnlimited | boolean | 范围题是否支持不限 |
| supportsImportance | boolean | 是否支持“设为重要” |

`Answer` 字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| questionId | string | 题号 |
| value | any | 答案值 |
| importance | boolean | 是否标记为重要 |

### 7.2 获取当前问卷进度

- 方法：`GET`
- 路径：`/api/v1/questionnaires/current/progress`
- 鉴权：是

返回字段是 `current` 的精简版，不包含 `questions` 和 `answers`，但会包含 `resultRevealAt`。

### 7.3 保存单题答案

- 方法：`PUT`
- 路径：`/api/v1/questionnaires/current/answers/{questionId}`
- 鉴权：是

请求体：

```json
{
  "value": "男",
  "importance": false
}
```

响应 `data`：

```json
{
  "questionId": "1",
  "value": "男",
  "importance": false,
  "resultRevealAt": "2026-03-27T21:20:00+08:00",
  "answeredCount": 1,
  "progressPercent": 10,
  "hasCompletedQuestionnaire": false
}
```

### 7.4 批量保存答案

- 方法：`PUT`
- 路径：`/api/v1/questionnaires/current/answers`
- 鉴权：是

请求体：

```json
{
  "answers": [
    { "questionId": "1", "value": "男" },
    { "questionId": "16", "value": ["直接说出来的夸奖和关心"] },
    { "questionId": "57", "value": "希望增加 MBTI 维度" }
  ],
  "action": "submit"
}
```

`action` 可选值：

- `save`
- `save_exit`
- `submit`

响应 `data`：

```json
{
  "resultRevealAt": "2026-03-27T21:20:00+08:00",
  "answeredCount": 20,
  "progressPercent": 80,
  "hasCompletedQuestionnaire": true
}
```

### 7.5 不同题型的 `value` 说明

单选题：

```json
"男"
```

多选题：

```json
["做饭烘焙", "旅行"]
```

数值题 / 量表题：

```json
23
```

范围题：

```json
{
  "min": 20,
  "max": 24
}
```

范围题不限：

```json
{
  "unlimited": true
}
```

文本题：

```json
"希望增加 MBTI 维度"
```

### 7.6 问卷开放时间说明

- 如果问卷未到开放时间，保存接口返回 `403`，`message = "问卷暂未开放"`
- 如果问卷已结束收集，保存接口返回 `403`，`message = "问卷收集已结束"`
- 前端可以通过 `collectionStatus` 和 `acceptingResponses` 控制是否允许编辑

## 8. 匹配结果模块

说明：

- 当前前端侧是只读接口
- 匹配执行由后台逻辑触发，不通过前端接口发起
- 若当前用户本轮无结果，接口会返回 `4043`
- 成功匹配结果受个人公布时间控制
- 匹配失败结果立即可见

### 8.1 获取当前匹配结果

- 方法：`GET`
- 路径：`/api/v1/matches/current`
- 鉴权：是

成功匹配示例：

```json
{
  "matchId": 1,
  "cycleId": 1,
  "cycleKey": "2026-w13",
  "cycleTitle": "2026 第 13 周匹配",
  "revealStatus": "revealed",
  "revealAt": "2026-03-27T21:20:00+08:00",
  "resultStatus": "matched",
  "message": "匹配成功",
  "score": 67,
  "reason": "本轮采用极简随机匹配方案，在完成版本 v0.1 问卷的用户中随机为你配对。",
  "partner": {
    "userId": 2,
    "nickname": "用户乙",
    "department": "经济学院",
    "grade": "本科二年级",
    "contacts": [
      {
        "type": "wechat",
        "label": "微信",
        "value": "it_wechat_2"
      }
    ]
  }
}
```

未匹配成功示例：

```json
{
  "matchId": 2,
  "cycleId": 1,
  "cycleKey": "2026-w13",
  "cycleTitle": "2026 第 13 周匹配",
  "revealStatus": "revealed",
  "resultStatus": "failed",
  "message": "本轮未匹配成功"
}
```

说明：

- 这里的 `revealAt` 是当前用户这条结果自己的公布时间，不是整轮统一公布时间
- 如果结果还未公布，接口返回 `4031`

### 8.2 获取匹配历史列表

- 方法：`GET`
- 路径：`/api/v1/matches/history`
- 鉴权：是

查询参数：

- `limit`：可选，默认 `20`，最大 `100`

响应 `data`：

```json
{
  "items": [
    {
      "matchId": 1,
      "cycleId": 1,
      "cycleKey": "2026-w13",
      "cycleTitle": "2026 第 13 周匹配",
      "revealStatus": "revealed",
      "revealAt": "2026-03-27T21:20:00+08:00",
      "resultStatus": "matched",
      "message": "匹配成功",
      "score": 67,
      "reason": "本轮采用极简随机匹配方案，在完成版本 v0.1 问卷的用户中随机为你配对。"
    }
  ]
}
```

说明：

- 尚未公布的结果不会出现在 `history` 里

### 8.3 获取某条匹配结果详情

- 方法：`GET`
- 路径：`/api/v1/matches/{matchId}`
- 鉴权：是

说明：

- 只能查看当前登录用户自己的匹配结果详情
- 返回结构与 `GET /matches/current` 一致
- 若该条结果未到公布时间，返回 `4031`

### 8.4 获取当前匹配周期状态

- 方法：`GET`
- 路径：`/api/v1/matches/cycle/status`
- 鉴权：是

响应 `data`：

```json
{
  "hasCurrentCycle": true,
  "cycleId": 1,
  "cycleKey": "2026-w13",
  "cycleTitle": "2026 第 13 周匹配",
  "revealStatus": "revealed",
  "revealAt": "2026-03-27T21:00:00+08:00",
  "participantCount": 3,
  "matchedUserCount": 2,
  "successRatePercent": 66,
  "hasUserResult": true,
  "userResultStatus": "matched",
  "userRevealAt": "2026-03-27T21:20:00+08:00",
  "canViewUserResult": false
}
```

字段说明：

- `revealAt`：本轮匹配周期的揭晓时间
- `userRevealAt`：当前用户这条结果自己的公布时间
- `canViewUserResult`：当前时刻是否已经可以查看自己的结果

## 9. 接口封闭时间窗说明

以下写接口可能在匹配计算期间被临时关闭：

- `PUT /api/v1/users/me/profile`
- `PUT /api/v1/users/me/contacts/{contactType}`
- `DELETE /api/v1/users/me/contacts/{contactType}`
- `PUT /api/v1/questionnaires/current/answers`
- `PUT /api/v1/questionnaires/current/answers/{questionId}`

命中关闭时间窗时，接口会返回 `4031`。

## 10. 前端联调建议

- 登录成功后请自行保存 `token`，后续请求统一带 `Authorization` 头
- 修改密码成功后旧 token 会失效，前端应跳回登录页
- 问卷页面进入时建议先调 `GET /questionnaires/current`
- 根据 `acceptingResponses` 判断问卷是否允许编辑
- 根据 `resultRevealAt` 展示“预计公布时间”
- 联系方式请优先使用 `contacts` 字段，不要只依赖兼容字段 `contactType/contactValue`
- 匹配结果页建议优先调 `GET /matches/cycle/status`
- 若 `canViewUserResult=false`，建议展示 `userRevealAt` 倒计时，而不是直接请求 `GET /matches/current`
