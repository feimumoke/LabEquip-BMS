# LabEquip-BMS 部署说明文档

## 📋 目录

- [快速部署](#快速部署)
- [手动部署](#手动部署)
- [Docker 部署](#docker-部署)
- [生产环境部署](#生产环境部署)
- [常见问题](#常见问题)

---

## 🚀 快速部署

### 方式一：使用一键部署脚本（推荐）

```bash
# 1. 给脚本添加执行权限
chmod +x deploy.sh

# 2. 执行部署脚本
./deploy.sh

# 3. 访问系统
# 浏览器打开: http://localhost
```

### 方式二：使用 Docker Compose

```bash
# 1. 构建并启动所有服务
docker-compose up -d

# 2. 查看服务状态
docker-compose ps

# 3. 查看日志
docker-compose logs -f

# 4. 访问系统
# 浏览器打开: http://localhost
```

---

## 🔧 手动部署

### 前置要求

- Go 1.17+
- Node.js 14+
- MySQL 5.7+ / 8.0+
- Nginx 1.18+ (可选)

### 步骤1：安装数据库

```bash
# 安装 MySQL
# macOS:
brew install mysql@8.0
brew services start mysql@8.0

# Linux:
sudo apt update
sudo apt install mysql-server
sudo systemctl start mysql
```

### 步骤2：初始化数据库

```bash
# 导入数据库结构
mysql -u root -p < database_schema.sql

# 验证数据库创建成功
mysql -u root -p -e "SHOW DATABASES;"
```

### 步骤3：配置后端

```bash
# 1. 修改配置文件
vim server/_config/conf.yaml
# 修改数据库密码为你的 MySQL 密码

# 2. 安装依赖
go mod download

# 3. 启动后端
cd server/api
go run main.go
```

### 步骤4：配置前端

```bash
# 1. 安装依赖
cd frontend
npm install

# 2. 启动前端（开发模式）
npm start

# 或者构建生产版本
npm run build
```

### 步骤5：配置 Nginx（生产环境）

```bash
# 1. 复制配置文件
sudo cp nginx/www.bms.com.conf /etc/nginx/sites-available/
sudo ln -s /etc/nginx/sites-available/www.bms.com.conf /etc/nginx/sites-enabled/

# 2. 修改配置中的路径
sudo vim /etc/nginx/sites-available/www.bms.com.conf

# 3. 重启 Nginx
sudo systemctl restart nginx
```

---

## 🐳 Docker 部署

### 环境要求

- Docker 20.10+
- Docker Compose 1.29+

### 部署步骤

#### 1. 克隆项目

```bash
git clone <repository-url>
cd LabEquip-BMS
```

#### 2. 修改配置

```bash
# 编辑数据库配置
vim server/_config/conf.yaml

# 将数据库主机从 127.0.0.1 改为 mysql
# 示例：
# masterDsn: "root:123456@(mysql:3306)/bms_basic_db"
```

#### 3. 构建镜像

```bash
# 构建后端和前端镜像
docker-compose build
```

#### 4. 启动服务

```bash
# 启动所有服务
docker-compose up -d
```

#### 5. 验证部署

```bash
# 查看服务状态
docker-compose ps

# 查看后端日志
docker-compose logs -f backend

# 查看数据库日志
docker-compose logs -f mysql
```

#### 6. 访问系统

打开浏览器访问：
- 前端：http://localhost
- 后端 API：http://localhost:8080

### Docker 常用命令

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 重启服务
docker-compose restart

# 查看日志
docker-compose logs -f

# 查看服务状态
docker-compose ps

# 进入容器
docker-compose exec backend sh
docker-compose exec mysql bash

# 重新构建
docker-compose build --no-cache

# 清理所有数据（包括数据库）
docker-compose down -v
```

---

## 🏭 生产环境部署

### 架构说明

生产环境建议采用以下架构：

```
Internet
   ↓
Nginx (80/443)
   ↓
Backend API (8080)
   ↓
MySQL (3306)
```

### 部署清单

#### 1. 服务器配置

- **最低配置**：2核 4GB 内存 40GB 磁盘
- **推荐配置**：4核 8GB 内存 100GB 磁盘
- **操作系统**：Ubuntu 20.04 LTS / CentOS 8

#### 2. 安全配置

```bash
# 配置防火墙
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 22/tcp
sudo ufw enable

# MySQL 只监听本地
# 编辑 /etc/mysql/mysql.conf.d/mysqld.cnf
bind-address = 127.0.0.1
```

#### 3. SSL 证书配置

```bash
# 使用 Let's Encrypt 申请免费证书
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d yourdomain.com
```

#### 4. Nginx 优化配置

```nginx
# 添加到 nginx 配置
client_max_body_size 100M;
gzip on;
gzip_types text/plain text/css application/json application/javascript;

# 添加缓存
location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

#### 5. 后端进程管理

使用 systemd 管理后端进程：

```bash
# 创建 systemd 服务文件
sudo vim /etc/systemd/system/bms-backend.service

# 内容：
[Unit]
Description=LabEquip-BMS Backend Service
After=network.target mysql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/path/to/LabEquip-BMS
ExecStart=/path/to/LabEquip-BMS/server/api/bms-api
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable bms-backend
sudo systemctl start bms-backend
```

#### 6. 数据库备份

```bash
# 创建备份脚本
vim /root/backup-db.sh

#!/bin/bash
BACKUP_DIR="/backup/mysql"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

mysqldump -u root -p123456 --databases bms_basic_db bms_inv_db bms_task_db > $BACKUP_DIR/bms_$DATE.sql

# 只保留最近7天的备份
find $BACKUP_DIR -name "bms_*.sql" -mtime +7 -delete

# 添加到 crontab
crontab -e
# 每天凌晨2点执行备份
0 2 * * * /root/backup-db.sh
```

#### 7. 日志管理

```bash
# 配置日志轮转
sudo vim /etc/logrotate.d/bms

/var/log/nginx/bms.*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 www-data adm
    sharedscripts
    postrotate
        [ -f /var/run/nginx.pid ] && kill -USR1 `cat /var/run/nginx.pid`
    endscript
}
```

#### 8. 监控和告警

```bash
# 使用 uptime kuma 或其他监控工具
# 监控以下指标：
# - 服务器 CPU/内存/磁盘使用率
# - MySQL 连接数和慢查询
# - Nginx 访问量和错误率
# - 后端 API 响应时间
```

---

## ❓ 常见问题

### 1. 数据库连接失败

**问题**：后端无法连接数据库

**解决方案**：
```bash
# 检查 MySQL 是否启动
sudo systemctl status mysql

# 检查密码是否正确
mysql -u root -p

# 检查配置文件中的连接字符串
vim server/_config/conf.yaml
```

### 2. 端口冲突

**问题**：端口已被占用

**解决方案**：
```bash
# 查看占用端口的进程
lsof -i :8080
netstat -tunlp | grep 8080

# 杀死进程或修改配置中的端口
```

### 3. 文件上传失败

**问题**：无法上传文件

**解决方案**：
```bash
# 检查 uploads 目录权限
ls -la uploads/
chmod 755 uploads/

# 检查磁盘空间
df -h

# 检查 Nginx 配置
client_max_body_size 100M;
```

### 4. 前端页面空白

**问题**：访问前端页面显示空白

**解决方案**：
```bash
# 检查前端是否构建
cd frontend && npm run build

# 检查 Nginx 配置中的 root 路径
# 检查浏览器控制台是否有错误
```

### 5. Docker 容器无法启动

**问题**：docker-compose up 失败

**解决方案**：
```bash
# 查看详细日志
docker-compose logs

# 检查端口是否被占用
docker-compose down
netstat -tunlp | grep 3306

# 清理并重新构建
docker-compose down -v
docker-compose build --no-cache
docker-compose up -d
```

### 6. Go 依赖下载失败

**问题**：go mod download 超时

**解决方案**：
```bash
# 配置 Go 代理
go env -w GOPROXY=https://goproxy.cn,direct
go mod download
```

### 7. npm 安装依赖失败

**问题**：npm install 报错

**解决方案**：
```bash
# 清理缓存
rm -rf node_modules package-lock.json
npm cache clean --force

# 配置镜像
npm config set registry https://registry.npmmirror.com

# 重新安装
npm install
```

---

## 📞 技术支持

如遇到其他问题，请：

1. 查看日志文件
2. 检查配置是否正确
3. 参考 README.md 中的详细说明
4. 搜索相关错误信息

---

**最后更新**: 2026-01-16  
**版本**: V1.0
