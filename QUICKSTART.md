# ⚡ 快速启动指南

## 🎯 适用场景

如果你是第一次部署本系统,或者对软件部署不太熟悉,请按照本指南操作。

---

## 📌 部署前准备

### 检查清单

- [ ] 你的电脑已安装 **Go 1.17+**
- [ ] 你的电脑已安装 **Node.js 14+**
- [ ] 你的电脑已安装 **MySQL 5.7+**
- [ ] 你的 MySQL 服务正在运行
- [ ] 你知道 MySQL 的 root 密码

如果有任何一项不满足,请先参考 [README.md](README.md) 第十章的详细安装说明。

---

## 🚀 5分钟快速部署

### 第 1 步：下载项目

```bash
cd ~/WorkSpace  # 或者你喜欢的目录
git clone <repository-url>
cd LabEquip-BMS
```

### 第 2 步：初始化数据库

```bash
# 登录 MySQL (输入你的密码)
mysql -u root -p

# 在 MySQL 中执行以下命令
source /Users/hui.zhu/WorkSpace/LabEquip-BMS/database_schema.sql;
SHOW DATABASES;  # 应该看到 bms_basic_db, bms_inv_db, bms_task_db
exit;
```

### 第 3 步：修改配置

```bash
# 打开配置文件
vim server/_config/conf.yaml

# 或者使用其他编辑器
open server/_config/conf.yaml  # macOS
gedit server/_config/conf.yaml  # Linux
notepad server/_config/conf.yaml  # Windows
```

**修改内容**：将所有 `root:123456` 中的 `123456` 改为你的 MySQL 密码

例如，如果你的密码是 `mypassword`,则改为：
```yaml
masterDsn: "root:mypassword@(127.0.0.1:3306)/bms_basic_db"
```

**重要**：需要修改 3 个地方（对应 3 个数据库）

### 第 4 步：启动后端

```bash
# 打开终端1,启动后端
cd server/api
go run main.go

# 看到 "Initialize success" 表示启动成功
# 不要关闭这个终端
```

### 第 5 步：启动前端

```bash
# 打开新终端2,启动前端
cd frontend
npm install  # 首次需要安装依赖,耗时约2-5分钟
npm start

# 看到 "Compiled successfully!" 表示启动成功
# 浏览器会自动打开 http://localhost:3000
```

### 第 6 步：访问系统

打开浏览器访问：**http://localhost:3000**

🎉 恭喜！系统已成功启动！

---

## 🐳 使用 Docker 部署（更简单）

如果你已经安装了 Docker 和 Docker Compose,可以使用一键部署：

```bash
# 1. 给脚本添加执行权限
chmod +x deploy.sh

# 2. 执行部署脚本
./deploy.sh

# 3. 等待 2-3 分钟,浏览器访问 http://localhost
```

---

## 🎓 注册第一个账号

1. 访问系统后,点击「注册」
2. 填写信息：
   - 邮箱：`admin@example.com`
   - 用户名：`管理员`
   - 密码：`admin123`（或你喜欢的密码）
   - 角色：选择「超级管理员」
3. 点击「注册」完成

---

## 🛠️ 常见启动问题

### 问题1：MySQL 连接失败

```
错误: Error 1045: Access denied for user 'root'@'localhost'
```

**原因**：配置文件中的密码不正确

**解决**：
1. 检查你的 MySQL 密码是否正确
2. 重新编辑 `server/_config/conf.yaml`
3. 确保修改了所有 3 个数据库的密码

### 问题2：端口被占用

```
错误: bind: address already in use
```

**原因**：端口 8080 或 3000 被其他程序占用

**解决**：
```bash
# 查看占用端口的进程
lsof -i :8080  # macOS/Linux
lsof -i :3000

# 杀死进程
kill -9 <PID>
```

### 问题3：数据库不存在

```
错误: Unknown database 'bms_basic_db'
```

**原因**：数据库初始化脚本未执行

**解决**：
```bash
mysql -u root -p < database_schema.sql
```

### 问题4：Go 依赖下载失败

```
错误: go: module xxx: Get "https://proxy.golang.org/...": timeout
```

**解决**：
```bash
# 配置国内镜像
go env -w GOPROXY=https://goproxy.cn,direct
go mod download
```

### 问题5：npm 安装失败

```
错误: npm ERR! code EACCES
```

**解决**：
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

## 📝 下一步

系统启动成功后,你可以：

1. ✅ 创建实验室信息
2. ✅ 添加实验器材
3. ✅ 管理库存
4. ✅ 创建借用申请
5. ✅ 审批和归还设备

---

## 📚 更多文档

- [README.md](README.md) - 完整部署文档
- [DEPLOY.md](DEPLOY.md) - 生产环境部署指南
- [数据库设计文档.md](数据库设计文档.md) - 数据库结构说明

---

## 💡 提示

- 开发模式下,修改代码后会自动重新编译
- 前端修改会自动刷新浏览器
- 后端修改需要重启 `go run main.go`
- 数据库配置修改后需要重启后端

---

**需要帮助？** 查看 [README.md](README.md) 第十三章的常见问题排查

**更新时间**: 2026-01-16
