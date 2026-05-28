# 部署说明

当前采用：

- GitHub 负责代码托管与自动测试
- 服务器手动拉取代码并执行部署

不再依赖 GitHub Actions 直接 SSH 登录服务器发布。

## GitHub

仓库内保留的自动化只有 CI：

- 推送到 `main` 时自动执行 `go test ./...`
- Pull Request 也会自动执行测试

工作流文件：

- [ci.yml](/opt/go/src/gitlab.meitu.com/xmqibu/.github/workflows/ci.yml)

## 服务器目录约定

建议线上目录如下：

1. Git 仓库工作目录
   - `/opt/xmqibu-frontier-site/repo`

2. 部署产物目录
   - `/opt/xmqibu-frontier-site`

3. 环境文件
   - `/opt/xmqibu-frontier-site/config/frontier-site.env`

4. 媒体目录
   - `/opt/xmqibu-frontier-site/uploads`

5. 日志目录
   - `/opt/xmqibu-frontier-site/logs`

6. 发布记录目录
   - `/opt/xmqibu-frontier-site/releases`

## 首次准备

服务器上首次执行：

```bash
mkdir -p /opt/xmqibu-frontier-site
cd /opt/xmqibu-frontier-site
git clone https://github.com/yanlc-1/xmqibu.git repo
mkdir -p config uploads logs releases
```

然后准备：

- `/opt/xmqibu-frontier-site/config/frontier-site.env`

内容至少包含：

```bash
PORT=18092
MYSQL_DSN=...
INGEST_TOKEN=...
MEDIA_DIR=/uploads
```

## 手动部署

之后每次发版，只需要在服务器执行：

```bash
cd /opt/xmqibu-frontier-site/repo
./scripts/deploy_from_git.sh
```

这条命令会做这些事：

1. `git fetch`
2. 对齐到 `origin/main`
3. 本机构建 Linux 二进制
4. 复制 `Dockerfile` 到部署目录
5. 调用 `deploy_release.sh`
6. 启动新容器
7. 访问 `/healthz` 做健康检查
8. 更新 Nginx 上游并 reload
9. 在 `releases/` 里写入当前发布记录

## 回滚

当前回滚方式仍然很直接：

1. 找到旧容器 IP
2. 修改 `/usr/local/nginx/conf/vhost/xmqibu-https.conf`
3. 把 `proxy_pass` 指回旧容器
4. reload 宿主机 Nginx

也可以先看：

- `/opt/xmqibu-frontier-site/releases/current`
- `/opt/xmqibu-frontier-site/releases/*.txt`

用来确认当前线上跑的是哪次发布。

## 说明

如果后面你想恢复成“GitHub 自动部署”，也可以，但建议等项目稳定后再做。  
对你现在这个阶段，`GitHub 自动测试 + 服务器手动部署` 更简单，也更容易排错。
