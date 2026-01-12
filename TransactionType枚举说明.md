# TransactionType 枚举说明

## 📋 枚举定义

### 1. TransactionType（交易类型）

**文件**：`defines/constant/transaction_enum.go`

| 值 | 常量名 | 中文名称 | 说明 | 颜色 |
|---|--------|---------|------|------|
| 1 | TransactionTypeIncrease | 增加库存 | 手动增加库存操作 | 🟢 绿色 |
| 2 | TransactionTypeDecrease | 扣减库存 | 手动扣减库存操作 | 🔴 红色 |
| 3 | TransactionTypeAllocate | 分配库存 | 审批通过，分配库存给借记任务 | 🔵 蓝色 |
| 4 | TransactionTypeBorrow | 借出设备 | 学生拿走设备 | 🟠 橙色 |
| 5 | TransactionTypeReturn | 归还设备 | 学生归还设备 | 🔷 青色 |
| 6 | TransactionTypeReject | 拒绝分配 | 审批拒绝，释放已分配的库存 | 🔥 火山红 |

### 2. TransactionSheetType（单据类型）

| 值 | 常量名 | 中文名称 | 说明 |
|---|--------|---------|------|
| 1 | TransactionSheetTypeInventory | 库存单 | 手动库存操作产生的单据 |
| 2 | TransactionSheetTypeBorrow | 借记单 | 借记任务产生的单据 |

---

## 🔧 后端实现

### 枚举注册

**文件**：`defines/constant/transaction_enum.go`

```go
// 导出给前端的交易类型枚举
var ExportTransactionTypeNameToValue = map[string]interface{}{
	"增加库存": TransactionTypeIncrease,  // 1
	"扣减库存": TransactionTypeDecrease,  // 2
	"分配库存": TransactionTypeAllocate,  // 3
	"借出设备": TransactionTypeBorrow,    // 4
	"归还设备": TransactionTypeReturn,    // 5
	"拒绝分配": TransactionTypeReject,    // 6
}

// 导出给前端的交易单据类型枚举
var ExportTransactionSheetTypeNameToValue = map[string]interface{}{
	"库存单": TransactionSheetTypeInventory,  // 1
	"借记单": TransactionSheetTypeBorrow,     // 2
}

func init() {
	RegisterEnumValues("TransactionType", ExportTransactionTypeNameToValue)
	RegisterEnumValues("TransactionSheetType", ExportTransactionSheetTypeNameToValue)
}
```

---

## 🎨 前端实现

### 页面组件

**文件**：`frontend/src/pages/Inventory/Transactions.jsx`

```javascript
import { observer } from 'mobx-react-lite';
import enumStore from '../../store/enumStore';

const InventoryTransactions = observer(() => {
  // 交易类型颜色映射
  const getTransTypeColor = (type) => {
    const colorMap = {
      1: 'green',    // 增加库存
      2: 'red',      // 扣减库存
      3: 'blue',     // 分配库存
      4: 'orange',   // 借出设备
      5: 'cyan',     // 归还设备
      6: 'volcano',  // 拒绝分配
    };
    return colorMap[type] || 'default';
  };

  const columns = [
    // ...
    {
      title: '交易类型',
      dataIndex: 'trans_type',
      key: 'trans_type',
      render: (type) => {
        const label = enumStore.getEnumLabel('TransactionType', type);
        const color = getTransTypeColor(type);
        return <Tag color={color}>{label}</Tag>;
      },
    },
    {
      title: '单据类型',
      dataIndex: 'sheet_type',
      key: 'sheet_type',
      render: (type) => {
        const label = enumStore.getEnumLabel('TransactionSheetType', type);
        return <span>{label}</span>;
      },
    },
    // ...
  ];
});
```

---

## 🎯 显示效果

### 交易类型标签

| 类型值 | 显示文本 | 标签颜色 | 使用场景 |
|-------|---------|---------|---------|
| 1 | 增加库存 | <Badge color="green">绿色</Badge> | 手动增加库存 |
| 2 | 扣减库存 | <Badge color="red">红色</Badge> | 手动扣减库存 |
| 3 | 分配库存 | <Badge color="blue">蓝色</Badge> | 审批通过，分配给学生 |
| 4 | 借出设备 | <Badge color="orange">橙色</Badge> | 学生拿走设备 |
| 5 | 归还设备 | <Badge color="cyan">青色</Badge> | 学生归还设备 |
| 6 | 拒绝分配 | <Badge color="volcano">火山红</Badge> | 审批拒绝，释放库存 |

### 单据类型显示

| 类型值 | 显示文本 | 说明 |
|-------|---------|------|
| 1 | 库存单 | 库存操作产生 |
| 2 | 借记单 | 借记任务产生 |

---

## 🔄 交易场景说明

### 场景 1：增加库存

**操作**：管理员在"库存操作"页面手动增加库存

**三级账记录**：
- 交易类型：`增加库存` (绿色标签)
- 单据类型：`库存单`
- 操作数量：`+10` (绿色)
- 总数：从 10 → 20
- 可用：从 5 → 15

---

### 场景 2：扣减库存

**操作**：管理员在"库存操作"页面手动扣减库存

**三级账记录**：
- 交易类型：`扣减库存` (红色标签)
- 单据类型：`库存单`
- 操作数量：`-5` (红色)
- 总数：从 20 → 15
- 可用：从 15 → 10

---

### 场景 3：分配库存

**操作**：教师审批通过学生的借记申请

**三级账记录**：
- 交易类型：`分配库存` (蓝色标签)
- 单据类型：`借记单`
- 操作数量：`0`
- 总数：15 (不变)
- 可用：从 10 → 7 (减少3)
- 分配：从 0 → 3 (增加3)

---

### 场景 4：借出设备

**操作**：学生从实验室拿走已分配的设备

**三级账记录**：
- 交易类型：`借出设备` (橙色标签)
- 单据类型：`借记单`
- 操作数量：`-3` (红色)
- 总数：15 (不变)
- 在手：从 15 → 12 (减少3)
- 借出：从 0 → 3 (增加3)
- 分配：从 3 → 0 (减少3)

---

### 场景 5：归还设备

**操作**：学生归还设备到实验室

**三级账记录**：
- 交易类型：`归还设备` (青色标签)
- 单据类型：`借记单`
- 操作数量：`+3` (绿色)
- 总数：15 (不变)
- 在手：从 12 → 15 (增加3)
- 可用：从 7 → 10 (增加3)
- 借出：从 3 → 0 (减少3)

---

### 场景 6：拒绝分配

**操作**：教师拒绝学生的借记申请

**三级账记录**：
- 交易类型：`拒绝分配` (火山红标签)
- 单据类型：`借记单`
- 操作数量：`0`
- 总数：15 (不变)
- 可用：从 7 → 10 (增加3)
- 分配：从 3 → 0 (减少3)

---

## 📊 三级账页面示例

### 表格显示

| 交易ID | 单据ID | 设备 | 交易类型 | 单据类型 | 操作数量 | 总数 | 在手 | 可用 | 借出 | 分配 |
|-------|-------|------|---------|---------|---------|-----|------|------|------|------|
| T001 | INV-001 | 显微镜 | <Tag color="green">增加库存</Tag> | 库存单 | +10 | 10 | 10 | 10 | 0 | 0 |
| T002 | BOR-001 | 显微镜 | <Tag color="blue">分配库存</Tag> | 借记单 | 0 | 10 | 10 | 7 | 0 | 3 |
| T003 | BOR-001 | 显微镜 | <Tag color="orange">借出设备</Tag> | 借记单 | -3 | 10 | 7 | 7 | 3 | 0 |
| T004 | INV-002 | 显微镜 | <Tag color="green">增加库存</Tag> | 库存单 | +5 | 15 | 12 | 12 | 3 | 0 |
| T005 | BOR-001 | 显微镜 | <Tag color="cyan">归还设备</Tag> | 借记单 | +3 | 15 | 15 | 15 | 0 | 0 |

---

## 🎨 颜色设计说明

### 交易类型颜色选择理由

| 类型 | 颜色 | 理由 |
|------|------|------|
| 增加库存 | 🟢 绿色 | 表示增长、正向操作 |
| 扣减库存 | 🔴 红色 | 表示减少、警示 |
| 分配库存 | 🔵 蓝色 | 表示规划、预留 |
| 借出设备 | 🟠 橙色 | 表示转移、使用中 |
| 归还设备 | 🔷 青色 | 表示恢复、回归 |
| 拒绝分配 | 🔥 火山红 | 表示取消、异常 |

### 单据类型显示

- 库存单：普通黑色文字（不突出）
- 借记单：普通黑色文字（不突出）

**设计理念**：单据类型是辅助信息，不需要特别突出，交易类型才是重点。

---

## 🧪 测试验证

### 验证清单

- [ ] 登录后枚举已加载（浏览器控制台显示 `TransactionType` 和 `TransactionSheetType`）
- [ ] 增加库存操作显示"增加库存"（绿色标签）
- [ ] 扣减库存操作显示"扣减库存"（红色标签）
- [ ] 审批通过后显示"分配库存"（蓝色标签）
- [ ] 学生拿走设备显示"借出设备"（橙色标签）
- [ ] 学生归还设备显示"归还设备"（青色标签）
- [ ] 审批拒绝显示"拒绝分配"（火山红标签）
- [ ] 单据类型正确显示"库存单"或"借记单"
- [ ] 所有交易类型都显示文字，不显示数字

### 测试步骤

1. **测试增加库存**：
   - 访问"库存管理" → "库存操作"
   - 点击"新增库存"，增加10个设备
   - 访问"三级账查询"
   - 验证最新记录显示"增加库存"（绿色）

2. **测试借记流程**：
   - 学生创建借记任务 → 教师审批通过 → 查看三级账
   - 验证显示"分配库存"（蓝色）
   - 学生拿走设备 → 查看三级账
   - 验证显示"借出设备"（橙色）
   - 学生归还设备 → 查看三级账
   - 验证显示"归还设备"（青色）

3. **测试拒绝场景**：
   - 学生创建借记任务 → 教师审批拒绝
   - 查看三级账
   - 验证显示"拒绝分配"（火山红）

---

## 🔍 调试方法

### 浏览器控制台检查

```javascript
// 1. 检查枚举是否已加载
console.log('TransactionType:', enumStore.getEnum('TransactionType'));
// 预期输出：{增加库存: 1, 扣减库存: 2, 分配库存: 3, ...}

console.log('TransactionSheetType:', enumStore.getEnum('TransactionSheetType'));
// 预期输出：{库存单: 1, 借记单: 2}

// 2. 测试枚举标签获取
console.log('类型1:', enumStore.getEnumLabel('TransactionType', 1));
// 预期输出："增加库存"

console.log('类型2:', enumStore.getEnumLabel('TransactionType', 2));
// 预期输出："扣减库存"

console.log('类型3:', enumStore.getEnumLabel('TransactionType', 3));
// 预期输出："分配库存"

// 3. 测试单据类型
console.log('单据1:', enumStore.getEnumLabel('TransactionSheetType', 1));
// 预期输出："库存单"

console.log('单据2:', enumStore.getEnumLabel('TransactionSheetType', 2));
// 预期输出："借记单"
```

### 如果显示数字而不是文字

**可能原因**：
1. 枚举未加载
2. 枚举 key 不匹配
3. 页面组件没有用 `observer` 包装

**解决方法**：
1. 检查后端是否正确注册枚举（重启后端）
2. 检查前端是否正确加载枚举（登录后加载）
3. 检查组件是否用 `observer` 包装

---

## 📚 相关文档

- [枚举管理框架](./BorrowTaskStatus枚举使用说明.md)
- [三级账字段说明](./三级账字段说明.md)
- [状态显示优化](./状态显示优化说明.md)

---

## 🎉 总结

### 已注册的枚举

1. ✅ **TransactionType**（交易类型）- 6 个值
2. ✅ **TransactionSheetType**（单据类型）- 2 个值

### 前端显示

- ✅ 交易类型：彩色标签，根据类型显示不同颜色
- ✅ 单据类型：普通文字，不强调

### 效果

- ✅ 三级账页面更直观易读
- ✅ 不同交易类型一目了然
- ✅ 颜色编码帮助快速识别

---

**文档版本**：v1.0  
**创建时间**：2026-01-12  
**最后更新**：2026-01-12
