import React, { useState, useEffect } from 'react';
import { Table, Button, Modal, Input, message, Tag, Space, Card } from 'antd';
import { CheckOutlined, CloseOutlined } from '@ant-design/icons';
import { observer } from 'mobx-react-lite';
import enumStore from '../../store/enumStore';
import { searchBorrowTask, approveBorrow } from '../../api/borrow';

const { TextArea } = Input;

const Approve = observer(() => {
  const [loading, setLoading] = useState(false);
  const [dataSource, setDataSource] = useState([]);
  const [approveModalVisible, setApproveModalVisible] = useState(false);
  const [currentRecord, setCurrentRecord] = useState(null);
  const [approveType, setApproveType] = useState(true); // true=通过, false=拒绝
  const [reason, setReason] = useState('');

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      // 查询待审批状态的任务（status = 1）
      const res = await searchBorrowTask({
        task_status: 1, // Pending 待审批状态
      });
      setDataSource(res.data?.list || []);
    } catch (error) {
      message.error('加载失败');
    } finally {
      setLoading(false);
    }
  };

  const showApproveModal = (record, approved) => {
    setCurrentRecord(record);
    setApproveType(approved);
    setApproveModalVisible(true);
  };

  const handleApprove = async () => {
    // 如果是拒绝操作，必须填写原因
    if (!approveType && !reason.trim()) {
      message.warning('请输入拒绝原因');
      return;
    }

    try {
      await approveBorrow({
        borrow_id: currentRecord.task_id,
        approved: approveType ? 1 : 0,
        reason: reason || (approveType ? '审批通过' : '审批拒绝'),
      });
      message.success(approveType ? '✅ 审批通过' : '❌ 已拒绝');
      setApproveModalVisible(false);
      setReason('');
      loadData();
    } catch (error) {
      message.error('操作失败');
    }
  };

  const columns = [
    {
      title: '任务ID',
      dataIndex: 'task_id',
      key: 'task_id',
      width: 180,
    },
    {
      title: '申请人',
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
        return <Tag color="orange">{label}</Tag>;
      },
    },
    {
      title: '申请时间',
      dataIndex: 'ctime',
      key: 'ctime',
      render: (time) => new Date(time * 1000).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_, record) => (
        <Space>
          <Button
            type="primary"
            icon={<CheckOutlined />}
            onClick={() => showApproveModal(record, true)}
          >
            通过
          </Button>
          <Button
            danger
            icon={<CloseOutlined />}
            onClick={() => showApproveModal(record, false)}
          >
            拒绝
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div className="page-container">
      <div className="page-header">
        <h1 className="page-title">审批管理</h1>
        <p className="page-description">
          审批学生的借记申请，通过后学生可以拿走设备
        </p>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={dataSource}
          rowKey="task_id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条待审批记录`,
          }}
        />
      </Card>

      <Modal
        title={
          <span style={{ fontSize: 18, fontWeight: 600 }}>
            {approveType ? '✅ 审批通过确认' : '❌ 审批拒绝确认'}
          </span>
        }
        visible={approveModalVisible}
        onCancel={() => {
          setApproveModalVisible(false);
          setReason('');
        }}
        onOk={handleApprove}
        okText={approveType ? '确认通过' : '确认拒绝'}
        cancelText="取消"
        width={600}
        okButtonProps={{ 
          danger: !approveType,
          type: approveType ? 'primary' : 'default'
        }}
      >
        <div style={{ padding: '16px 0' }}>
          <div style={{ 
            background: approveType ? '#f6ffed' : '#fff2e8', 
            border: `1px solid ${approveType ? '#b7eb8f' : '#ffbb96'}`,
            borderRadius: 8,
            padding: 16,
            marginBottom: 24
          }}>
            <p style={{ margin: 0, color: approveType ? '#52c41a' : '#fa8c16', fontWeight: 600 }}>
              {approveType ? '⚠️ 请确认是否通过该借记申请' : '⚠️ 请确认是否拒绝该借记申请'}
            </p>
          </div>

          <div style={{ marginBottom: 16 }}>
            <p style={{ marginBottom: 8, color: '#666' }}>
              <strong>申请人：</strong>
              <span style={{ color: '#1890ff', marginLeft: 8 }}>{currentRecord?.operator}</span>
            </p>
            <p style={{ marginBottom: 8, color: '#666' }}>
              <strong>设备名称：</strong>
              <span style={{ marginLeft: 8 }}>{currentRecord?.equip_name?.[0] || '-'}</span>
            </p>
            <p style={{ marginBottom: 8, color: '#666' }}>
              <strong>实验室：</strong>
              <span style={{ marginLeft: 8 }}>{currentRecord?.lab_name || '-'}</span>
            </p>
            <p style={{ marginBottom: 8, color: '#666' }}>
              <strong>借用数量：</strong>
              <span style={{ color: '#f5222d', fontWeight: 600, marginLeft: 8 }}>
                {currentRecord?.borrow_qty}
              </span>
            </p>
          </div>

          <div>
            <label style={{ display: 'block', marginBottom: 8, fontWeight: 600 }}>
              审批意见{!approveType && <span style={{ color: '#f5222d' }}>（拒绝必填）</span>}:
            </label>
            <TextArea
              rows={4}
              placeholder={approveType ? '请输入审批意见（可选）' : '请输入拒绝原因（必填）'}
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              maxLength={200}
              showCount
            />
          </div>

          {!approveType && (
            <div style={{ 
              marginTop: 16, 
              padding: 12, 
              background: '#fffbe6', 
              border: '1px solid #ffe58f',
              borderRadius: 4
            }}>
              <p style={{ margin: 0, color: '#d46b08', fontSize: 13 }}>
                💡 提示：拒绝后将释放已分配的库存，学生可以重新申请
              </p>
            </div>
          )}
        </div>
      </Modal>
    </div>
  );
});

export default Approve;

