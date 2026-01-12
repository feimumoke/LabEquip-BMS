import React, { useState, useEffect } from 'react';
import { Table, Tag, message } from 'antd';
import { searchInventoryTask } from '../../api/inventory';

const InventoryTasks = () => {
  const [loading, setLoading] = useState(false);
  const [dataSource, setDataSource] = useState([]);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const res = await searchInventoryTask({});
      setDataSource(res.data?.list || []);
    } catch (error) {
      message.error('加载失败');
    } finally {
      setLoading(false);
    }
  };

  const taskTypeMap = {
    1: '增加',
    2: '扣减',
  };

  const columns = [
    {
      title: '任务ID',
      dataIndex: 'task_id',
      key: 'task_id',
      width: 180,
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
      title: '任务类型',
      dataIndex: 'task_type',
      key: 'task_type',
      render: (type) => (
        <Tag color={type === 1 ? 'green' : 'red'}>{taskTypeMap[type]}</Tag>
      ),
    },
    {
      title: '操作数量',
      dataIndex: 'total_qty',
      key: 'total_qty',
    },
    {
      title: '总库存',
      dataIndex: 'total_qty',
      key: 'total_qty_after',
    },
    {
      title: '可用库存',
      dataIndex: 'available_qty',
      key: 'available_qty',
    },
    {
      title: '已借出',
      dataIndex: 'borrowed_qty',
      key: 'borrowed_qty',
    },
  ];

  return (
    <div>
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
                  {log.operator}：{log.message}
                </p>
              ))}
            </div>
          ),
        }}
      />
    </div>
  );
};

export default InventoryTasks;

