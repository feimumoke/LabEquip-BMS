# BorrowTaskStatus 枚举使用说明

## 📋 枚举定义

### 后端定义 (`defines/constant/borrow_enum.go`)

```go
type BorrowTaskStatus = int64

const (
    BorrowTaskStatusPending  BorrowTaskStatus = 1 // 待审批（创建了，没分配库存）
    BorrowTaskStatusApproval BorrowTaskStatus = 2 // 已审批（审批通过）
    BorrowTaskStatusAllocate BorrowTaskStatus = 3 // 已分配（分配了库存）
    BorrowTaskStatusOngoing  BorrowTaskStatus = 4 // 进行中（已拿走）
    BorrowTaskStatusReject   BorrowTaskStatus = 7 // 已拒绝（审批拒绝）
    BorrowTaskStatusDone     BorrowTaskStatus = 8 // 已归还
    BorrowTaskStatusCancel   BorrowTaskStatus = 9 // 已取消
)
```

### 前端枚举映射

系统会自动通过 `/api/apps/common/enums` 接口返回：

```json
{
  "retcode": 0,
  "data": {
    "BorrowTaskStatus": {
      "待审批": 1,
      "已审批": 2,
      "已分配": 3,
      "进行中": 4,
      "已拒绝": 7,
      "已归还": 8,
      "已取消": 9
    }
  }
}
```

---

## 🔄 状态流转

### 正常流程

```
1. 待审批 (Pending, 1)
   ├─ 创建借记任务
   └─ 等待分配库存
   
   ↓
   
2. 已审批 (Approval, 2)
   ├─ 库存分配成功
   └─ 等待教师审批
   
   ↓
   
3. 已分配 (Allocate, 3)
   ├─ 教师审批通过
   └─ 学生可以拿走设备
   
   ↓
   
4. 进行中 (Ongoing, 4)
   ├─ 学生已拿走设备
   └─ 等待归还
   
   ↓
   
8. 已归还 (Done, 8)
   └─ 设备已归还，任务完成
```

### 异常流程

```
待审批 (1) → 已取消 (9)
   └─ 学生主动取消

已分配 (3) → 已拒绝 (7)
   └─ 教师拒绝审批
```

---

## 💻 前端使用示例

### 1. 显示状态文本

#### 在表格中显示

```jsx
import { observer } from 'mobx-react-lite';
import enumStore from '../../store/enumStore';
import { Tag } from 'antd';

const BorrowTaskTable = observer(() => {
  const columns = [
    {
      title: '状态',
      dataIndex: 'task_status',
      key: 'task_status',
      render: (status) => {
        // 使用 enumStore 获取状态文本
        const label = enumStore.getEnumLabel('BorrowTaskStatus', status);
        
        // 根据状态显示不同颜色
        const colorMap = {
          1: 'orange',    // 待审批
          2: 'blue',      // 已审批
          3: 'cyan',      // 已分配
          4: 'green',     // 进行中
          7: 'red',       // 已拒绝
          8: 'default',   // 已归还
          9: 'default',   // 已取消
        };
        
        return <Tag color={colorMap[status]}>{label}</Tag>;
      },
    },
  ];
  
  return <Table columns={columns} dataSource={data} />;
});
```

### 2. 状态筛选器

```jsx
import { Select } from 'antd';
import { observer } from 'mobx-react-lite';
import enumStore from '../../store/enumStore';

const StatusFilter = observer(() => {
  const [status, setStatus] = useState(null);
  
  // 获取状态选项
  const statusOptions = enumStore.getEnumOptions('BorrowTaskStatus');
  
  return (
    <Select
      placeholder="请选择状态"
      value={status}
      onChange={setStatus}
      allowClear
      style={{ width: 200 }}
    >
      {statusOptions.map(option => (
        <Select.Option key={option.value} value={option.value}>
          {option.label}
        </Select.Option>
      ))}
    </Select>
  );
});
```

### 3. 根据状态显示不同操作

```jsx
const getActions = (record) => {
  const { task_status, task_id } = record;
  
  switch (task_status) {
    case 1: // 待审批 - 可以取消
      return (
        <Button onClick={() => handleCancel(task_id)}>
          取消
        </Button>
      );
      
    case 2: // 已审批 - 教师可以审批
      if (authStore.isTeacher) {
        return (
          <Space>
            <Button type="primary" onClick={() => handleApprove(task_id)}>
              通过
            </Button>
            <Button danger onClick={() => handleReject(task_id)}>
              拒绝
            </Button>
          </Space>
        );
      }
      break;
      
    case 3: // 已分配 - 学生可以拿走
      return (
        <Button type="primary" onClick={() => handleTake(task_id)}>
          拿走设备
        </Button>
      );
      
    case 4: // 进行中 - 可以归还
      return (
        <Button onClick={() => handleReturn(task_id)}>
          归还设备
        </Button>
      );
      
    default:
      return null;
  }
};
```

### 4. 状态统计

```jsx
import { Card, Statistic, Row, Col } from 'antd';
import { observer } from 'mobx-react-lite';
import enumStore from '../../store/enumStore';

const BorrowStatistics = observer(({ tasks }) => {
  // 统计各状态数量
  const statusCount = tasks.reduce((acc, task) => {
    acc[task.task_status] = (acc[task.task_status] || 0) + 1;
    return acc;
  }, {});
  
  return (
    <Row gutter={16}>
      <Col span={6}>
        <Card>
          <Statistic
            title={enumStore.getEnumLabel('BorrowTaskStatus', 1)}
            value={statusCount[1] || 0}
            valueStyle={{ color: '#faad14' }}
          />
        </Card>
      </Col>
      <Col span={6}>
        <Card>
          <Statistic
            title={enumStore.getEnumLabel('BorrowTaskStatus', 4)}
            value={statusCount[4] || 0}
            valueStyle={{ color: '#52c41a' }}
          />
        </Card>
      </Col>
      <Col span={6}>
        <Card>
          <Statistic
            title={enumStore.getEnumLabel('BorrowTaskStatus', 7)}
            value={statusCount[7] || 0}
            valueStyle={{ color: '#ff4d4f' }}
          />
        </Card>
      </Col>
      <Col span={6}>
        <Card>
          <Statistic
            title={enumStore.getEnumLabel('BorrowTaskStatus', 8)}
            value={statusCount[8] || 0}
          />
        </Card>
      </Col>
    </Row>
  );
});
```

---

## 🎨 推荐的颜色方案

根据状态的业务含义，推荐使用以下颜色：

| 状态值 | 状态名称 | 推荐颜色 | Ant Design 颜色 | 说明 |
|-------|---------|---------|----------------|------|
| 1 | 待审批 | 橙色 | `orange` | 等待中，需要关注 |
| 2 | 已审批 | 蓝色 | `blue` | 审批通过，流程进行中 |
| 3 | 已分配 | 青色 | `cyan` | 已分配库存 |
| 4 | 进行中 | 绿色 | `green` | 正在使用中 |
| 7 | 已拒绝 | 红色 | `red` | 拒绝/失败 |
| 8 | 已归还 | 灰色 | `default` | 已完成 |
| 9 | 已取消 | 灰色 | `default` | 已取消 |

### 使用示例

```jsx
const getStatusTagColor = (status) => {
  const colorMap = {
    1: 'orange',
    2: 'blue',
    3: 'cyan',
    4: 'green',
    7: 'red',
    8: 'default',
    9: 'default',
  };
  return colorMap[status] || 'default';
};

// 使用
<Tag color={getStatusTagColor(task.task_status)}>
  {enumStore.getEnumLabel('BorrowTaskStatus', task.task_status)}
</Tag>
```

---

## 📊 常见查询场景

### 1. 查询待审批的任务（需要教师处理）

```javascript
// 前端筛选
const pendingTasks = tasks.filter(task => task.task_status === 1);

// 或者后端查询
const res = await searchBorrowTask({
  task_status: 1, // 待审批
});
```

### 2. 查询进行中的任务（学生正在使用）

```javascript
const ongoingTasks = tasks.filter(task => task.task_status === 4);
```

### 3. 查询已完成的任务（归还、拒绝、取消）

```javascript
const finishedTasks = tasks.filter(task => 
  [7, 8, 9].includes(task.task_status)
);
```

### 4. 查询我的待处理任务（学生视角）

```javascript
// 待审批、已分配、进行中
const myActiveTasks = tasks.filter(task => 
  [1, 2, 3, 4].includes(task.task_status)
);
```

---

## 🔐 权限控制

### 学生权限

- ✅ 创建借记任务（状态变为 1-待审批）
- ✅ 取消待审批的任务（1 → 9）
- ✅ 拿走已分配的设备（3 → 4）
- ✅ 归还进行中的设备（4 → 8）
- ✅ 查看自己的借记任务

### 教师权限

- ✅ 审批待审批的任务（1 → 2 或 7）
- ✅ 查看所有学生的借记任务
- ✅ 所有学生权限

### 管理员权限

- ✅ 所有教师权限
- ✅ 管理库存分配

---

## 🧪 测试场景

### 1. 创建借记任务

```javascript
const task = await createBorrow({
  equip_id: 'EQ001',
  lab_code: 'LAB001',
  borrow_qty: 2,
  reason: '实验需要',
});

// 预期：task.task_status = 1 (待审批)
console.assert(task.task_status === 1);
```

### 2. 审批流程

```javascript
// 审批通过
await approveBorrow({
  task_id: 'TASK001',
  is_approved: true,
});
// 预期：task_status = 2 (已审批)

// 审批拒绝
await approveBorrow({
  task_id: 'TASK002',
  is_approved: false,
  reject_reason: '库存不足',
});
// 预期：task_status = 7 (已拒绝)
```

### 3. 取消任务

```javascript
await cancelBorrow({
  task_id: 'TASK003',
});
// 预期：task_status = 9 (已取消)
```

---

## 📝 注意事项

### 1. 状态值不连续

状态值为 1, 2, 3, 4, 7, 8, 9，**不是连续的**！

- ❌ 不要使用 `for (let i = 1; i <= 9; i++)` 遍历
- ✅ 使用 `enumStore.getEnumOptions('BorrowTaskStatus')` 获取所有选项

### 2. 状态转换限制

某些状态转换是不允许的：

```javascript
// ❌ 不允许的转换
4 (进行中) → 1 (待审批)  // 不能回退
8 (已归还) → 4 (进行中)  // 已完成的不能重开
7 (已拒绝) → 2 (已审批)  // 已拒绝的不能再审批
```

### 3. 权限检查

在执行状态转换操作前，务必检查用户权限：

```javascript
// 审批操作只有教师可以执行
if (!authStore.isTeacher) {
  message.error('无权限执行此操作');
  return;
}

await approveBorrow({...});
```

---

## 🎉 总结

`BorrowTaskStatus` 枚举已成功注册，现在可以：

1. ✅ 前端通过 `enumStore.getEnumLabel()` 获取状态文本
2. ✅ 前端通过 `enumStore.getEnumOptions()` 获取下拉选项
3. ✅ 使用统一的枚举管理，前后端保持一致
4. ✅ 支持中文显示，用户友好

**状态 1（待审批）** 是需要审批的初始状态，适用于借记任务创建后等待教师审批的场景。

---

**文档版本**: v1.0  
**最后更新**: 2026-01-12
