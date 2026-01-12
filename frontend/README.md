# 实验设备借记管理系统 - 前端

基于 React + Ant Design 构建的实验设备借记管理系统前端应用。

## 技术栈

- React 18
- React Router v6
- Ant Design 5
- MobX 6
- Axios
- Day.js

## 项目结构

```
frontend/
├── public/              # 静态资源
├── src/
│   ├── api/            # API 接口
│   │   ├── auth.js     # 认证相关
│   │   ├── equip.js    # 设备管理
│   │   ├── inventory.js # 库存管理
│   │   └── borrow.js   # 借记管理
│   ├── components/     # 组件
│   │   ├── Layout/     # 布局组件
│   │   └── PermissionCheck.jsx  # 权限检查
│   ├── pages/          # 页面
│   │   ├── Login/      # 登录/注册
│   │   ├── Home/       # 首页
│   │   ├── Users/      # 用户管理
│   │   ├── Equipment/  # 设备管理
│   │   ├── Inventory/  # 库存管理
│   │   └── Borrow/     # 借记管理
│   ├── store/          # 状态管理
│   │   └── authStore.js # 认证状态
│   ├── utils/          # 工具函数
│   │   ├── request.js  # HTTP 请求
│   │   └── auth.js     # 认证工具
│   ├── App.js          # 应用入口
│   └── index.js        # React 入口
└── package.json
```

## 功能模块

### 1. 用户认证
- 用户注册（学生）
- 用户登录
- 保持登录态
- 权限控制

### 2. 用户管理
- 查看所有用户（超级管理员和教师）
- 用户角色管理

### 3. 设备管理
- 创建设备（超级管理员和教师）
- 查询设备（所有人）
- 设备分类管理

### 4. 库存管理
- 增加库存（超级管理员和教师）
- 扣减库存（超级管理员和教师）
- 查询库存（所有人）
- 库存任务查询
- 三级账查询

### 5. 借记管理
- 创建借记任务（所有人）
- 取消借记任务（所有人）
- 拿走借记物品（所有人）
- 归还借记物品（所有人）
- 审批借记（教师）
- 借记任务查询
  - 教师：查看所有借记任务
  - 学生：只能查看自己的借记任务

## 角色权限

- **超级管理员（Super Admin）**：所有权限
- **管理员（Admin）**：用户管理、设备管理、库存管理、审批借记
- **教师（Teacher）**：用户管理、设备管理、库存管理、审批借记
- **学生（Student）**：查看设备、查看库存、创建和管理自己的借记任务

## 开发指南

### 安装依赖

```bash
cd frontend
npm install
```

### 启动开发服务器

```bash
npm start
```

应用将在 http://localhost:3000 启动

### 构建生产版本

```bash
npm run build
```

### API 代理配置

开发环境下，API 请求会代理到 `http://localhost:8080`，在 `package.json` 中配置：

```json
{
  "proxy": "http://localhost:8080"
}
```

## 待实现页面

以下页面组件需要根据具体需求实现：

- [x] 登录/注册页面
- [ ] 首页（统计数据展示）
- [ ] 用户管理页面
- [ ] 设备管理页面
- [ ] 库存列表页面
- [ ] 库存操作页面
- [ ] 库存任务页面
- [ ] 三级账查询页面
- [ ] 我的借记页面
- [ ] 所有借记页面
- [ ] 审批管理页面

## 注意事项

1. 所有 API 请求都需要在 Header 中携带 Token 和 UserEmail
2. Token 存储在 localStorage 中
3. 401 错误会自动跳转到登录页
4. 权限检查通过 PermissionCheck 组件实现
5. 使用 MobX 管理全局认证状态

