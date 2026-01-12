import React, { useState, useEffect } from 'react';
import { Table, Tag, Space, Input, Button, Card } from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import { observer } from 'mobx-react-lite';
import enumStore from '../../store/enumStore';
import EnumSelect from '../../components/EnumSelect';
import { searchBorrowTask } from '../../api/borrow';

const AllTasks = observer(() => {
  const [loading, setLoading] = useState(false);
  const [dataSource, setDataSource] = useState([]);
  const [searchParams, setSearchParams] = useState({});

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const res = await searchBorrowTask(searchParams);
      setDataSource(res.data?.list || []);
    } catch (error) {
      console.error('加载失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = () => {
    loadData();
  };

  // 状态颜色映射
  const getStatusColor = (status) => {
    const colorMap = {
      1: 'orange',    // 待审批
      2: 'blue',      // 已审批
      3: 'cyan',      // 已分配
      4: 'green',     // 进行中
      7: 'red',       // 已拒绝
      8: 'default',   // 已归还
      9: 'default',   // 已取消
    };
    return colorMap[status] || 'default';
  };

  const columns = [
    {
      title: '任务ID',
      dataIndex: 'task_id',
      key: 'task_id',
      width: 180,
    },
    {
      title: '操作人',
      dataIndex: 'operator',
      key: 'operator',
    },
    {
      title: '设备名称',
      dataIndex: 'equip_name',
      key: 'equip_name',
      render: (text) => text?.[0] || '-',
    },
    {
      title: '实验室',
      dataIndex: 'lab_name',
      key: 'lab_name',
    },
    {
      title: '借用数量',
      dataIndex: 'borrow_qty',
      key: 'borrow_qty',
    },
    {
      title: '状态',
      dataIndex: 'task_status',
      key: 'task_status',
      render: (status) => {
        const label = enumStore.getEnumLabel('BorrowTaskStatus', status);
        const color = getStatusColor(status);
        return <Tag color={color}>{label}</Tag>;
      },
    },
    {
      title: '创建时间',
      dataIndex: 'ctime',
      key: 'ctime',
      render: (time) => new Date(time * 1000).toLocaleString(),
    },
  ];

  return (
    <div className="page-container">
      <div className="page-header">
        <h1 className="page-title">所有借记任务</h1>
        <p className="page-description">
          查看和管理所有用户的借记任务
        </p>
      </div>

      <Card style={{ marginBottom: 24 }}>
        <Space size="middle" style={{ flexWrap: 'wrap' }}>
          <Input
            placeholder="操作人邮箱"
            onChange={(e) =>
              setSearchParams({ ...searchParams, operator: e.target.value })
            }
            style={{ width: 200 }}
            allowClear
          />
          <EnumSelect
            enumKey="BorrowTaskStatus"
            placeholder="任务状态"
            allowClear
            onChange={(value) =>
              setSearchParams({ ...searchParams, task_status: value })
            }
            style={{ width: 150 }}
          />
          <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>
            搜索
          </Button>
        </Space>
      </Card>

      <Card>
        <Table
        columns={columns}
        dataSource={dataSource}
        rowKey="task_id"
        loading={loading}
        expandable={{
          expandedRowRender: (record) => (
            <div>
              <p style={{ margin: 0 }}>
                <strong>任务日志：</strong>
              </p>
              {record.log_list?.map((log, index) => (
                <p key={index} style={{ marginLeft: 16 }}>
                  {new Date(log.ctime * 1000).toLocaleString()} - {log.operator}：{log.message}
                </p>
              ))}
            </div>
          ),
        }}
        pagination={{
          pageSize: 10,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
      />
      </Card>
    </div>
  );
});

export default AllTasks;

