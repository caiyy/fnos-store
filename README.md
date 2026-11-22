# Cocode Store

面向飞牛 fnOS 的第三方应用商店，用于安装、检测更新和维护
`caiyy/Agent2API_FnOS` 发布的原生应用。

## 站点

GitHub Pages：https://caiyy.github.io/fnos-store/

## 工作方式

- GitHub Actions 每 5 分钟读取 `caiyy/Agent2API_FnOS` 最新 Release。
- 收到打包仓库的发布 `repository_dispatch` 时立即同步，不等下一个定时任务。
- 自动生成根目录 `apps.json`，并提交回本仓库。
- 商店默认每 10 分钟重新读取应用源，自动安装可安全升级的新版本。
- 更新通过飞牛应用中心的数据保留 RPC 执行，不会直接覆盖应用目录。
- 商店本身以 root 运行，只允许受信任的 GitHub Release 作为应用源。

默认应用源：

```text
https://raw.githubusercontent.com/caiyy/fnos-store/main/apps.json
```

默认 Release 下载源：

```text
https://github.com/caiyy/Agent2API_FnOS/releases/download
```

## 安装

从 [Releases](https://github.com/caiyy/fnos-store/releases) 下载对应架构的
`cocode-store_*.fpk`，在飞牛应用中心手动安装。

- Intel / AMD：`cocode-store_<version>_x86.fpk`
- ARM：`cocode-store_<version>_arm.fpk`

安装后通过飞牛桌面打开 **Cocode Store**。普通用户默认不可见，需要使用
NAS 管理员账号访问。

设置中的「自动安装更新」默认开启。关闭后商店仍会检查版本，但只在界面中
显示更新按钮，不执行安装。

## 构建

依赖：

- Go 1.25+
- Node.js 22+
- 官方 `fnpack 1.3.3`

```bash
# macOS / Linux，需要把 fnpack 放到 build/fnpack
./build.sh
```

产物位于：

```text
dist/cocode-store_<version>_x86.fpk
dist/cocode-store_<version>_arm.fpk
```

## 安全边界

- 商店只在应用源中发布白名单内的 `agent2api`。
- 更新包由飞牛应用中心校验并执行，不使用会删除数据的
  `appcenter-cli install-fpk` 路径。
- 商店是特权应用。应只从本仓库的 Release 安装，并保护 GitHub 发布权限。
- 如果飞牛版本不支持安全升级 RPC，商店会拒绝自动更新并提示手动安装 FPK。

## 许可证

本项目基于 [conversun/fnos-store](https://github.com/conversun/fnos-store)
修改，使用 MIT License。第三方应用和商标归各自所有者所有。
