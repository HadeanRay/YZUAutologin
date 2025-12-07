# YZUAutologin 项目文档

## 项目概述

YZUAutologin 是一个专为扬州大学校园网设计的自动登录应用程序，采用 Go 后端和 Web 前端结合的架构，基于 Wails 框架构建。该应用程序通过模拟网页操作实现校园网自动登录，用户数据存储在本地 JSON 文件中，支持开机自启动功能。

### 技术栈

- **后端**: Go 1.21+，使用 Wails v2 框架
- **前端**: HTML/CSS/JavaScript，使用 Sober UI 组件库
- **自动化**: Go-rod 库用于网页操作模拟
- **构建工具**: Vite (前端构建)
- **数据存储**: 本地 JSON 文件 (data.json)

## 项目结构

```
YZUAutologin/
├── main.go              # 主入口文件，应用初始化
├── app.go               # 应用核心逻辑，数据处理和自动启动设置
├── login.go             # 登录逻辑，网页操作模拟实现
├── go.mod               # Go 模块依赖
├── wails.json           # Wails 配置文件
├── frontend/            # 前端资源目录
│   ├── index.html       # 主页面
│   ├── package.json     # 前端依赖配置
│   └── src/
│       ├── main.js      # 前端逻辑
│       └── style.css    # 样式文件
└── build/               # 构建输出目录
```

## 核心功能

1. **自动登录**: 通过 Go-rod 模拟浏览器操作，输入用户名密码并选择运营商进行登录
2. **智能网页检测**: 自动检测校园网登录页面，无需手动输入URL
3. **数据持久化**: 用户配置保存在可执行文件同目录的 data.json 文件中
4. **开机自启动**: 通过修改 Windows 注册表实现开机自启动
5. **UI 界面**: 基于 Sober 组件库的现代化界面
6. **网络状态检测**: 实时检测网络连接状态和认证需求

## 构建和运行

### 环境要求

- Go 1.21+
- Node.js (用于前端开发)
- Wails v2 CLI

### 开发模式

```bash
# 安装依赖
go mod tidy
cd frontend && npm install && cd ..

# 启动开发服务器
wails dev
```

### 构建应用

```bash
# 构建生产版本
wails build
```

### 前端单独开发

```bash
cd frontend
npm run dev    # 开发模式
npm run build  # 构建
```

## 开发约定

### 代码风格

- Go 代码遵循标准 Go 格式化规范
- 前端代码使用模块化 JavaScript
- UI 组件基于 Sober 库，保持一致的视觉风格

### 数据结构

用户配置数据结构 (data.json):
```json
{
  "webindex": "校园网登录页面URL",
  "countindex": "用户名",
  "passwordindex": "密码",
  "operatorindex": "运营商选择(a/b/c/d)",
  "autostartindex": "是否开机自启动(true/false)"
}
```

### 新增功能说明

#### 智能网页检测功能
1. **自动检测原理**: 通过访问多个测试网站（百度、Google、校园官网等），跟踪网络重定向链，自动识别校园网认证页面
2. **检测策略**: 
   - 尝试访问常见网站，观察是否被重定向到认证网关
   - 分析最终URL特征，识别登录页面
   - 支持多种校园网网关地址模式（10.x.x.x, 192.168.x.x等）
3. **使用方式**: 点击"自动检测登录页面"按钮，系统会自动检测并填充登录URL

#### 网络状态检测
1. **连通性测试**: 检测当前网络是否已连接互联网
2. **认证需求检测**: 判断网络是否需要校园网认证
3. **状态显示**: 提供详细的网络状态信息面板

### 运营商代码

- a: 校园网
- b: 联通
- c: 移动
- d: 电信

## 已知问题

1. 重复隐藏、显示应用会导致系统托盘卡住，无法打开菜单
2. 开机自启动设置可能被 Windows Defender 阻止导致程序崩溃
3. 目前主要针对 Windows 平台优化，Mac 兼容性未经测试

## 扩展开发

### 添加新功能

1. 后端功能在 `app.go` 或 `login.go` 中实现
2. 前端界面修改在 `frontend/index.html` 和 `frontend/src/main.js` 中
3. 新增依赖需在 `go.mod` 和 `frontend/package.json` 中分别声明

### 调试技巧

- 查看控制台输出进行前端调试
- 使用 Go 的 `fmt.Println` 进行后端调试
- 检查 `data.json` 文件确认配置保存情况

## 许可证

请参考项目根目录的 LICENSE 文件了解许可证信息。