#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
测试用例扫描脚本

版本: 3.0.0
更新日志:
  - 3.0.0: 简化为仅扫描 git diff release 中新增/修改的测试函数
  - 2.2.0: 脚本目录结构调整，与 templates 平级
  - 2.1.0: 扫描结果为空时也更新 test-stats.json，清除旧数据
  - 2.0.0: 移除分支依赖，自动查找最新需求目录，新增 git diff release 对比功能
  - 1.0.0: 初始版本

功能：
1. 自动查找最新的需求目录
2. 从 git diff release 提取新增/修改的测试函数
3. 保存到需求的 tests 目录

使用方法：
    python .workflow/scripts/scan_tests.py
    python .workflow/scripts/scan_tests.py --release origin/release  # 指定 release 分支
"""

# 脚本版本号（用于自动更新检查）
__version__ = '3.0.0'

import os
import re
import json
import subprocess
import sys
import argparse
from datetime import datetime

# ============================================
# 配置
# ============================================

# 默认 release 分支
DEFAULT_RELEASE_BRANCH = 'origin/release'

# 测试文件模式
TEST_FILE_PATTERN = r'_test\.go$'

# 测试函数模式（匹配 func TestXxx( 格式）
TEST_FUNC_PATTERN = r'func\s+(Test\w+)\s*\('

# 需求目录模式
REQUIREMENT_DIR_PATTERN = r'^R-\d{4}'

# ============================================
# 工具函数
# ============================================

def run_command(cmd, check=True):
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
            print(f"Error running command: {cmd}")
            print(f"Error: {e.stderr}")
        return None

def find_workflow_dir():
    """查找 .workflow 目录"""
    current_dir = os.getcwd()
    while current_dir != '/':
        workflow_dir = os.path.join(current_dir, '.workflow')
        if os.path.isdir(workflow_dir):
            return workflow_dir
        current_dir = os.path.dirname(current_dir)
    
    print("Error: .workflow directory not found")
    print("Please run this script from within a project that has a .workflow directory")
    sys.exit(1)

def find_latest_requirement_dir(workflow_dir):
    """
    查找最新的需求目录
    
    按以下规则排序：
    1. 目录名匹配 R-xxxx 格式
    2. 按目录的修改时间倒序排列
    3. 返回最新的需求目录
    """
    requirement_dirs = []
    
    for item in os.listdir(workflow_dir):
        item_path = os.path.join(workflow_dir, item)
        if os.path.isdir(item_path) and re.match(REQUIREMENT_DIR_PATTERN, item):
            # 获取目录修改时间
            mtime = os.path.getmtime(item_path)
            requirement_dirs.append((item, item_path, mtime))
    
    if not requirement_dirs:
        return None
    
    # 按修改时间倒序排序
    requirement_dirs.sort(key=lambda x: x[2], reverse=True)
    
    return requirement_dirs[0][1]  # 返回路径

def get_project_root(workflow_dir):
    """获取项目根目录"""
    return os.path.dirname(workflow_dir)

# ============================================
# 从 git diff release 提取新增/修改的测试函数
# ============================================

def parse_test_functions_from_file(file_path):
    """
    解析文件中的所有测试函数
    返回: [(func_name, line_number), ...]
    """
    if not os.path.exists(file_path):
        return []
    
    test_functions = []
    
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        for match in re.finditer(TEST_FUNC_PATTERN, content):
            func_name = match.group(1)
            line_number = content[:match.start()].count('\n') + 1
            test_functions.append((func_name, line_number))
    
    except Exception as e:
        print(f"Warning: Error parsing file {file_path}: {e}")
    
    return test_functions

def get_tests_from_git_diff_release(release_branch, project_root):
    """
    从 git diff release 中提取新增/修改的测试函数
    
    逻辑：与 release 分支对比，提取新增的测试函数定义（仅从新增行中提取）
    返回: {(file_path, func_name): {'line': line_number}}
    """
    print(f"  Comparing with release branch: {release_branch}")
    
    # 先 fetch 确保有最新的 release
    run_command('git fetch origin release:refs/remotes/origin/release 2>/dev/null || true', check=False)
    
    # 检查 release 分支是否存在
    branch_exists = run_command(f'git rev-parse --verify {release_branch} 2>/dev/null', check=False)
    if not branch_exists:
        print(f"  Warning: Release branch '{release_branch}' not found")
        return {}
    
    # 获取 diff（只显示新增的行）
    diff_output = run_command(f'git diff {release_branch} --unified=0 -- "*.go"', check=False)
    
    if not diff_output:
        print("  No changes found in diff with release")
        return {}
    
    # 解析 diff，提取新增的测试函数定义
    tests = {}
    current_file = None
    
    for line in diff_output.split('\n'):
        # 检测文件变化
        if line.startswith('+++ b/'):
            current_file = line[6:]  # 去掉 '+++ b/' 前缀
            continue
        
        # 跳过非测试文件
        if not current_file or not re.search(TEST_FILE_PATTERN, current_file):
            continue
        
        # 只处理新增的行（以 + 开头，但不是 +++）
        if line.startswith('+') and not line.startswith('+++'):
            line_content = line[1:]  # 去掉 '+' 前缀
            
            # 检查是否是测试函数定义
            match = re.search(TEST_FUNC_PATTERN, line_content)
            if match:
                func_name = match.group(1)
                key = (current_file, func_name)
                if key not in tests:
                    tests[key] = {}
    
    if not tests:
        print("  No new test functions found in diff")
        return tests
    
    print(f"  Found changes in test file(s), extracting test functions...")
    
    # 获取行号信息（从实际文件中读取）
    for (file_path, func_name), info in tests.items():
        full_path = os.path.join(project_root, file_path)
        functions = parse_test_functions_from_file(full_path)
        for fn, ln in functions:
            if fn == func_name:
                info['line'] = ln
                break
    
    print(f"  Extracted {len(tests)} new/modified test function(s)")
    return tests

# ============================================
# 转换为标准格式
# ============================================

def format_tests(tests):
    """
    转换为标准化的测试数据结构
    
    返回标准化的测试数据列表
    """
    result = []
    for (file_path, func_name), info in tests.items():
        # 计算 package 路径
        package_path = './' + os.path.dirname(file_path)
        
        result.append({
            'filePath': file_path,
            'functionName': func_name,
            'line': info.get('line'),
            'source': 'git_diff_release',
            'status': 'pending',
            'rerunCommand': f"go test -v -run {func_name} {package_path}"
        })
    
    return result

# ============================================
# 保存数据
# ============================================

def save_test_data(tests_dir, test_cases, clear_existing=False):
    """保存测试数据到 test-stats.json（合并模式或清除模式）
    
    Args:
        tests_dir: 测试目录路径
        test_cases: 测试用例列表
        clear_existing: 如果为 True，清除旧数据而不是合并
    """
    os.makedirs(tests_dir, exist_ok=True)
    
    stats_file = os.path.join(tests_dir, 'test-stats.json')
    
    # 如果是清除模式，直接创建新数据
    if clear_existing:
        total_count = len(test_cases)
        stats = {
            'version': '1.0',
            'timestamp': datetime.now().isoformat(),
            'gitCommit': run_command('git rev-parse HEAD', check=False) or '',
            'summary': {
                'passedCount': 0,
                'failedCount': 0,
                'totalCount': total_count,
                'passRate': 0,
                'coveragePercent': 0
            },
            'testCases': test_cases
        }
        print(f"Cleared existing data, created new test-stats.json with {total_count} test case(s)")
        
        try:
            with open(stats_file, 'w', encoding='utf-8') as f:
                json.dump(stats, f, indent=2, ensure_ascii=False)
            print(f"\nTest stats saved to: {stats_file}")
            print(f"Total test cases: {stats['summary']['totalCount']}")
            return stats_file
        except Exception as e:
            print(f"Error saving test stats: {e}")
            sys.exit(1)
    
    # 读取现有数据（合并模式）
    existing_stats = None
    if os.path.exists(stats_file):
        try:
            with open(stats_file, 'r', encoding='utf-8') as f:
                existing_stats = json.load(f)
            print(f"Loaded existing test-stats.json")
        except Exception as e:
            print(f"Warning: Failed to load existing test-stats.json: {e}")
    
    # 合并测试用例
    if existing_stats and 'testCases' in existing_stats:
        existing_keys = set()
        for tc in existing_stats['testCases']:
            key = f"{tc.get('filePath', '')}:{tc.get('functionName', '')}"
            existing_keys.add(key)
        
        added_count = 0
        for tc in test_cases:
            key = f"{tc['filePath']}:{tc['functionName']}"
            if key not in existing_keys:
                existing_stats['testCases'].append(tc)
                added_count += 1
        
        print(f"Added {added_count} new test case(s), skipped {len(test_cases) - added_count} existing")
        
        # 更新统计
        total_count = len(existing_stats['testCases'])
        passed_count = sum(1 for tc in existing_stats['testCases'] if tc.get('status') == 'passed')
        failed_count = sum(1 for tc in existing_stats['testCases'] if tc.get('status') == 'failed')
        
        existing_stats['summary']['totalCount'] = total_count
        existing_stats['summary']['passedCount'] = passed_count
        existing_stats['summary']['failedCount'] = failed_count
        existing_stats['summary']['passRate'] = round(passed_count / total_count * 100, 1) if total_count > 0 else 0
        existing_stats['timestamp'] = datetime.now().isoformat()
        
        stats = existing_stats
    else:
        total_count = len(test_cases)
        stats = {
            'version': '1.0',
            'timestamp': datetime.now().isoformat(),
            'gitCommit': run_command('git rev-parse HEAD', check=False) or '',
            'summary': {
                'passedCount': 0,
                'failedCount': 0,
                'totalCount': total_count,
                'passRate': 0,
                'coveragePercent': 0
            },
            'testCases': test_cases
        }
        print(f"Created new test-stats.json with {total_count} test case(s)")
    
    # 保存
    try:
        with open(stats_file, 'w', encoding='utf-8') as f:
            json.dump(stats, f, indent=2, ensure_ascii=False)
        
        print(f"\nTest stats saved to: {stats_file}")
        print(f"Total test cases: {stats['summary']['totalCount']}")
        
        return stats_file
    
    except Exception as e:
        print(f"Error saving test stats: {e}")
        sys.exit(1)

# ============================================
# 主函数
# ============================================

def main():
    """主函数"""
    # 解析命令行参数
    parser = argparse.ArgumentParser(description='Scan test cases from git changes')
    parser.add_argument('--version', '-v', action='version', 
                        version=f'scan_tests.py v{__version__}')
    parser.add_argument('--release', default=DEFAULT_RELEASE_BRANCH,
                        help=f'Release branch to compare against (default: {DEFAULT_RELEASE_BRANCH})')
    args = parser.parse_args()
    
    print("=" * 60)
    print(f"Test Case Scanner v{__version__}")
    print("=" * 60)
    
    # 1. 查找 .workflow 目录
    workflow_dir = find_workflow_dir()
    print(f"\nWorkflow directory: {workflow_dir}")
    
    # 2. 查找最新的需求目录
    req_dir = find_latest_requirement_dir(workflow_dir)
    
    if not req_dir:
        print("\n" + "=" * 60)
        print("ERROR: No requirement directory found!")
        print("=" * 60)
        print("\nPlease run TD analysis first to create a requirement.")
        print("Steps:")
        print("  1. Open the Workflow Sidebar in VS Code/Cursor")
        print("  2. Use 'TD Split Feature' to analyze your technical design document")
        print("  3. This will create a requirement directory (R-xxxx)")
        print("  4. Then run this script again")
        print("\nAlternatively, you can manually create a requirement directory:")
        print(f"  mkdir -p {workflow_dir}/R-0001-your-requirement")
        sys.exit(1)
    
    print(f"Latest requirement directory: {req_dir}")
    
    # 3. 设置 tests 目录
    tests_dir = os.path.join(req_dir, 'tests')
    print(f"Tests directory: {tests_dir}")
    
    # 4. 获取项目根目录
    project_root = get_project_root(workflow_dir)
    print(f"Project root: {project_root}")
    
    # 5. 从 git diff release 提取新增/修改的测试函数
    print(f"\n{'=' * 60}")
    print("Scanning new/modified test functions from git diff")
    print(f"{'=' * 60}")
    
    tests = get_tests_from_git_diff_release(args.release, project_root)
    
    if not tests:
        print("\nNo new or modified test functions found.")
        print("\nTips:")
        print("  - Make sure you have changes in *_test.go files compared to release branch")
        print("  - New test functions must be added with 'func Test*(' signature")
        # 仍然保存空的测试数据，以清除旧数据
        print(f"\n{'=' * 60}")
        print("Saving test data (empty)")
        print(f"{'=' * 60}")
        save_test_data(tests_dir, [], clear_existing=True)
        print(f"\n{'=' * 60}")
        print("Done!")
        print(f"{'=' * 60}")
        sys.exit(0)
    
    # 6. 转换为标准格式
    print(f"\n{'=' * 60}")
    print("Formatting test data")
    print(f"{'=' * 60}")
    
    formatted_tests = format_tests(tests)
    print(f"  Total test functions: {len(formatted_tests)}")
    
    # 7. 保存数据（清除旧数据，只保留本次扫描结果）
    print(f"\n{'=' * 60}")
    print("Saving test data (replacing old data)")
    print(f"{'=' * 60}")
    
    save_test_data(tests_dir, formatted_tests, clear_existing=True)
    
    print(f"\n{'=' * 60}")
    print("Done!")
    print(f"{'=' * 60}")
    print(f"\nYou can now view the test report in the workflow sidebar.")

if __name__ == '__main__':
    main()
