import React, { useState, useEffect } from 'react';
import { Table, message } from 'antd';
import { searchInventory } from '../../api/inventory';

const InventoryList = () => {
  const [loading, setLoading] = useState(false);
  const [dataSource, setDataSource] = useState([]);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const res = await searchInventory({});
      setDataSource(res.data?.list || []);
    } catch (error) {
      message.error('加载失败');
    } finally {
      setLoading(false);
    }
  };

  const columns = [
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
      title: '总库存',
      dataIndex: 'total_qty',
      key: 'total_qty',
    },
    {
      title: '可用库存',
      dataIndex: 'available_qty',
      key: 'available_qty',
      render: (text) => <span style={{ color: '#52c41a', fontWeight: 'bold' }}>{text}</span>,
    },
    {
      title: '已借出',
      dataIndex: 'borrowed_qty',
      key: 'borrowed_qty',
    },
    {
      title: '预分配',
      dataIndex: 'pre_allocated_qty',
      key: 'pre_allocated_qty',
    },
    {
      title: '保留',
      dataIndex: 'reserved_qty',
      key: 'reserved_qty',
    },
  ];

  return (
    <div>
      <Table
        columns={columns}
        dataSource={dataSource}
        rowKey={(record) => `${record.lab_code}_${record.equip_id}`}
        loading={loading}
      />
    </div>
  );
};

export default InventoryList;

