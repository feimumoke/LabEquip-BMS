#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
增量覆盖率计算脚本

版本: 1.3.2

功能：
1. 解析 git diff 获取与 release 分支的代码变更
2. 解析 coverage.out 获取测试覆盖率数据
3. 计算增量覆盖率（变更代码的覆盖情况）
4. 计算函数级覆盖率（全量覆盖）
5. 生成风险评估报告

更新说明（v1.3.0）：
- 全量覆盖现在包含增量覆盖中函数外的可执行行
- 确保全量覆盖行数 >= 增量覆盖行数，避免逻辑混淆

使用方法：
    python .workflow/scripts/calc_incremental_coverage.py
    python .workflow/scripts/calc_incremental_coverage.py --release origin/release
    python .workflow/scripts/calc_incremental_coverage.py --coverage-file tests/coverage.out
"""

__version__ = '1.3.2'

import os
import re
import json
import subprocess
import sys
import argparse
from datetime import datetime
from collections import defaultdict
from typing import Dict, List, Set, Tuple, Optional, Any

# ============================================
# 配置
# ============================================

DEFAULT_RELEASE_BRANCH = 'origin/release'
GO_FILE_PATTERN = r'\.go$'
TEST_FILE_PATTERN = r'_test\.go$'
FUNC_PATTERN = r'^func\s+(?:\([^)]+\)\s+)?(\w+)\s*\('
REQUIREMENT_DIR_PATTERN = r'^R-\d{4}'

# 风险阈值
HIGH_RISK_THRESHOLD = 60  # 覆盖率低于此值为高风险
MEDIUM_RISK_THRESHOLD = 80

# ============================================
# 工具函数
# ============================================

def run_command(cmd: str, check: bool = True) -> Optional[str]:
    """运行 shell 命令并返回输出"""
    try:
        result = subprocess.run(
            cmd,
            shell=True,
            capture_output=True,
            text=True,
            check=check
        )
        return result.stdout.strip()
    except subprocess.CalledProcessError as e:
        if check:
            print(f"Error running command: {cmd}", file=sys.stderr)
            print(f"Error: {e.stderr}", file=sys.stderr)
        return None

def find_workflow_dir() -> str:
    """查找 .workflow 目录"""
    current_dir = os.getcwd()
    while current_dir != '/':
        workflow_dir = os.path.join(current_dir, '.workflow')
        if os.path.isdir(workflow_dir):
            return workflow_dir
        current_dir = os.path.dirname(current_dir)
    
    print("Error: .workflow directory not found", file=sys.stderr)
    sys.exit(1)

def find_latest_requirement_dir(workflow_dir: str) -> Optional[str]:
    """查找最新的需求目录"""
    requirement_dirs = []
    
    for item in os.listdir(workflow_dir):
        item_path = os.path.join(workflow_dir, item)
        if os.path.isdir(item_path) and re.match(REQUIREMENT_DIR_PATTERN, item):
            mtime = os.path.getmtime(item_path)
            requirement_dirs.append((item, item_path, mtime))
    
    if not requirement_dirs:
        return None
    
    requirement_dirs.sort(key=lambda x: x[2], reverse=True)
    return requirement_dirs[0][1]

def get_project_root(workflow_dir: str) -> str:
    """获取项目根目录"""
    return os.path.dirname(workflow_dir)

# ============================================
# Git Diff 解析
# ============================================

def get_changed_files(release_branch: str) -> List[str]:
    """
    获取与 release 分支对比的变更文件列表
    只返回 .go 文件，排除测试文件
    """
    # 先 fetch 确保有最新的 release（静默模式，不要求输入密码）
    # GIT_TERMINAL_PROMPT=0 阻止 Git 请求密码输入
    run_command('GIT_TERMINAL_PROMPT=0 git fetch -q origin release:refs/remotes/origin/release 2>/dev/null || true', check=False)
    
    # 检查 release 分支是否存在
    branch_exists = run_command(f'git rev-parse --verify {release_branch} 2>/dev/null', check=False)
    if not branch_exists:
        print(f"Warning: Release branch '{release_branch}' not found", file=sys.stderr)
        # 尝试使用 HEAD~10 作为基准
        print("Falling back to HEAD~10 as base", file=sys.stderr)
        release_branch = 'HEAD~10'
    
    # 获取变更的文件列表
    diff_output = run_command(f'git diff {release_branch} --name-only -- "*.go"', check=False)
    
    if not diff_output:
        return []
    
    files = []
    for line in diff_output.split('\n'):
        line = line.strip()
        if line and re.search(GO_FILE_PATTERN, line) and not re.search(TEST_FILE_PATTERN, line):
            files.append(line)
    
    return files

def get_changed_lines_for_file(file_path: str, release_branch: str) -> List[int]:
    """
    获取指定文件中变更的行号列表
    使用 git diff 的 unified=0 格式解析
    """
    # 检查 release 分支是否存在
    branch_exists = run_command(f'git rev-parse --verify {release_branch} 2>/dev/null', check=False)
    if not branch_exists:
        release_branch = 'HEAD~10'
    
    diff_output = run_command(
        f'git diff {release_branch} --unified=0 -- "{file_path}"',
        check=False
    )
    
    if not diff_output:
        return []
    
    changed_lines = []
    
    # 解析 @@ -start,count +start,count @@ 格式
    hunk_pattern = r'@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@'
    
    for match in re.finditer(hunk_pattern, diff_output):
        start_line = int(match.group(1))
        count = int(match.group(2)) if match.group(2) else 1
        
        for i in range(count):
            changed_lines.append(start_line + i)
    
    return changed_lines

def get_all_changed_lines(release_branch: str) -> Dict[str, List[int]]:
    """获取所有变更文件的变更行号"""
    changed_files = get_changed_files(release_branch)
    result = {}
    
    for file_path in changed_files:
        lines = get_changed_lines_for_file(file_path, release_branch)
        if lines:
            result[file_path] = lines
    
    return result

# ============================================
# Coverage.out 解析
# ============================================

def parse_coverage_profile(coverage_file: str) -> List[Dict[str, Any]]:
    """
    解析 coverage.out 文件
    格式: mode: set
          package/file.go:startLine.startCol,endLine.endCol numStatements count
    """
    if not os.path.exists(coverage_file):
        print(f"Warning: Coverage file not found: {coverage_file}", file=sys.stderr)
        return []
    
    coverage_data = []
    
    try:
        with open(coverage_file, 'r', encoding='utf-8') as f:
            for line in f:
                line = line.strip()
                
                # 跳过 mode 行
                if line.startswith('mode:') or not line:
                    continue
                
                # 解析格式: file:startLine.startCol,endLine.endCol numStatements count
                match = re.match(
                    r'^(.+):(\d+)\.(\d+),(\d+)\.(\d+)\s+(\d+)\s+(\d+)$',
                    line
                )
                
                if match:
                    file_path = match.group(1)
                    start_line = int(match.group(2))
                    end_line = int(match.group(4))
                    num_statements = int(match.group(6))
                    count = int(match.group(7))
                    
                    # 保留原始路径，路径匹配将在后续使用后缀匹配法处理
                    coverage_data.append({
                        'filePath': file_path,
                        'startLine': start_line,
                        'endLine': end_line,
                        'numStatements': num_statements,
                        'count': count
                    })
    
    except Exception as e:
        print(f"Error parsing coverage file: {e}", file=sys.stderr)
        return []
    
    return coverage_data

class PathMatcher:
    """
    路径匹配器：使用后缀匹配法解决 coverage.out 路径与 git diff 路径不匹配的问题
    
    例如:
    - coverage.out 路径: automation/infra/configs/router_config.go
    - git diff 路径: infra/configs/router_config.go
    
    通过后缀匹配，可以自动适配任何项目结构，无需维护已知目录列表
    """
    
    def __init__(self, coverage_paths: List[str]):
        """
        初始化路径匹配器
        
        Args:
            coverage_paths: coverage.out 中的所有文件路径列表
        """
        self._coverage_paths = set(coverage_paths)
        # 构建后缀索引: 后缀 -> 原始路径
        # 例如: "infra/configs/file.go" -> "automation/infra/configs/file.go"
        self._suffix_index: Dict[str, str] = {}
        
        for path in coverage_paths:
            # 为每个路径生成所有可能的后缀
            parts = path.split('/')
            for i in range(len(parts)):
                suffix = '/'.join(parts[i:])
                # 只保存第一个匹配的路径（避免歧义时的覆盖）
                if suffix not in self._suffix_index:
                    self._suffix_index[suffix] = path
    
    def find_coverage_path(self, git_path: str) -> Optional[str]:
        """
        根据 git diff 路径找到对应的 coverage.out 路径
        
        Args:
            git_path: git diff 返回的相对路径
            
        Returns:
            匹配的 coverage.out 路径，如果没找到返回 None
        """
        # 1. 精确匹配
        if git_path in self._coverage_paths:
            return git_path
        
        # 2. git_path 是 coverage_path 的后缀
        if git_path in self._suffix_index:
            return self._suffix_index[git_path]
        
        # 3. 尝试 git_path 的各种后缀（处理 git_path 可能有额外前缀的情况）
        parts = git_path.split('/')
        for i in range(1, len(parts)):
            suffix = '/'.join(parts[i:])
            if suffix in self._suffix_index:
                return self._suffix_index[suffix]
        
        return None
    
    def get_all_mappings(self) -> Dict[str, str]:
        """获取后缀索引（用于调试）"""
        return self._suffix_index.copy()


def build_coverage_maps(coverage_data: List[Dict]) -> Tuple[Dict[str, Set[int]], Dict[str, Set[int]], PathMatcher]:
    """
    构建覆盖率映射和路径匹配器
    返回: (covered_lines_map, executable_lines_map, path_matcher)
    """
    covered_lines_map: Dict[str, Set[int]] = defaultdict(set)
    executable_lines_map: Dict[str, Set[int]] = defaultdict(set)
    
    # 收集所有文件路径用于创建路径匹配器
    all_paths = set()
    
    for cov in coverage_data:
        file_path = cov['filePath']
        start_line = cov['startLine']
        end_line = cov['endLine']
        count = cov['count']
        
        all_paths.add(file_path)
        
        # 记录所有可执行行
        for line in range(start_line, end_line + 1):
            executable_lines_map[file_path].add(line)
        
        # 记录被覆盖的行
        if count > 0:
            for line in range(start_line, end_line + 1):
                covered_lines_map[file_path].add(line)
    
    # 创建路径匹配器
    path_matcher = PathMatcher(list(all_paths))
    
    return covered_lines_map, executable_lines_map, path_matcher

# ============================================
# 函数级分析
# ============================================

def parse_functions_from_file(file_path: str, project_root: str) -> List[Dict[str, Any]]:
    """
    解析文件中的函数定义
    返回: [{name, startLine, endLine}, ...]
    """
    full_path = os.path.join(project_root, file_path)
    
    if not os.path.exists(full_path):
        return []
    
    functions = []
    
    try:
        with open(full_path, 'r', encoding='utf-8') as f:
            lines = f.readlines()
        
        current_func = None
        brace_count = 0
        
        for i, line in enumerate(lines):
            line_num = i + 1
            
            # 检查函数定义
            func_match = re.match(FUNC_PATTERN, line)
            if func_match and current_func is None:
                current_func = {
                    'name': func_match.group(1),
                    'startLine': line_num,
                    'braceCount': line.count('{') - line.count('}')
                }
                brace_count = current_func['braceCount']
                continue
            
            # 跟踪大括号
            if current_func:
                brace_count += line.count('{') - line.count('}')
                
                if brace_count <= 0:
                    functions.append({
                        'name': current_func['name'],
                        'startLine': current_func['startLine'],
                        'endLine': line_num
                    })
                    current_func = None
                    brace_count = 0
    
    except Exception as e:
        print(f"Error parsing file {file_path}: {e}", file=sys.stderr)
    
    return functions

def get_changed_functions(
    file_path: str,
    changed_lines: List[int],
    project_root: str
) -> List[Dict[str, Any]]:
    """获取包含变更行的函数"""
    functions = parse_functions_from_file(file_path, project_root)
    changed_funcs = []
    
    changed_set = set(changed_lines)
    
    for func in functions:
        func_lines = set(range(func['startLine'], func['endLine'] + 1))
        if func_lines & changed_set:
            changed_funcs.append(func)
    
    return changed_funcs

# ============================================
# 指标计算
# ============================================

def calculate_incremental_coverage(
    changed_lines_map: Dict[str, List[int]],
    covered_lines_map: Dict[str, Set[int]],
    executable_lines_map: Dict[str, Set[int]],
    path_matcher: 'PathMatcher'
) -> Dict[str, Any]:
    """
    计算增量覆盖率
    增量覆盖率 = 被覆盖的增量可执行行数 / 增量可执行行数
    
    使用 PathMatcher 进行路径匹配，解决 coverage.out 路径与 git diff 路径不一致的问题
    """
    total_changed_executable = 0
    total_covered = 0
    file_stats = []
    
    for file_path, changed_lines in changed_lines_map.items():
        # 使用 PathMatcher 查找对应的 coverage 路径
        coverage_path = path_matcher.find_coverage_path(file_path)
        
        if coverage_path:
            covered_set = covered_lines_map.get(coverage_path, set())
            executable_set = executable_lines_map.get(coverage_path, set())
        else:
            # 如果找不到匹配的路径，尝试直接使用原始路径（兼容旧逻辑）
            covered_set = covered_lines_map.get(file_path, set())
            executable_set = executable_lines_map.get(file_path, set())
        
        # 只统计变更行中的可执行行
        executable_changed = 0
        covered_count = 0
        changed_executable_lines = []  # 记录具体的变更可执行行号
        
        for line in changed_lines:
            if line in executable_set:
                executable_changed += 1
                changed_executable_lines.append(line)
                if line in covered_set:
                    covered_count += 1
        
        if executable_changed > 0:
            total_changed_executable += executable_changed
            total_covered += covered_count
            
            coverage_percent = round((covered_count / executable_changed) * 100, 1)
            
            file_stats.append({
                'filePath': file_path,
                'changedLines': executable_changed,
                'coveredLines': covered_count,
                'coveragePercent': coverage_percent,
                'changedLineNumbers': changed_executable_lines  # 新增：具体的变更行号
            })
    
    incremental_percent = 0
    if total_changed_executable > 0:
        incremental_percent = round((total_covered / total_changed_executable) * 100, 1)
    
    return {
        'changedLines': total_changed_executable,
        'coveredChangedLines': total_covered,
        'incrementalCoveragePercent': incremental_percent,
        'changedFiles': file_stats
    }

def calculate_function_coverage(
    changed_lines_map: Dict[str, List[int]],
    covered_lines_map: Dict[str, Set[int]],
    executable_lines_map: Dict[str, Set[int]],
    project_root: str,
    path_matcher: 'PathMatcher'
) -> Dict[str, Any]:
    """
    计算函数级覆盖率（全量覆盖率）
    
    全量覆盖 = 改动函数内的可执行行 + 增量覆盖中函数外的可执行行
    这样确保：全量覆盖行数 >= 增量覆盖行数，避免逻辑混淆
    
    使用 PathMatcher 进行路径匹配，解决 coverage.out 路径与 git diff 路径不一致的问题
    """
    # 第一步：统计改动函数内的可执行行
    total_func_lines = 0
    total_covered_lines = 0
    function_stats = []
    all_function_lines = set()  # 记录所有函数内的行号（用于后续识别函数外的行）
    
    for file_path, changed_lines in changed_lines_map.items():
        # 使用 PathMatcher 查找对应的 coverage 路径
        coverage_path = path_matcher.find_coverage_path(file_path)
        
        if coverage_path:
            covered_set = covered_lines_map.get(coverage_path, set())
            executable_set = executable_lines_map.get(coverage_path, set())
        else:
            # 如果找不到匹配的路径，尝试直接使用原始路径（兼容旧逻辑）
            covered_set = covered_lines_map.get(file_path, set())
            executable_set = executable_lines_map.get(file_path, set())
        
        # 获取变更的函数
        changed_funcs = get_changed_functions(file_path, changed_lines, project_root)
        
        for func in changed_funcs:
            # 计算函数内的可执行行和覆盖行
            func_executable = 0
            func_covered = 0
            
            for line in range(func['startLine'], func['endLine'] + 1):
                # 记录所有函数内的行号
                all_function_lines.add((file_path, line))
                
                if line in executable_set:
                    func_executable += 1
                    if line in covered_set:
                        func_covered += 1
            
            if func_executable > 0:
                total_func_lines += func_executable
                total_covered_lines += func_covered
                
                coverage_percent = round((func_covered / func_executable) * 100, 1)
                
                function_stats.append({
                    'filePath': file_path,
                    'functionName': func['name'],
                    'startLine': func['startLine'],
                    'endLine': func['endLine'],
                    'totalLines': func_executable,
                    'coveredLines': func_covered,
                    'coveragePercent': coverage_percent
                })
    
    # 第二步：统计增量覆盖中函数外的可执行行（避免全量 < 增量）
    outside_func_lines = 0
    outside_covered_lines = 0
    
    for file_path, changed_lines in changed_lines_map.items():
        # 使用 PathMatcher 查找对应的 coverage 路径
        coverage_path = path_matcher.find_coverage_path(file_path)
        
        if coverage_path:
            covered_set = covered_lines_map.get(coverage_path, set())
            executable_set = executable_lines_map.get(coverage_path, set())
        else:
            covered_set = covered_lines_map.get(file_path, set())
            executable_set = executable_lines_map.get(file_path, set())
        
        # 找出在增量统计中，但不在任何函数内的可执行行
        for line in changed_lines:
            if line in executable_set and (file_path, line) not in all_function_lines:
                outside_func_lines += 1
                if line in covered_set:
                    outside_covered_lines += 1
    
    # 合并函数内和函数外的统计
    total_lines_final = total_func_lines + outside_func_lines
    total_covered_final = total_covered_lines + outside_covered_lines
    
    full_coverage_percent = 0
    if total_lines_final > 0:
        full_coverage_percent = round((total_covered_final / total_lines_final) * 100, 1)
    
    # 统计覆盖率分布
    covered_funcs = sum(1 for f in function_stats if f['coveragePercent'] > 0)
    uncovered_funcs = [f for f in function_stats if f['coveragePercent'] == 0]
    
    return {
        'totalLinesInChangedFunctions': total_lines_final,
        'coveredLinesInChangedFunctions': total_covered_final,
        'fullCoveragePercent': full_coverage_percent,
        'changedFunctions': function_stats,
        'changedFunctionCount': len(function_stats),
        'coveredFunctionCount': covered_funcs,
        'uncoveredFunctions': uncovered_funcs,
        # 新增字段：函数外的可执行行统计（用于调试和说明）
        'outsideFunctionLines': outside_func_lines,
        'outsideCoveredLines': outside_covered_lines
    }

def calculate_risk_assessment(
    incremental_stats: Dict[str, Any],
    function_stats: Dict[str, Any]
) -> Dict[str, Any]:
    """
    计算风险评估指标
    """
    # 高风险文件：覆盖率 < 60%
    high_risk_files = [
        f for f in incremental_stats.get('changedFiles', [])
        if f['coveragePercent'] < HIGH_RISK_THRESHOLD
    ]
    
    # 中风险文件：覆盖率 60-80%
    medium_risk_files = [
        f for f in incremental_stats.get('changedFiles', [])
        if HIGH_RISK_THRESHOLD <= f['coveragePercent'] < MEDIUM_RISK_THRESHOLD
    ]
    
    # 未覆盖的函数
    uncovered_functions = function_stats.get('uncoveredFunctions', [])
    
    # 测试债务：所有未覆盖的行数
    test_debt = 0
    for f in incremental_stats.get('changedFiles', []):
        test_debt += f['changedLines'] - f['coveredLines']
    
    # 风险等级判定
    risk_level = 'low'
    if len(high_risk_files) > 0 or incremental_stats.get('incrementalCoveragePercent', 0) < 60:
        risk_level = 'high'
    elif len(medium_risk_files) > 0 or incremental_stats.get('incrementalCoveragePercent', 0) < 80:
        risk_level = 'medium'
    
    return {
        'riskLevel': risk_level,
        'highRiskFiles': high_risk_files,
        'mediumRiskFiles': medium_risk_files,
        'uncoveredFunctions': uncovered_functions,
        'testDebt': test_debt,
        'highRiskFileCount': len(high_risk_files),
        'mediumRiskFileCount': len(medium_risk_files),
        'uncoveredFunctionCount': len(uncovered_functions)
    }

def calculate_all_metrics(
    release_branch: str,
    coverage_file: str,
    project_root: str
) -> Dict[str, Any]:
    """计算所有指标"""
    
    # 1. 获取变更行
    print("Step 1: Getting changed lines from git diff...", file=sys.stderr)
    changed_lines_map = get_all_changed_lines(release_branch)
    
    if not changed_lines_map:
        print("  No changed Go files found", file=sys.stderr)
    else:
        print(f"  Found {len(changed_lines_map)} changed file(s)", file=sys.stderr)
    
    # 2. 解析覆盖率数据
    print(f"\nStep 2: Parsing coverage file: {coverage_file}", file=sys.stderr)
    coverage_data = parse_coverage_profile(coverage_file)
    
    if not coverage_data:
        print("  No coverage data found", file=sys.stderr)
    else:
        print(f"  Parsed {len(coverage_data)} coverage entries", file=sys.stderr)
    
    # 3. 构建覆盖率映射和路径匹配器
    covered_lines_map, executable_lines_map, path_matcher = build_coverage_maps(coverage_data)
    
    print(f"  Built path matcher with {len(path_matcher._coverage_paths)} unique file paths", file=sys.stderr)
    
    # 4. 计算增量覆盖率
    print("\nStep 3: Calculating incremental coverage...", file=sys.stderr)
    incremental_stats = calculate_incremental_coverage(
        changed_lines_map, covered_lines_map, executable_lines_map, path_matcher
    )
    print(f"  Incremental coverage: {incremental_stats['incrementalCoveragePercent']}%", file=sys.stderr)
    print(f"  Changed executable lines: {incremental_stats['changedLines']}", file=sys.stderr)
    print(f"  Covered changed lines: {incremental_stats['coveredChangedLines']}", file=sys.stderr)
    
    # 5. 计算函数级覆盖率
    print("\nStep 4: Calculating function-level coverage...", file=sys.stderr)
    function_stats = calculate_function_coverage(
        changed_lines_map, covered_lines_map, executable_lines_map, project_root, path_matcher
    )
    print(f"  Full coverage: {function_stats['fullCoveragePercent']}%", file=sys.stderr)
    print(f"  Changed functions: {function_stats['changedFunctionCount']}", file=sys.stderr)
    print(f"  Covered functions: {function_stats['coveredFunctionCount']}", file=sys.stderr)
    
    # 6. 计算风险评估
    print("\nStep 5: Calculating risk assessment...", file=sys.stderr)
    risk_stats = calculate_risk_assessment(incremental_stats, function_stats)
    print(f"  Risk level: {risk_stats['riskLevel']}", file=sys.stderr)
    print(f"  High risk files: {risk_stats['highRiskFileCount']}", file=sys.stderr)
    print(f"  Test debt: {risk_stats['testDebt']} lines", file=sys.stderr)
    
    return {
        'version': __version__,
        'timestamp': datetime.now().isoformat(),
        'releaseBranch': release_branch,
        'coverageFile': coverage_file,
        'incrementalCoverage': incremental_stats,
        'functionCoverage': function_stats,
        'riskAssessment': risk_stats,
        'summary': {
            'incrementalCoveragePercent': incremental_stats['incrementalCoveragePercent'],
            'fullCoveragePercent': function_stats['fullCoveragePercent'],
            'changedLines': incremental_stats['changedLines'],
            'coveredLines': incremental_stats['coveredChangedLines'],
            'changedFunctions': function_stats['changedFunctionCount'],
            'coveredFunctions': function_stats['coveredFunctionCount'],
            'riskLevel': risk_stats['riskLevel'],
            'testDebt': risk_stats['testDebt']
        }
    }

# ============================================
# 主函数
# ============================================

def main():
    """主函数"""
    parser = argparse.ArgumentParser(description='Calculate incremental coverage metrics')
    parser.add_argument('--version', '-v', action='version', version=f'v{__version__}')
    parser.add_argument('--release', default=DEFAULT_RELEASE_BRANCH,
                        help=f'Release branch to compare against (default: {DEFAULT_RELEASE_BRANCH})')
    parser.add_argument('coverage_file', nargs='?', default=None,
                        help='Path to coverage.out file (default: auto-detect from tests directory)')
    parser.add_argument('--coverage-file', '-c', dest='coverage_file_opt',
                        help='Path to coverage.out file (alternative to positional argument)')
    parser.add_argument('--output', '-o',
                        help='Output JSON file path (default: print to stdout)')
    parser.add_argument('--quiet', '-q', action='store_true',
                        help='Quiet mode, only output JSON')
    args = parser.parse_args()
    
    if not args.quiet:
        print("=" * 60, file=sys.stderr)
        print(f"Incremental Coverage Calculator v{__version__}", file=sys.stderr)
        print("=" * 60, file=sys.stderr)
    
    # 查找项目根目录
    workflow_dir = find_workflow_dir()
    project_root = get_project_root(workflow_dir)
    
    if not args.quiet:
        print(f"\nProject root: {project_root}", file=sys.stderr)
    
    # 确定覆盖率文件路径
    coverage_file = args.coverage_file or args.coverage_file_opt
    if not coverage_file:
        # 自动查找 tests 目录下的 coverage.out
        req_dir = find_latest_requirement_dir(workflow_dir)
        if req_dir:
            tests_dir = os.path.join(req_dir, 'tests')
            coverage_file = os.path.join(tests_dir, 'coverage.out')
        else:
            coverage_file = os.path.join(project_root, 'coverage.out')
    
    if not args.quiet:
        print(f"Coverage file: {coverage_file}", file=sys.stderr)
        print(f"Release branch: {args.release}", file=sys.stderr)
        print(file=sys.stderr)
    
    # 计算所有指标
    if args.quiet:
        # 静默模式：不打印进度
        import io
        import contextlib
        f = io.StringIO()
        with contextlib.redirect_stderr(f):
            metrics = calculate_all_metrics(args.release, coverage_file, project_root)
    else:
        metrics = calculate_all_metrics(args.release, coverage_file, project_root)
    
    # 输出结果
    json_output = json.dumps(metrics, indent=2, ensure_ascii=False)
    
    if args.output:
        with open(args.output, 'w', encoding='utf-8') as f:
            f.write(json_output)
        if not args.quiet:
            print(f"\n{'=' * 60}", file=sys.stderr)
            print(f"Results saved to: {args.output}", file=sys.stderr)
    else:
        if not args.quiet:
            print(f"\n{'=' * 60}", file=sys.stderr)
            print("Results:", file=sys.stderr)
            print("=" * 60, file=sys.stderr)
        print(json_output)
    
    # 返回退出码（风险等级为 high 时返回 1）
    if metrics['summary']['riskLevel'] == 'high':
        sys.exit(1)
    sys.exit(0)

if __name__ == '__main__':
    main()
