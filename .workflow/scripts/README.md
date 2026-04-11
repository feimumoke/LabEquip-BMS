# Workflow Scripts

本目录包含工作流自动化脚本。

## 脚本列表

### scan_tests.py (v3.0.0)

测试用例扫描脚本，用于从 git diff 中扫描新增/修改的测试函数。

**功能：**
- 自动查找最新的需求目录
- 从 git diff release 提取新增的测试函数（仅扫描 `func Test*` 定义行）
- 保存到需求的 tests 目录

**使用方法：**

```bash
# 查看版本
python .workflow/scripts/scan_tests.py --version

# 直接运行（使用默认 release 分支：origin/release）
python .workflow/scripts/scan_tests.py

# 指定 release 分支
python .workflow/scripts/scan_tests.py --release origin/main

# 或通过插件侧边栏的"扫描测试"按钮
```

**工作原理：**

1. 与 release 分支对比，获取 diff 输出
2. 从 diff 新增行（以 `+` 开头）中提取 `func Test*` 定义
3. 清除旧数据，仅保存本次扫描的新增/修改测试函数

### calc_incremental_coverage.py (v1.1.0)

增量覆盖率计算脚本，用于计算与 release 分支对比的代码变更覆盖情况。

**功能：**
- 解析 git diff 获取与 release 分支的代码变更
- 解析 coverage.out 获取测试覆盖率数据
- 计算增量覆盖率（变更代码的覆盖情况）
- 计算函数级覆盖率
- 生成风险评估报告

**使用方法：**

```bash
# 查看版本
python .workflow/scripts/calc_incremental_coverage.py --version

# 直接运行（使用默认 release 分支和自动检测 coverage 文件）
python .workflow/scripts/calc_incremental_coverage.py

# 指定 coverage 文件
python .workflow/scripts/calc_incremental_coverage.py coverage.out

# 指定 release 分支
python .workflow/scripts/calc_incremental_coverage.py --release origin/main

# 静默模式（只输出 JSON）
python .workflow/scripts/calc_incremental_coverage.py -q coverage.out
```

**输出示例（JSON 格式）：**

```json
{
  "summary": {
    "incrementalCoveragePercent": 85.5,
    "fullCoveragePercent": 78.2,
    "changedLines": 120,
    "coveredLines": 103,
    "changedFunctions": 8,
    "coveredFunctions": 6,
    "riskLevel": "medium",
    "testDebt": 17
  }
}
```

## 版本管理

脚本版本由 `VERSION.json` 文件管理。当插件更新时，如果模板版本高于本地版本，脚本会自动更新。

**更新日志：**

**scan_tests.py:**
- **v3.0.0**: 简化为仅扫描 git diff release 中新增/修改的测试函数，移除 A/B 集合逻辑
- **v2.2.0**: 脚本目录结构调整，与 templates 平级
- **v2.1.0**: 扫描结果为空时也更新 test-stats.json，清除旧数据
- **v2.0.0**: 移除分支依赖，自动查找最新需求目录，新增 git diff release 对比功能
- **v1.0.0**: 初始版本

**calc_incremental_coverage.py:**
- **v1.1.0**: 优化输出格式，日志输出到 stderr，JSON 输出到 stdout
- **v1.0.0**: 初始版本

## 自定义

如果您需要自定义脚本行为，建议：
1. 复制脚本到其他位置
2. 修改 `.workflow/config.json` 中的脚本路径配置
3. 这样可以避免插件更新时覆盖您的修改
