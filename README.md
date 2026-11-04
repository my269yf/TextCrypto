# 本机加密终端 · Wails 桌面版

把原来的 `text-crypto.html` 包装成单文件 Windows 桌面程序：不需要浏览器标签页、
不需要 miniserve、不需要端口，关闭窗口即退出进程，`account.json` 自动生成在
exe 同目录。

## 一、需要安装的东西

| 依赖 | 说明 | 获取方式 |
|---|---|---|
| **Go 1.21+** | 编译 Go 后端 | `winget install --id GoLang.Go -e`，或从 https://go.dev/dl/ 下 MSI |
| **Wails CLI** | 打包工具，用 Go 安装 | 装好 Go 后执行 `go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.0` |
| **WebView2 运行时** | Win10/11 基本自带，用于渲染界面 | 若 `wails doctor` 报缺失：https://developer.microsoft.com/microsoft-edge/webview2/ 装 Evergreen Runtime |
| Node / npm | **本项目不需要**（前端是静态文件，构建时加 `-s` 跳过） | 已装则忽略 |
| Git | 非必需（Go 走模块代理下载依赖） | 可选 |

安装完 Go 后确认：

```bat
go version
```

并检查 `%USERPROFILE%\go\bin` 已加入 PATH（Wails 会装到这里），否则会提示找不到 `wails` 命令。
装完 Wails CLI 后执行环境自检：

```bat
wails doctor
```

输出里 `WebView2 Installed` 和 `Go Version` 正常即可。

## 二、构建

在本目录双击 `build.bat`，或命令行执行：

```bat
cd /d "路径\wails-app"
wails build -s
```

- `-s` 表示跳过前端构建（本项目前端就是 `frontend/dist/index.html` 静态文件）
- 首次构建会下载 Go 依赖，需要联网，耗时几分钟
- 产物：`build\bin\TextCrypto.exe`（约 8–12 MB）

## 三、运行与数据

- 双击 `TextCrypto.exe` 即可，关闭窗口进程即结束（不再有后台残留）
- **账户只保存在用户目录**：`%APPDATA%\TextCrypto\account.json`，不再依赖浏览器的 IndexedDB
  - 好处：exe 可以随便移动、替换、重新编译，账户始终在这台电脑上
  - 删掉这个文件后再启动，会回到「初始化账户」界面，不会自动恢复
- `account.migrated` 是迁移标记（旧版本把账户存在 WebView2 里时用来自动导出一次），
  下次启动会自动重建；删掉它不影响已存在账户的正常登录
- 旧版本放在 exe 旁边的那两个文件，会在首次启动时自动搬到用户目录，不需要手工处理
- 想沿用网页版数据：把网页版目录里的 `account.json` 复制到 `%APPDATA%\TextCrypto\` 即可
- `build.bat` 特意不使用 `-clean`，避免清掉构建目录里的其他文件

## 四、开发调试

```bat
wails dev -s -assetdir frontend/dist
```

会启动热重载窗口；注意 dev 模式的 exe 位于临时构建目录，`account.json` 也会落在那里，
属于正常现象。

## 五、代码结构

```
wails-app/
├─ main.go                   窗口、尺寸、资源嵌入
├─ app.go                    读写 account.json 的 Go 绑定（LoadAccount / SaveAccount / 迁移标记）
├─ wails.json                项目配置
├─ build/appicon.png         应用图标
├─ build.bat                 一键构建
├─ recovered/                从旧版 WebView2 存储中抢救出来的账户快照（备用）
└─ frontend/dist/index.html  界面（由网页版 text-crypto.html 生成，仅替换了文件读写方式）
```

前端检测到 `window.go.main.App` 时走原生读写，否则回退到浏览器版的上传接口，
所以这份 `index.html` 直接丢进浏览器也能跑。

### 修改界面后如何同步

桌面版用的是 `frontend/dist/index.html` 这份副本。改完根目录的 `text-crypto.html`
后，复制过来，并只把 CSP 那一行改成允许同源资源（默认那份是 `script-src 'unsafe-inline'`）：

```
default-src 'none'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';
img-src 'self' data:; connect-src 'self'; font-src 'self' data:; ...
```
