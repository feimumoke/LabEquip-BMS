import React, { useEffect, useState } from 'react';
import { Card, Table, Tag, Button, Space, message } from 'antd';
import { observer } from 'mobx-react-lite';
import enumStore from '../../store/enumStore';

/**
 * 枚举测试页面
 * 用于验证枚举是否正确加载和显示
 */
const EnumTest = observer(() => {
  const [testResults, setTestResults] = useState([]);

  useEffect(() => {
    runTests();
  }, []);

  const runTests = () => {
    const results = [];

    // 测试 1: 检查枚举是否已加载
    results.push({
      test: '枚举是否已加载',
      result: enumStore.isLoaded ? '✅ 已加载' : '❌ 未加载',
      status: enumStore.isLoaded ? 'success' : 'error',
    });

    // 测试 2: 检查 BorrowTaskStatus 枚举是否存在
    const borrowTaskStatusEnum = enumStore.getEnum('BorrowTaskStatus');
    const hasBorrowTaskStatus = Object.keys(borrowTaskStatusEnum).length > 0;
    results.push({
      test: 'BorrowTaskStatus 枚举是否存在',
      result: hasBorrowTaskStatus ? '✅ 存在' : '❌ 不存在',
      status: hasBorrowTaskStatus ? 'success' : 'error',
      details: JSON.stringify(borrowTaskStatusEnum, null, 2),
    });

    // 测试 3: 测试各个状态值的显示
    const statusTests = [
      { value: 1, expected: '待审批' },
      { value: 2, expected: '已审批' },
      { value: 3, expected: '已分配' },
      { value: 4, expected: '进行中' },
      { value: 7, expected: '已拒绝' },
      { value: 8, expected: '已归还' },
      { value: 9, expected: '已取消' },
    ];

    statusTests.forEach(({ value, expected }) => {
      const label = enumStore.getEnumLabel('BorrowTaskStatus', value);
      const isCorrect = label === expected;
      results.push({
        test: `状态值 ${value} 的显示`,
        result: isCorrect ? `✅ ${label}` : `❌ ${label} (期望: ${expected})`,
        status: isCorrect ? 'success' : 'error',
      });
    });

    // 测试 4: 检查所有可用的枚举
    const allEnumKeys = Object.keys(enumStore.enums);
    results.push({
      test: '所有可用的枚举',
      result: allEnumKeys.length > 0 ? `✅ ${allEnumKeys.length} 个枚举` : '❌ 无枚举',
      status: allEnumKeys.length > 0 ? 'success' : 'error',
      details: allEnumKeys.join(', '),
    });

    setTestResults(results);
  };

  const handleReloadEnums = async () => {
    try {
      enumStore.isLoaded = false; // 强制重新加载
      await enumStore.loadEnums();
      message.success('枚举重新加载成功');
      runTests();
    } catch (error) {
      message.error('枚举重新加载失败');
    }
  };

  const columns = [
    {
      title: '测试项',
      dataIndex: 'test',
      key: 'test',
      width: 250,
    },
    {
      title: '结果',
      dataIndex: 'result',
      key: 'result',
      render: (text, record) => (
        <Tag color={record.status === 'success' ? 'green' : 'red'}>
          {text}
        </Tag>
      ),
    },
    {
      title: '详情',
      dataIndex: 'details',
      key: 'details',
      render: (text) => text ? <pre style={{ margin: 0, fontSize: 12 }}>{text}</pre> : '-',
    },
  ];

  const statusDemoData = [
    { id: 1, name: '任务A', status: 1 },
    { id: 2, name: '任务B', status: 2 },
    { id: 3, name: '任务C', status: 3 },
    { id: 4, name: '任务D', status: 4 },
    { id: 5, name: '任务E', status: 7 },
    { id: 6, name: '任务F', status: 8 },
    { id: 7, name: '任务G', status: 9 },
  ];

  const demoColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
    },
    {
      title: '任务名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '状态值',
      dataIndex: 'status',
      key: 'status',
      render: (status) => <Tag>{status}</Tag>,
    },
    {
      title: '状态显示',
      dataIndex: 'status',
      key: 'statusLabel',
      render: (status) => {
        const label = enumStore.getEnumLabel('BorrowTaskStatus', status);
        const isNumber = /^\d+$/.test(label);
        return (
          <Tag color={isNumber ? 'red' : 'green'}>
            {label} {isNumber && '❌ 显示了数字！'}
          </Tag>
        );
      },
    },
  ];

  return (
    <div className="page-container">
      <div className="page-header">
        <h1 className="page-title">枚举测试页面</h1>
        <p className="page-description">
          验证枚举是否正确加载和显示
        </p>
      </div>

      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Card
          title="操作"
          extra={
            <Button type="primary" onClick={handleReloadEnums}>
              重新加载枚举
            </Button>
          }
        >
          <Space>
            <Button onClick={runTests}>重新测试</Button>
            <Button onClick={() => console.log('enumStore:', enumStore)}>
              在控制台查看 enumStore
            </Button>
            <Button onClick={() => console.log('enums:', enumStore.enums)}>
              在控制台查看 enums
            </Button>
          </Space>
        </Card>

        <Card title="测试结果">
          <Table
            columns={columns}
            dataSource={testResults}
            rowKey="test"
            pagination={false}
          />
        </Card>

        <Card title="状态显示演示">
          <Table
            columns={demoColumns}
            dataSource={statusDemoData}
            rowKey="id"
            pagination={false}
          />
        </Card>

        <Card title="原始枚举数据">
          <pre style={{ background: '#f5f5f5', padding: 16, borderRadius: 4 }}>
            {JSON.stringify(enumStore.enums, null, 2)}
          </pre>
        </Card>
      </Space>
    </div>
  );
});

export default EnumTest;
