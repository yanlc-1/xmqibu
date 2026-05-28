# 内容写入接口文档

本文档用于对接“科技前沿内容站”的机器写入接口，供定时任务、抓取任务或内容生产任务调用。

## 概览

- 服务端鉴权方式：`Bearer Token`
- 内容类型：
  - 长文：`POST /api/ingest/articles`
  - 短讯：`POST /api/ingest/briefs`
- 内容发布策略：写入成功后直接发布
- 幂等规则：以 `external_id` 去重；相同内容重复推送不会重复入库

## 基础信息

- Base URL

```text
http://<host>:<port>
```

- 请求头

```http
Authorization: Bearer <INGEST_TOKEN>
Content-Type: application/json
```

## 栏目标签约定

首页固定按以下标签归类，请在 `tags` 中传入对应 slug：

| 标签 slug | 栏目名称 | 说明 |
|---|---|---|
| `overseas-ai` | 海外AI公司动态 | 海外 AI 公司发布、融资、合作、模型能力更新 |
| `china-ai` | 国内AI公司动态 | 国内 AI 公司动态、产品进展、组织变化 |
| `github-top` | GitHub 近3天 Star 增长 Top5 | 开源项目榜单、工具趋势、热度增长 |
| `ai-products` | AI应用与产品动态 | AI 应用发布、功能更新、产品迭代 |

一个内容可以带多个标签，但建议至少带 1 个上述首页栏目标签。

## 1. 写入长文

- 接口地址

```http
POST /api/ingest/articles
```

- 请求体字段

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `external_id` | string | 是 | 外部唯一 ID，用于幂等去重 |
| `title` | string | 是 | 文章标题 |
| `slug` | string | 否 | 文章 URL slug；不传则服务端自动生成 |
| `summary` | string | 是 | 文章摘要 |
| `cover_image` | string | 否 | 封面图 URL |
| `body` | string | 是 | 文章正文，当前按纯文本/普通字符串存储 |
| `author` | string | 是 | 作者名或机器人名 |
| `source_links` | string[] | 否 | 来源链接列表 |
| `tags` | string[] | 否 | 标签 slug 列表 |
| `published_at` | string | 是 | 发布时间，RFC3339 格式 |
| `featured` | boolean | 否 | 是否作为重点文章候选 |

- 请求示例

```json
{
  "external_id": "article-20260520-001",
  "title": "OpenAI 新模型能力观察",
  "summary": "对最新能力和产品信号的快速整理。",
  "cover_image": "https://example.com/cover.jpg",
  "body": "这里是正文内容。",
  "author": "Automation",
  "source_links": [
    "https://example.com/source-1",
    "https://example.com/source-2"
  ],
  "tags": [
    "overseas-ai"
  ],
  "published_at": "2026-05-20T10:00:00+08:00",
  "featured": true
}
```

如果站点启用了本地媒体目录，`cover_image` 也可以填写站内路径，例如：

```json
{
  "cover_image": "/media/github-top/top8-cover.png"
}
```

推荐流程：

1. 先把图片保存到 `MEDIA_DIR`
2. 拿到站内 URL，例如 `/media/...`
3. 再把这个 URL 写入 `cover_image`

- 成功响应

首次写入成功：

```http
201 Created
Content-Type: application/json
```

```json
{
  "created": true,
  "slug": "openai-new-model-observation"
}
```

重复写入同一 `external_id`：

```http
200 OK
Content-Type: application/json
```

```json
{
  "created": false,
  "slug": "openai-new-model-observation"
}
```

## 2. 写入短讯

- 接口地址

```http
POST /api/ingest/briefs
```

- 请求体字段

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `external_id` | string | 是 | 外部唯一 ID，用于幂等去重 |
| `title` | string | 是 | 短讯标题 |
| `slug` | string | 否 | 短讯 URL slug；不传则服务端自动生成 |
| `summary` | string | 是 | 短讯摘要 |
| `external_url` | string | 否 | 外部原文链接 |
| `importance` | int | 否 | 重要级，建议 1-5 |
| `source_links` | string[] | 否 | 来源链接列表 |
| `tags` | string[] | 否 | 标签 slug 列表 |
| `published_at` | string | 是 | 发布时间，RFC3339 格式 |

- 请求示例

```json
{
  "external_id": "brief-20260520-001",
  "title": "GitHub 热度增长项目速览",
  "summary": "近 3 天 Star 增长最快的项目观察。",
  "external_url": "https://github.com/trending",
  "importance": 5,
  "source_links": [
    "https://github.com/trending"
  ],
  "tags": [
    "github-top"
  ],
  "published_at": "2026-05-20T10:30:00+08:00"
}
```

- 成功响应

首次写入成功：

```http
201 Created
Content-Type: application/json
```

```json
{
  "created": true,
  "slug": "github-trending-top-projects"
}
```

重复写入同一 `external_id`：

```http
200 OK
Content-Type: application/json
```

```json
{
  "created": false,
  "slug": "github-trending-top-projects"
}
```

## 错误响应

### 1. 鉴权失败

```http
401 Unauthorized
Content-Type: text/plain
```

响应体示例：

```text
unauthorized
```

### 2. 请求参数不合法

```http
400 Bad Request
Content-Type: text/plain
```

响应体示例：

```text
summary is required
```

说明：

- 缺少必填字段会返回 `400`
- JSON 中包含服务端未定义字段，也会返回 `400`
- `published_at` 必须是可解析的 RFC3339 时间格式

### 3. 服务端处理失败

```http
500 Internal Server Error
Content-Type: text/plain
```

响应体示例：

```text
failed to ingest article
```

或：

```text
failed to ingest brief
```

## 对接建议

- `external_id` 一定要稳定，不要每次重试都变化，否则无法幂等
- 如果上游内容本身没有 slug，可以不传，由服务端根据标题自动生成
- `tags` 建议统一使用英文 slug，不要混用中文名称
- 如果一条内容属于首页固定栏目，请务必带上对应栏目标签
- `source_links` 建议尽量保留原始来源，方便后续页面展示和溯源

## curl 示例

### 写入长文

```bash
curl -X POST 'http://127.0.0.1:18091/api/ingest/articles' \
  -H 'Authorization: Bearer temporary-frontier-token' \
  -H 'Content-Type: application/json' \
  --data '{
    "external_id": "article-001",
    "title": "Perplexity 新能力观察",
    "summary": "一篇关于海外 AI 公司动态的文章。",
    "body": "正文内容",
    "author": "Automation",
    "tags": ["overseas-ai"],
    "published_at": "2026-05-20T10:00:00+08:00",
    "featured": true
  }'
```

### 写入短讯

```bash
curl -X POST 'http://127.0.0.1:18091/api/ingest/briefs' \
  -H 'Authorization: Bearer temporary-frontier-token' \
  -H 'Content-Type: application/json' \
  --data '{
    "external_id": "brief-001",
    "title": "GitHub 热度增长项目观察",
    "summary": "这是榜单型短讯。",
    "external_url": "https://github.com/trending",
    "importance": 5,
    "tags": ["github-top"],
    "published_at": "2026-05-20T10:00:00+08:00"
  }'
```
