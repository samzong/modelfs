# Release Guide

本项目使用 [GoReleaser](https://goreleaser.com/) 进行版本发布管理和多平台 Docker 镜像构建。

## 发布流程

### 1. 创建 Git Tag

```bash
# 创建并推送版本 tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### 2. 自动触发发布

当推送 tag 到 GitHub 后，`.github/workflows/release.yml` 会自动触发，执行以下操作：

- 使用 GoReleaser 构建 Linux 多平台二进制文件（linux/amd64, linux/arm64）
- 构建多平台 Docker 镜像（linux/amd64, linux/arm64）
- 推送镜像到 `ghcr.io/samzong/modelfs`
- 创建 GitHub Release 并上传构建产物

### 3. 手动触发发布

也可以通过 GitHub Actions 界面手动触发发布：

1. 进入 Actions -> Release
2. 点击 "Run workflow"
3. 输入版本 tag（如 `v1.0.0`）

## 镜像使用

发布后，可以从以下位置拉取镜像：

```bash
# 特定版本
docker pull ghcr.io/samzong/modelfs:v1.0.0

# 最新版本
docker pull ghcr.io/samzong/modelfs:latest
```

## 本地测试

本地测试 GoReleaser 配置和构建流程：

```bash
GITHUB_REPOSITORY_OWNER="samzong" goreleaser release --snapshot --clean
```

这个命令会：

- 构建 Linux 多平台二进制文件（linux/amd64, linux/arm64）
- 构建多平台 Docker 镜像（linux/amd64, linux/arm64）
- 生成归档文件和校验和
- 验证 `.goreleaser.yml` 配置

## 配置说明

- `.goreleaser.yml`: GoReleaser 配置文件

  - 构建配置：支持多平台二进制构建
  - Docker 配置：使用 dockers_v2 构建多平台镜像
  - Release 配置：自动创建 GitHub Release

- `.github/workflows/release.yml`: 发布工作流

  - 触发条件：推送 tag（v\*）或手动触发
  - 执行步骤：设置环境、构建、发布

- `.github/workflows/ci.yml`: CI 工作流
  - 触发条件：推送到 main 分支或 PR
  - 执行步骤：测试、lint、构建验证

## 常见问题

### 1. 如何验证多平台镜像构建成功

在 GitHub Actions 中构建并推送后，可以验证镜像支持多个平台：

```bash
# 检查镜像支持的平台
docker buildx imagetools inspect ghcr.io/samzong/modelfs:latest

# 或者检查特定版本
docker buildx imagetools inspect ghcr.io/samzong/modelfs:v1.0.0
```

输出应该显示 `linux/amd64` 和 `linux/arm64` 两个平台。
