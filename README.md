# 一、软件基本信息（Go Web 专用版）

**软件名称**：实验器材借还管理系统
**英文名称**：Laboratory Equipment Borrowing Management System
**软件简称**：LabEquip-BMS
**版本号**：V1.0
**软件类型**：Web 应用软件

**开发语言**：Go
**运行环境**：

* 操作系统：Linux / Windows / macOS
* Web 服务器：Nginx
* 运行框架：Gin
* 数据库：MySQL
* 浏览器：Chrome / Edge

---

# 二、软件功能简介（软著标准写法）

> ⚠️ 这一段你可以**原样使用**

> 本软件是一套基于 Go 语言开发的 Web 实验器材借还管理系统，主要用于高校实验室教学场景中实验器材的借用、归还及库存管理。系统通过 Web 方式提供统一访问入口，支持实验器材信息维护、借还申请、审批管理、归还登记及借还记录查询等功能，实现实验器材借还流程的信息化管理，提高实验教学管理效率，减少人工统计和管理错误。

---

# 三、系统整体架构说明（Go Web 很加分）

**系统架构：**

* 前端：HTML + CSS + JavaScript
* 后端：Go（Gin Web 框架）
* 数据层：MySQL
* 架构模式：前后端分离 / MVC 架构

**架构特点说明（申报可用）：**

> 系统采用分层架构设计，将表现层、业务逻辑层和数据访问层进行解耦，提升系统的可维护性和扩展性。

---

# 四、主要功能模块（审核重点）

## 1️⃣ 用户管理模块

* 支持学生、助教、管理员三类角色
* 用户登录、退出及身份认证
* 不同角色具备不同操作权限

## 2️⃣ 实验器材管理模块

* 实验器材基础信息维护（编号、名称、类别、库存数量）
* 器材状态管理（可借用、停用）
* 支持器材信息查询与更新

## 3️⃣ 借用申请模块

* 学生通过 Web 页面提交借用申请
* 填写借用数量、借用时间和归还时间
* 系统自动校验库存可用情况

## 4️⃣ 借用审批模块

* 助教或管理员对借用申请进行审核
* 支持审批通过或驳回
* 审批结果实时记录并可追溯

## 5️⃣ 归还管理模块

* 助教登记实验器材归还情况
* 支持正常归还和超期归还标记
* 系统自动更新库存数量

## 6️⃣ 借还记录查询模块

* 按用户、器材、时间条件查询借还记录
* 借还记录长期保存
* 支持实验室管理统计分析

---

# 五、技术特点与“轻创新点”（软著够用）

### 技术特点

* 基于 Go 语言开发，具备高并发和高性能特点
* 使用 RESTful 接口规范，便于系统扩展
* 采用权限控制机制，保障系统数据安全

### 创新点（软著友好写法）

* 将实验器材借还流程抽象为状态流转模型，实现借还全过程管理
* 借用审批与库存校验联动，减少人工管理失误
* 通过 Web 系统统一管理实验器材数据，提高教学辅助效率

---

# 六、助教身份合理性说明（非常关键）

> 这一段可以直接放在【开发目的】里

> 本软件由实验课程助教在教学辅助管理过程中设计开发，针对实验器材借还流程分散、管理效率低的问题，结合实验教学实际需求进行功能设计，具有明确的教学辅助应用价值和实用性。

---

# 七、代码结构示例（为后续"代码页"做准备）

> ⚠️ 软著不审代码质量，只看"像不像系统"

```text
LabEquip-BMS/
├── server/                    # 后端服务
│   ├── api/                   # API 入口
│   │   └── main.go           # 主程序
│   └── _config/              # 配置文件
│       └── conf.yaml         # 系统配置
├── apps/                      # 业务模块
│   ├── basic/                # 基础模块(用户/实验室)
│   ├── bms/                  # 借还管理模块
│   └── common/               # 公共模块
├── framework/                 # 框架层
│   ├── web/                  # Web 框架
│   ├── orm/                  # 数据访问
│   └── config/               # 配置管理
├── defines/                   # 实体定义
│   ├── entity/               # 数据实体
│   └── constant/             # 枚举常量
├── frontend/                  # 前端项目
│   ├── src/                  # 源代码
│   │   ├── pages/            # 页面组件
│   │   ├── components/       # 通用组件
│   │   └── api/              # API 接口
│   └── package.json          # 依赖配置
├── nginx/                     # Nginx 配置
├── database_schema.sql        # 数据库结构
└── uploads/                   # 文件上传目录
```

---

# 八、环境要求

| 软件/服务 | 版本要求 | 说明 |
|---------|---------|------|
| Go | 1.17+ | 后端开发语言 |
| Node.js | 14.0+ | 前端开发环境 |
| MySQL | 5.7+ / 8.0+ | 数据库 |
| Nginx | 1.18+ | Web 服务器(生产环境) |
| Git | 2.0+ | 代码管理 |

**操作系统支持**：Linux / macOS / Windows

**浏览器支持**：Chrome 80+、Edge 80+、Safari 13+

---

# 九、快速开始（5 分钟部署）

```bash
# 1. 克隆项目
git clone <repository-url>
cd LabEquip-BMS

# 2. 初始化数据库
mysql -u root -p < database_schema.sql

# 3. 修改配置文件
# 编辑 server/_config/conf.yaml，修改数据库密码

# 4. 启动后端
cd server/api
go run main.go

# 5. 启动前端(新终端)
cd frontend
npm install
npm start
```

访问：`http://localhost:3000`

---

# 十、详细部署步骤（零基础版）

## 10.1 安装 MySQL 数据库

### macOS
```bash
# 使用 Homebrew 安装
brew install mysql@8.0

# 启动 MySQL 服务
brew services start mysql@8.0

# 初始化 root 密码(可选)
mysql_secure_installation
```

### Linux (Ubuntu/Debian)
```bash
# 更新包管理器
sudo apt update

# 安装 MySQL
sudo apt install mysql-server

# 启动服务
sudo systemctl start mysql
sudo systemctl enable mysql

# 初始化配置
sudo mysql_secure_installation
```

### Windows
1. 下载 MySQL Installer：https://dev.mysql.com/downloads/installer/
2. 运行安装程序，选择 "Developer Default"
3. 设置 root 密码(建议：`123456`)
4. 完成安装

---

## 10.2 初始化数据库

```bash
# 登录 MySQL
mysql -u root -p
# 输入密码后回车

# 执行数据库初始化脚本
source /path/to/LabEquip-BMS/database_schema.sql

# 或者在命令行直接导入
mysql -u root -p < database_schema.sql

# 验证数据库创建成功
mysql -u root -p -e "SHOW DATABASES;"
# 应该看到: bms_basic_db, bms_inv_db, bms_task_db
```

**默认创建的数据库**：
- `bms_basic_db`：用户、实验室、设备等基础信息
- `bms_inv_db`：库存管理
- `bms_task_db`：借还任务管理

---

## 10.3 安装 Go 环境

### macOS
```bash
brew install go@1.17
```

### Linux
```bash
wget https://golang.google.cn/dl/go1.17.13.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.17.13.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Windows
1. 下载：https://golang.google.cn/dl/go1.17.13.windows-amd64.msi
2. 运行安装程序
3. 验证：`go version`

### 配置 Go 模块代理(国内加速)
```bash
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

---

## 10.4 安装 Node.js 环境

### macOS
```bash
brew install node@14
```

### Linux
```bash
curl -fsSL https://deb.nodesource.com/setup_14.x | sudo -E bash -
sudo apt-get install -y nodejs
```

### Windows
1. 下载：https://nodejs.org/dist/v14.21.3/node-v14.21.3-x64.msi
2. 运行安装程序
3. 验证：`node -v` 和 `npm -v`

### 配置 npm 镜像(国内加速)
```bash
npm config set registry https://registry.npmmirror.com
```

---

## 10.5 下载项目并安装依赖

```bash
# 克隆项目
git clone <repository-url>
cd LabEquip-BMS

# 安装后端依赖
go mod download

# 安装前端依赖
cd frontend
npm install
cd ..
```

---

## 10.6 修改配置文件

### 10.6.1 修改数据库配置

编辑文件：`server/_config/conf.yaml`

```yaml
datasource:
  default:
    driver: mysql
    maxOpenConns: 200
    connMaxLifetime: 3600
    maxIdleConns: 10
    autoReport: true
    groups:
      # ⚠️ 修改这里的数据库密码
      - masterDsn: "root:你的MySQL密码@(127.0.0.1:3306)/bms_basic_db"
        replicasDsn:
          - "root:你的MySQL密码@(127.0.0.1:3306)/bms_basic_db"
  inv:
    driver: mysql
    maxOpenConns: 200
    connMaxLifetime: 3600
    maxIdleConns: 10
    autoReport: true
    groups:
      # ⚠️ 修改这里的数据库密码
      - masterDsn: "root:你的MySQL密码@(127.0.0.1:3306)/bms_inv_db"
        replicasDsn:
          - "root:你的MySQL密码@(127.0.0.1:3306)/bms_inv_db"
  bms:
    driver: mysql
    maxOpenConns: 200
    connMaxLifetime: 3600
    maxIdleConns: 10
    autoReport: true
    groups:
      # ⚠️ 修改这里的数据库密码
      - masterDsn: "root:你的MySQL密码@(127.0.0.1:3306)/bms_task_db"
        replicasDsn:
          - "root:你的MySQL密码@(127.0.0.1:3306)/bms_task_db"

# 删除或注释掉 kafka 和 es 相关配置(如果不需要)
# kafka: ...
# es: ...
# cache: ...
```

### 10.6.2 创建上传文件目录

```bash
# 在项目根目录创建 uploads 目录
mkdir -p uploads
chmod 755 uploads
```

### 10.6.3 修改前端配置(可选)

如果需要修改后端 API 地址，编辑 `frontend/package.json`：

```json
{
  "proxy": "http://localhost:8080"
}
```

---

## 10.7 启动后端服务

```bash
# 方式1: 开发模式(实时编译)
cd server/api
go run main.go

# 方式2: 编译后运行(推荐生产环境)
cd server/api
go build -o bms-api main.go
./bms-api

# 看到以下输出表示启动成功:
# configPath: /path/to/LabEquip-BMS/server/_config/conf.yaml
# Initialize success
# [GIN-debug] Listening and serving HTTP on 0.0.0.0:8080
```

**默认端口**：
- API 服务：`8080`
- 管理后台：`8082`(如需要)

---

## 10.8 启动前端服务

```bash
# 打开新终端
cd frontend
npm start

# 看到以下输出表示启动成功:
# Compiled successfully!
# You can now view labequip-bms-frontend in the browser.
# Local: http://localhost:3000
```

**访问地址**：`http://localhost:3000`

---

## 10.9 Nginx 生产环境配置(可选)

### 10.9.1 安装 Nginx

```bash
# macOS
brew install nginx

# Linux
sudo apt install nginx

# Windows
# 下载：http://nginx.org/en/download.html
```

### 10.9.2 配置 Nginx

1. **复制配置文件**

```bash
# Linux/macOS
sudo cp nginx/www.bms.com.conf /etc/nginx/sites-available/
sudo ln -s /etc/nginx/sites-available/www.bms.com.conf /etc/nginx/sites-enabled/

# macOS (Homebrew)
cp nginx/www.bms.com.conf /usr/local/etc/nginx/servers/
```

2. **修改配置文件**

编辑 `nginx/www.bms.com.conf`，根据实际情况修改：

```nginx
# 修改前端服务地址(如果前端已构建为静态文件)
location / {
    # 开发模式: 代理到前端开发服务器
    proxy_pass  http://127.0.0.1:3000;
    
    # 生产模式: 使用静态文件
    # root /path/to/LabEquip-BMS/frontend/build;
    # try_files $uri $uri/ /index.html;
}

# 修改日志路径
access_log /var/log/nginx/bms.access.log;
error_log /var/log/nginx/bms.error.log;
```

3. **修改 hosts 文件(本地测试)**

```bash
# macOS/Linux
sudo vim /etc/hosts
# 添加: 127.0.0.1 www.bms.com

# Windows
# 编辑: C:\Windows\System32\drivers\etc\hosts
# 添加: 127.0.0.1 www.bms.com
```

4. **重启 Nginx**

```bash
# macOS
brew services restart nginx

# Linux
sudo systemctl restart nginx

# 测试配置
nginx -t
```

5. **访问系统**

打开浏览器访问：`http://www.bms.com`

---

## 10.10 前端构建为生产版本

```bash
cd frontend
npm run build

# 构建完成后，静态文件在 frontend/build 目录
# 配置 Nginx 指向该目录即可
```

---

# 十一、Docker 部署（推荐）

## 11.1 创建 Dockerfile

在项目根目录创建 `Dockerfile`：

```dockerfile
# 多阶段构建 - 后端
FROM golang:1.17-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN cd server/api && go build -o /app/bms-api main.go

# 多阶段构建 - 前端
FROM node:14-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# 最终镜像
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# 复制后端可执行文件
COPY --from=backend-builder /app/bms-api .
COPY --from=backend-builder /app/server/_config ./server/_config

# 复制前端构建文件
COPY --from=frontend-builder /app/build ./frontend/build

# 创建上传目录
RUN mkdir -p uploads

EXPOSE 8080
CMD ["./bms-api"]
```

## 11.2 创建 docker-compose.yml

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: bms-mysql
    environment:
      MYSQL_ROOT_PASSWORD: 123456
      MYSQL_DATABASE: bms_basic_db
    ports:
      - "3306:3306"
    volumes:
      - mysql-data:/var/lib/mysql
      - ./database_schema.sql:/docker-entrypoint-initdb.d/init.sql
    networks:
      - bms-network

  backend:
    build: .
    container_name: bms-backend
    ports:
      - "8080:8080"
    depends_on:
      - mysql
    environment:
      - DB_HOST=mysql
      - DB_PORT=3306
      - DB_USER=root
      - DB_PASSWORD=123456
    volumes:
      - ./uploads:/root/uploads
      - ./server/_config:/root/server/_config
    networks:
      - bms-network

  nginx:
    image: nginx:alpine
    container_name: bms-nginx
    ports:
      - "80:80"
    volumes:
      - ./nginx/www.bms.com.conf:/etc/nginx/conf.d/default.conf
      - ./frontend/build:/usr/share/nginx/html
    depends_on:
      - backend
    networks:
      - bms-network

volumes:
  mysql-data:

networks:
  bms-network:
    driver: bridge
```

## 11.3 启动 Docker 容器

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 停止并删除数据
docker-compose down -v
```

访问：`http://localhost`

---

# 十二、配置说明

## 12.1 核心配置文件

| 文件路径 | 说明 | 需要修改 |
|---------|------|---------|
| `server/_config/conf.yaml` | 后端主配置(数据库/端口) | ✅ 必须 |
| `frontend/package.json` | 前端依赖和代理配置 | ⚠️ 可选 |
| `nginx/www.bms.com.conf` | Nginx 路由配置 | ⚠️ 可选 |
| `database_schema.sql` | 数据库初始化脚本 | ❌ 不需要 |

## 12.2 上传文件目录配置

**默认上传目录**：`项目根目录/uploads`

如需修改，需要同步修改：
1. 后端文件上传逻辑(在 `framework/web/upload_file_strategy.go` 中)
2. Nginx 静态文件配置

```bash
# 修改上传目录权限
chmod -R 755 uploads
chown -R www-data:www-data uploads  # Linux
```

## 12.3 端口配置

**默认端口**：
- 后端 API：`8080`
- 前端开发服务器：`3000`
- Nginx：`80`

**修改端口**：
- 后端：修改 `server/_config/conf.yaml` 中的 `server.addr.api`
- 前端：修改 `frontend/package.json` 中的 `scripts.start`
- Nginx：修改 `nginx/www.bms.com.conf` 中的 `listen`

---

# 十三、常见问题排查

## 13.1 数据库连接失败

**错误信息**：`Error 1045: Access denied for user 'root'@'localhost'`

**解决方案**：
```bash
# 重置 MySQL root 密码
mysql -u root
ALTER USER 'root'@'localhost' IDENTIFIED BY '新密码';
FLUSH PRIVILEGES;

# 修改 conf.yaml 中的数据库密码
```

## 13.2 Go 依赖下载失败

**错误信息**：`go: module xxx: Get "https://proxy.golang.org/...": dial tcp xxx: i/o timeout`

**解决方案**：
```bash
# 配置国内代理
go env -w GOPROXY=https://goproxy.cn,direct
go mod download
```

## 13.3 前端启动失败

**错误信息**：`Error: EACCES: permission denied`

**解决方案**：
```bash
# 清理 node_modules 重新安装
rm -rf node_modules package-lock.json
npm cache clean --force
npm install
```

## 13.4 端口被占用

**错误信息**：`bind: address already in use`

**解决方案**：
```bash
# 查看占用端口的进程
lsof -i :8080  # macOS/Linux
netstat -ano | findstr 8080  # Windows

# 杀死进程或修改配置文件中的端口
```

## 13.5 文件上传失败

**检查清单**：
1. `uploads` 目录是否存在
2. 目录权限是否正确(`chmod 755 uploads`)
3. 磁盘空间是否充足(`df -h`)
4. Nginx 配置是否正确

---

# 十四、系统默认账号

**初始管理员账号**（需要通过注册接口创建）：

- 邮箱：`admin@example.com`
- 密码：需要首次注册时设置

**角色说明**：
- `1` - 超级管理员
- `2` - 管理员
- `3` - 教师/助教
- `4` - 学生

---

# 十五、开发者指南

## 15.1 开发工具推荐

- **IDE**：GoLand / VS Code
- **API 测试**：Postman / Apifox
- **数据库管理**：Navicat / DBeaver
- **Git 工具**：SourceTree / GitKraken

## 15.2 代码规范

- Go 代码遵循 `gofmt` 规范
- 前端代码遵循 ESLint 规范
- 提交信息遵循 Conventional Commits

## 15.3 项目启动顺序

1. MySQL 数据库
2. 后端 API 服务
3. 前端开发服务器
4. Nginx(生产环境)

---

# 十六、技术支持

如遇到部署问题，请检查：

1. ✅ 是否按照文档顺序执行
2. ✅ 数据库是否正确初始化
3. ✅ 配置文件中的密码是否正确
4. ✅ 端口是否被占用
5. ✅ 防火墙是否放行端口

**日志位置**：
- 后端日志：控制台输出
- Nginx 日志：`/var/log/nginx/`
- MySQL 日志：`/var/log/mysql/`

---

# 十七、许可证

本软件仅用于教学和学习目的，未经授权不得用于商业用途。

---**更新时间**: 2026-01-16  
**版本**: V1.0  
**维护者**: LabEquip-BMS Team
