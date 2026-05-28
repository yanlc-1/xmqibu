# GitHub CI/CD 部署说明

这套站点现在支持通过 GitHub Actions 自动测试并发布到线上服务器，不再依赖手工 `scp + ssh + 改 nginx`。

## 流程

1. 推送到 `main`
2. GitHub Actions 执行 `go test ./...`
3. 构建 Linux 发布二进制 `frontier-site`
4. 通过 SSH 上传以下文件到服务器
   - `frontier-site`
   - `Dockerfile`
   - `scripts/deploy_release.sh`
5. 服务器侧脚本完成这些动作
   - 用上传的二进制构建新镜像
   - 启动一个新的旁路容器
   - 访问 `/healthz` 做健康检查
   - 更新 `/usr/local/nginx/conf/vhost/xmqibu-https.conf`
   - reload 宿主机 Nginx
   - 用 `xmqibu.com` 和 `www.xmqibu.com` 做 HTTPS 验证

## 需要的 GitHub Secrets

在 GitHub 仓库里配置这些 secrets：

1. `DEPLOY_HOST`
   - 线上服务器地址
   - 当前值示例：`111.231.204.222`

2. `DEPLOY_PORT`
   - SSH 端口
   - 当前值示例：`34185`

3. `DEPLOY_USER`
   - SSH 用户
   - 当前值示例：`root`

4. `DEPLOY_SSH_KEY`
   - 对应服务器可登录私钥
   - 建议使用专门的 deploy key，不要复用个人常用私钥

## 服务器预置条件

服务器上需要提前存在这些内容：

1. `/opt/xmqibu-frontier-site/config/frontier-site.env`
   - 包含 `PORT`、`MYSQL_DSN`、`INGEST_TOKEN`

2. `/opt/xmqibu-frontier-site/uploads`
   - 本地媒体目录

3. `docker`
   - 用于启动新版本容器

4. `nginx-proxy` 网络
   - 当前容器依赖该 Docker network

5. `/usr/local/nginx/conf/vhost/xmqibu-https.conf`
   - 宿主机 HTTPS 入口配置

## 健康检查

站点新增了：

- `GET /healthz`

成功时返回：

```text
ok
```

GitHub Actions 只在新容器通过 `/healthz` 后才会切换 Nginx 上游。

## 回滚

这套脚本默认保留旧容器，不会自动删除历史版本。回滚时可以：

1. 找到旧容器 IP
2. 把 `/usr/local/nginx/conf/vhost/xmqibu-https.conf` 的 `proxy_pass` 改回旧 IP
3. reload 宿主机 Nginx

## 当前工作流文件

- [deploy.yml](/opt/go/src/gitlab.meitu.com/xmqibu/.github/workflows/deploy.yml)
- [deploy_release.sh](/opt/go/src/gitlab.meitu.com/xmqibu/scripts/deploy_release.sh)
