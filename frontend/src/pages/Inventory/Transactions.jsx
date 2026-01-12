import React, { useState, useEffect } from 'react';
import { Table, Tag, message, Card } from 'antd';
import { observer } from 'mobx-react-lite';
import enumStore from '../../store/enumStore';
import { searchTransaction } from '../../api/inventory';
import dayjs from 'dayjs';

const InventoryTransactions = observer(() => {
  const [loading, setLoading] = useState(false);
  const [dataSource, setDataSource] = useState([]);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const res = await searchTransaction({});
      setDataSource(res.data?.list || []);
    } catch (error) {
      message.error('加载失败');
    } finally {
      setLoading(false);
    }
  };

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
    {
      title: '交易ID',
      dataIndex: 'transaction_id',
      key: 'transaction_id',
      width: 150,
      fixed: 'left',
    },
    {
      title: '单据ID',
      dataIndex: 'sheet_id',
      key: 'sheet_id',
      width: 180,
    },
    {
      title: '设备名称',
      dataIndex: 'equip_name',
      key: 'equip_name',
      width: 150,
      render: (text) => text?.[0] || '-',
    },
    {
      title: '实验室',
      dataIndex: 'lab_name',
      key: 'lab_name',
      width: 120,
    },
    {
      title: '交易类型',
      dataIndex: 'trans_type',
      key: 'trans_type',
      width: 100,
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
      width: 100,
      render: (type) => {
        const label = enumStore.getEnumLabel('TransactionSheetType', type);
        return <span>{label}</span>;
      },
    },
    {
      title: '操作数量',
      dataIndex: 'op_qty',
      key: 'op_qty',
      width: 100,
      render: (qty) => (
        <span style={{ color: qty > 0 ? '#52c41a' : '#ff4d4f', fontWeight: 600 }}>
          {qty > 0 ? `+${qty}` : qty}
        </span>
      ),
    },
    {
      title: '总数',
      dataIndex: 'total_qty',
      key: 'total_qty',
      width: 80,
      render: (qty) => <span style={{ fontWeight: 600 }}>{qty || 0}</span>,
    },
    {
      title: '在手',
      dataIndex: 'on_hand_qty',
      key: 'on_hand_qty',
      width: 80,
      render: (qty) => qty || 0,
    },
    {
      title: '可用',
      dataIndex: 'avail_qty',
      key: 'avail_qty',
      width: 80,
      render: (qty) => <span style={{ color: '#1890ff' }}>{qty || 0}</span>,
    },
    {
      title: '借出',
      dataIndex: 'borrowed_qty',
      key: 'borrowed_qty',
      width: 80,
      render: (qty) => <span style={{ color: '#ff4d4f' }}>{qty || 0}</span>,
    },
    {
      title: '分配',
      dataIndex: 'allocated_qty',
      key: 'allocated_qty',
      width: 80,
      render: (qty) => <span style={{ color: '#faad14' }}>{qty || 0}</span>,
    },
    {
      title: '操作人',
      dataIndex: 'operator',
      key: 'operator',
      width: 120,
    },
    {
      title: '备注',
      dataIndex: 'remark',
      key: 'remark',
      width: 200,
      ellipsis: true,
    },
    {
      title: '时间',
      dataIndex: 'ctime',
      key: 'ctime',
      width: 180,
      fixed: 'right',
      render: (time) => dayjs.unix(time).format('YYYY-MM-DD HH:mm:ss'),
    },
  ];

  return (
    <div className="page-container">
      <div className="page-header">
        <h1 className="page-title">三级账查询</h1>
        <p className="page-description">
          查看所有库存交易记录的详细信息，包括总数、在手、可用、借出、分配等数量
        </p>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={dataSource}
          rowKey="transaction_id"
          loading={loading}
          scroll={{ x: 1800 }}
          pagination={{
            pageSize: 20,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条交易记录`,
            pageSizeOptions: ['10', '20', '50', '100'],
          }}
        />
      </Card>
    </div>
  );
});

export default InventoryTransactions;

