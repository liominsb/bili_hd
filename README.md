# Go Bili 后端

Go Bili 的 REST API 服务，使用 Go 和 Gin 开发。服务负责用户认证、视频与评论管理、互动数据、播放历史、文件上传及 GitHub OAuth，并使用 MySQL 持久化数据、Redis 缓存和 RabbitMQ 处理异步任务。

## 功能

- 用户注册、登录、JWT 双令牌刷新
- GitHub OAuth 登录
- 用户资料和密码管理
- 视频列表、详情、搜索及增删改
- 点赞、收藏、关注和评论
- 播放历史与进度上报
- Redis 视频缓存、播放量聚合与定时回写
- 文件上传至腾讯云 COS
- RabbitMQ 异步任务消费
- 启动时通过 GORM AutoMigrate 同步表结构
- HTTP 服务和后台任务优雅退出

## 技术栈

- Go 1.25
- Gin
- GORM + MySQL 8
- Redis
- RabbitMQ
- JWT
- Viper
- 腾讯云 COS SDK

## 环境要求

- Go 1.25+
- MySQL 8
- Redis
- RabbitMQ
- FFmpeg（处理本地视频文件；Docker 镜像已包含）
- 腾讯云 COS 存储桶及访问密钥（使用上传功能时必需）

服务启动时会立即连接 MySQL、Redis 和 RabbitMQ，任一依赖不可用都会导致启动失败。

## 本地启动

### 1. 准备依赖服务

创建数据库，字符集建议使用 `utf8mb4`：

```sql
CREATE DATABASE go_bili CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

同时确保 Redis 和 RabbitMQ 已启动。默认开发地址可参考 `config/config.example.yml`。

### 2. 创建配置

复制配置模板：

```bash
cp config/config.example.yml config/config.yml
```

PowerShell：

```powershell
Copy-Item config/config.example.yml config/config.yml
```

然后编辑 `config/config.yml`：

```yaml
app:
  name: Go Bili
  port: :3000

database:
  dsn: "root:password@tcp(127.0.0.1:3306)/go_bili?charset=utf8mb4&parseTime=True&loc=Local"
  MaxIdleConns: 10
  MaxOpenConns: 100
  Addr: "127.0.0.1:6379"
  Password: ""
  SubSwitch: false

JWT:
  Key: "请替换为足够长的随机字符串"

RabbitMQ:
  Url: "amqp://guest:guest@127.0.0.1:5672/"

GitHub:
  ClientID: ""
  ClientSecret: ""
  RedirectURI: "http://localhost:3000/api/auth/github/callback"
  FrontendURL: "http://localhost:5173/oauth/callback"

Cos:
  SecretID: ""
  SecretKey: ""
  BucketURL: "https://<bucket>-<appid>.cos.<region>.myqcloud.com"
```

`config/config.yml` 已被 Git 忽略，请勿提交真实密码或密钥。以下环境变量会覆盖配置文件中的对应密钥：

- `GITHUB_CLIENT_SECRET`
- `COS_SECRET_ID`
- `COS_SECRET_KEY`

### 3. 运行服务

```bash
go mod download
go run .
```

默认监听 <http://localhost:3000>。健康检查：

```bash
curl http://localhost:3000/ping
```

正常响应为 `pong`。

## API 概览

需要登录的接口必须携带请求头：

```http
Authorization: Bearer <access-token>
```

公开接口也会解析可选 Token，以便返回当前用户的点赞、收藏或关注状态。

| 方法 | 路径 | 说明 | 鉴权 |
| --- | --- | --- | --- |
| `POST` | `/api/auth/register` | 注册 | 否 |
| `POST` | `/api/auth/login` | 登录 | 否 |
| `POST` | `/api/auth/refreshTokens` | 刷新双令牌 | 否 |
| `GET` | `/api/auth/github/login` | 发起 GitHub 登录 | 否 |
| `GET` | `/api/auth/github/callback` | GitHub 回调 | 否 |
| `GET` | `/api/v1/videos` | 视频列表 | 否 |
| `GET` | `/api/v1/videos/:id` | 视频详情 | 否 |
| `GET` | `/api/v1/videos/search` | 按标题搜索 | 否 |
| `POST` | `/api/v1/videos` | 发布视频记录 | 是 |
| `PUT/DELETE` | `/api/v1/videos/:id` | 修改或删除视频 | 是 |
| `POST` | `/api/v1/upload` | 上传文件 | 是 |
| `GET/POST` | `/api/v1/videos/:id/comments` | 评论列表或发表评论 | 写操作需要 |
| `PUT/DELETE` | `/api/v1/comments/:id` | 修改或删除评论 | 是 |
| `PUT` | `/api/v1/videos/:id/like` | 切换点赞状态 | 是 |
| `PUT/DELETE` | `/api/v1/videos/:id/favorite` | 收藏或取消收藏 | 是 |
| `GET` | `/api/v1/users/me/favorites` | 我的收藏 | 是 |
| `PUT/DELETE` | `/api/v1/users/:id/follow` | 关注或取关用户 | 是 |
| `GET` | `/api/v1/users/:id/followers` | 粉丝列表 | 否 |
| `GET` | `/api/v1/users/:id/following` | 关注列表 | 否 |
| `GET/PUT` | `/api/v1/users/me` | 获取或更新本人资料 | 是 |
| `PUT` | `/api/v1/users/me/psw` | 修改密码 | 是 |
| `PUT/DELETE` | `/api/v1/videos/:id/history` | 上报或删除播放历史 | 是 |
| `GET/DELETE` | `/api/v1/users/me/history` | 查询或清空播放历史 | 是 |

列表和搜索接口使用 `offset`、`limit` 查询参数分页。

## 目录结构

```text
go_bili/
├─ api/
│  ├─ controllers/      # HTTP 参数解析与响应
│  ├─ service/          # 业务逻辑
│  └─ repository/       # 数据访问
├─ cmd/bulkupload/      # 批量上传命令行工具
├─ config/              # 配置加载及基础设施初始化
├─ docs/                # 设计与性能优化记录
├─ global/              # 数据库、缓存等共享连接
├─ middlewares/         # JWT 鉴权中间件
├─ models/              # GORM 模型和 DTO
├─ router/              # 路由与依赖注入
├─ test/                # 性能及并发相关测试
├─ utils/               # JWT、消息队列、COS 与通用工具
├─ Dockerfile
└─ main.go
```

代码调用方向为 `controller -> service -> repository`，依赖在 `router/wire.go` 中组装。

## 测试

```bash
go test ./...
```

运行指定测试或基准测试：

```bash
go test ./utils -run TestName -v
go test ./test -bench=. -benchmem
```

## Docker 部署

完整编排文件位于仓库上级目录。生产部署前先构建前端，再启动服务：

```bash
cd ../go_bili_qd
npm ci
npm run build

cd ..
docker compose -f docker-compose.prod.yml up -d --build
```

生产编排会启动 MySQL、Redis、RabbitMQ、后端与 Nginx，并持久化数据库、队列和上传目录。

上线前务必完成以下检查：

- 替换 `config/config.prod.yml` 中的 MySQL 密码和 JWT Key
- 配置正确的 GitHub OAuth 回调地址
- 通过容器 `environment` 显式传入 `GITHUB_CLIENT_SECRET`、`COS_SECRET_ID`、`COS_SECRET_KEY`
- 确认 COS 存储桶的访问权限和跨域规则
- 配置 HTTPS、域名和防火墙

查看服务状态与日志：

```bash
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs -f server
```

停止服务：

```bash
docker compose -f docker-compose.prod.yml down
```
