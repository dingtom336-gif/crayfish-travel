# 飞猪内部 API 对接需求（含降级方案）

> 小龙虾旅行项目 - 反向竞价旅行平台
> 核心流程：搜套餐 -> 锁库存 -> 支付 -> 退改
> 日期：2026-03-30

---

## API 1: 套餐报价查询

**理想方案：机+酒打包查询接口**
- 一个接口传入目的地+日期+人数，返回组合好的套餐列表
- 每个套餐包含酒店+机票+门票的打包价

```
POST /package/search

Request:
{
  "destination": "三亚",
  "check_in": "2026-07-15",
  "check_out": "2026-07-20",
  "adults": 2,
  "children": 1,
  "budget_max_cents": 800000,
  "sort_by": "price_asc",
  "limit": 10
}

Response:
{
  "packages": [{
    "package_id": "FG20260715SANYA001",
    "hotel_name": "...",
    "flight_info": {...},
    "total_price_cents": 720000,
    "base_price_cents": 710000,
    "inclusions": [...],
    "highlights": [...],
    "star_rating": 4.8,
    "review_count": 1200,
    "image_url": "...",
    "supplier_name": "...",
    "cancellation_policy": {...}
  }]
}
```

**替代方案 A：分别查酒店+机票，我们侧组合**
- 需要两个接口：`hotel/search` + `flight/search`
- 我们在 bidding 模块做组合逻辑：同目的地同日期的酒店+机票配对，算打包价
- 额外需要确认：酒店和机票是否有统一的 destination_id 可以关联？日期格式是否一致？

**替代方案 B：用 FlyAI CLI 做搜索层**
- 用已安装的 `flyai search-hotels` + `flyai search-flight` 做初始筛选
- 再通过内部接口查精确价格和库存
- 缺点：CLI 是消费端数据，价格可能与 B2B 不一致

**最低可行：保持 MockSupplier + 人工报价后台**
- 运营人员手动在后台录入真实套餐
- 适合冷启动阶段，用户量小时可行

---

## API 2: 套餐锁定/预留

**理想方案：飞猪侧有库存锁定接口**
- 锁定 15 分钟，超时自动释放，双方都有释放能力

```
POST /package/lock

Request:
{
  "package_id": "FG20260715SANYA001",
  "session_id": "uuid",
  "lock_duration_seconds": 900,
  "trace_id": "uuid"
}

Response:
{
  "lock_id": "...",
  "expires_at": "2026-07-15T10:15:00Z",
  "status": "LOCKED"
}
```

**替代方案 A：飞猪有"预下单"接口**
- 类似购物车/待支付订单，创建后有一个支付窗口期
- 我们将这个窗口期映射为锁定期
- 需要确认：预下单的有效期是多久？能否自定义？超时后库存自动回滚吗？

**替代方案 B：飞猪有"库存查询"但无锁定**
- 展示套餐时标注"实时库存，先到先得"
- 用户选择后直接进入支付，跳过 15 分钟锁定
- 我们的 Saga 简化为：查库存 -> 直接支付 -> 建单
- 改动：lock 模块改为"软锁"（Redis 侧标记，不锁供应商库存），UI 倒计时改为"建议尽快支付"

**最低可行：完全去掉供应商侧锁定**
- 只在我们侧做 session 级别的软锁（Redis 15min TTL）
- 支付时实时检查库存，库存不足则提示用户换方案
- Saga 补偿只处理支付失败场景

---

## API 3: 套餐/商品详情

**理想方案：通过 package_id 查打包详情**
- 行程安排、酒店详情、航班信息、退改政策一次返回

```
GET /package/detail?package_id=FG20260715SANYA001
```

**替代方案 A：通过 hotel_id + flight_id 分别查**
- 如果没有打包详情，用酒店详情接口 + 机票详情接口分别查
- 需要确认：是否有 `hotel/detail?hotel_id=xxx` 和 `flight/detail?flight_id=xxx`？

**替代方案 B：用 FlyAI 补充详情**
- 搜索接口返回的基础信息（名称、价格、星级）已足够展示卡片
- 详情页用 FlyAI 的 search-hotels 补充酒店图片、设施、评分
- 行程安排由我们根据机票+酒店信息自动生成模板

**最低可行：只展示搜索结果中的字段**
- 不做独立详情页，方案卡片上直接展示所有信息
- 用户点击"选择"直接进锁定流程

---

## API 4: 退改政策查询

**理想方案：精确到套餐级别的阶梯退款规则**
- 输入 package_id + 购买日期，返回各时间段的退款比例

```
GET /package/cancellation-policy?package_id=FG20260715SANYA001&order_date=2026-07-10

Response:
{
  "rules": [
    {"days_before": 7, "refundable_percentage": 100},
    {"days_before": 3, "refundable_percentage": 80},
    {"days_before": 1, "refundable_percentage": 50},
    {"days_before": 0, "refundable_percentage": 0}
  ]
}
```

**替代方案 A：飞猪有酒店级别的退改政策**
- 大部分退改成本在酒店侧，机票退改相对标准
- 用酒店退改政策 + 航司通用退改规则组合计算
- 需要确认：酒店详情接口是否包含 `cancellation_policy` 字段？

**替代方案 B：飞猪提供退改费率表（非实时接口）**
- 不是接口调用，而是提供一份退改费率配置表（按酒店星级、提前天数等维度）
- 我们写入 `date_configs` 表，hedge_calculator 从本地查
- 定期同步更新

**最低可行：固定退改比例**
- 保持当前 mock 策略：提前 7 天以上 80% 回收，7 天内 50%，当天 0%
- 资金池兜底补差
- 后续根据实际运营数据调整

---

## API 5: 退款发起

**理想方案：我们调接口发起退款，飞猪处理后回调通知**

```
POST /order/refund

Request:
{
  "order_id": "...",
  "package_id": "...",
  "reason": "用户主动取消",
  "trace_id": "uuid"
}

Response:
{
  "refund_id": "...",
  "refund_amount_cents": 720000,
  "refund_status": "processing",
  "estimated_completion": "2026-07-15T12:00:00Z"
}
```

**替代方案 A：飞猪有取消订单接口，退款自动触发**
- 我们调"取消订单"，飞猪自动处理退款到用户支付宝
- 需要确认：取消后退款金额是按飞猪政策还是我们可以指定？如果飞猪按政策只退 80%，剩下 20% 由我们资金池补给用户

**替代方案 B：退款走支付宝侧，不走飞猪**
- 飞猪侧只做订单取消
- 退款金额由我们通过支付宝转账/退款接口直接退给用户
- 供应商回收的部分再通过内部结算回到我们账户
- 适合场景："全额无忧退"是我们的产品承诺，不依赖供应商退款金额

**最低可行：人工处理退款**
- 系统标记退款请求，运营人员在飞猪商家后台手动操作
- 适合早期用户量少的阶段

---

## 网关/安全统一问题

| # | 问题 | 影响 | 选项 A | 选项 B |
|---|------|------|--------|--------|
| 1 | 调用方式 | 架构 | 内网 HSF/Dubbo（需 Dubbo-go SDK 或 Java sidecar） | HTTP Gateway（直接 HTTP 调用，简单） |
| 2 | 外网访问 | 部署 | IP 白名单（提供腾讯云 IP `150.158.192.237`） | VPN 隧道回阿里云内网 |
| 3 | 如果必须走内网 | 架构 | 提供一个 HTTP 网关代理（Go 后端直接调） | 考虑将后端迁移到阿里云 ECS |
| 4 | 签名方式 | 安全 | HMAC-SHA256 | RSA / 内部 Token |
| 5 | 回调通知 | 退款流程 | HTTP Webhook（提供公网回调 URL） | RocketMQ/Kafka（需要内网环境） |

---

## 给研发同事的一句话版本

> 我们做一个反向竞价旅行平台，核心流程是：**搜套餐 -> 锁库存 -> 支付 -> 退改**。理想情况需要 5 个接口（套餐查询、锁定、详情、退改政策、退款）。如果没有打包套餐接口，能分别给酒店查询+机票查询也行，我们自己组合。锁定接口如果没有，预下单接口也可以。每个接口都有三级降级方案。最关键要确认的是：**调用走内网还是外网 HTTP Gateway，以及鉴权方式**，这决定了我们的部署架构。
