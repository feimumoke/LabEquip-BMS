import React, { useState, useEffect } from 'react';
import { Table, Button, Modal, Form, Input, InputNumber, Select, message, Card, Tag, Space } from 'antd';
import { PlusOutlined, CloseOutlined, CheckCircleOutlined, RollbackOutlined } from '@ant-design/icons';
import { observer } from 'mobx-react-lite';
import authStore from '../../store/authStore';
import enumStore from '../../store/enumStore';
import { createBorrow, cancelBorrow, taskBorrow, returnBorrow, searchBorrowTask } from '../../api/borrow';
import { searchEquip, searchLab } from '../../api/equip';

const { Option } = Select;

const MyTasks = observer(() => {
  const [loading, setLoading] = useState(false);
  const [dataSource, setDataSource] = useState([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [equipList, setEquipList] = useState([]);
  const [labList, setLabList] = useState([]);
  const [form] = Form.useForm();

  useEffect(() => {
    loadData();
    loadEquipList();
    loadLabList();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const res = await searchBorrowTask({
        creator: authStore.user?.email || '',
      });
      setDataSource(res.data?.list || []);
    } catch (error) {
      console.error('加载失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadEquipList = async () => {
    try {
      const res = await searchEquip({});
      setEquipList(res.data?.list || []);
    } catch (error) {
      console.error('加载设备列表失败:', error);
    }
  };

  const loadLabList = async () => {
    try {
      const res = await searchLab({});
      setLabList(res.data?.list || []);
    } catch (error) {
      console.error('加载实验室列表失败:', error);
    }
  };

  const handleCreate = async (values) => {
    try {
      await createBorrow({
        equip_id: values.equip_id,
        lab_code: values.lab_code,
        borrow_qty: values.borrow_qty,
        reason: values.reason || '',
      });
      message.success('创建借记任务成功');
      setModalVisible(false);
      form.resetFields();
      loadData();
    } catch (error) {
      message.error('创建失败');
    }
  };

  const handleCancel = async (record) => {
    Modal.confirm({
      title: '⚠️ 确认取消借记任务',
      content: (
        <div style={{ padding: '16px 0' }}>
          <div style={{ 
            background: '#fff7e6', 
            border: '1px solid #ffd591',
            borderRadius: 8,
            padding: 16,
            marginBottom: 16
          }}>
            <p style={{ margin: 0, color: '#fa8c16', fontWeight: 600 }}>
              取消后将无法恢复，需要重新申请！
            </p>
          </div>
          <p style={{ marginBottom: 8, color: '#666' }}>
            <strong>任务ID：</strong>
            <span style={{ marginLeft: 8 }}>{record.task_id}</span>
          </p>
          <p style={{ marginBottom: 8, color: '#666' }}>
            <strong>设备：</strong>
            <span style={{ marginLeft: 8 }}>{record.equip_name || record.equip_id}</span>
          </p>
          <p style={{ marginBottom: 8, color: '#666' }}>
            <strong>数量：</strong>
            <span style={{ marginLeft: 8, fontWeight: 600 }}>{record.borrow_qty}</span>
          </p>
        </div>
      ),
      okText: '确认取消',
      cancelText: '我再想想',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await cancelBorrow({ task_id: record.task_id });
          message.success('✅ 取消成功');
          loadData();
        } catch (error) {
          message.error('❌ 取消失败');
        }
      },
    });
  };

  const handleTakeBorrow = async (record) => {
    Modal.confirm({
      title: '📦 确认拿走设备',
      content: (
        <div style={{ padding: '16px 0' }}>
          <div style={{ 
            background: '#e6f7ff', 
            border: '1px solid #91d5ff',
            borderRadius: 8,
            padding: 16,
            marginBottom: 16
          }}>
            <p style={{ margin: 0, color: '#1890ff', fontWeight: 600 }}>
              确认从实验室拿走设备后，状态将变更为"进行中"
            </p>
          </div>
          <p style={{ marginBottom: 8, color: '#666' }}>
            <strong>任务ID：</strong>
            <span style={{ marginLeft: 8 }}>{record.task_id}</span>
          </p>
          <p style={{ marginBottom: 8, color: '#666' }}>
            <strong>设备：</strong>
            <span style={{ marginLeft: 8 }}>{record.equip_name || record.equip_id}</span>
          </p>
          <p style={{ marginBottom: 8, color: '#666' }}>
            <strong>实验室：</strong>
            <span style={{ marginLeft: 8 }}>{record.lab_name || record.lab_code}</span>
          </p>
          <p style={{ marginBottom: 8, color: '#666' }}>
            <strong>数量：</strong>
            <span style={{ marginLeft: 8, fontWeight: 600, color: '#1890ff' }}>{record.borrow_qty}</span>
          </p>
          <div style={{ 
            marginTop: 16, 
            padding: 12, 
            background: '#fffbe6', 
            border: '1px solid #ffe58f',
            borderRadius: 4
          }}>
            <p style={{ margin: 0, color: '#d46b08', fontSize: 13 }}>
              💡 提示：拿走设备后请妥善保管，使用完毕后及时归还
            </p>
          </div>
        </div>
      ),
      okText: '确认拿走',
      cancelText: '取消',
      okButtonProps: { type: 'primary' },
      onOk: async () => {
        try {
          await taskBorrow({ 
            borrow_id: record.task_id,
            code_list: [] // 可以传入具体的设备编码，这里传空数组
          });
          message.success('✅ 拿走成功，请妥善保管设备');
          loadData();
        } catch (error) {
          message.error('❌ 操作失败');
        }
      },
    });
  };

  const handleReturn = async (record) => {
    Modal.confirm({
      title: '🔄 确认归还设备',
      content: (
        <div style={{ padding: '16px 0' }}>
          <div style={{ 
            background: '#f6ffed', 
            border: '1px solid #b7eb8f',
            borderRadius: 8,
            padding: 16,
            marginBottom: 16
          }}>
            <p style={{ margin: 0, color: '#52c41a', fontWeight: 600 }}>
              确认归还设备后，状态将变更为"已归还"，任务完成
            </p>
          </div>
          <p style={{ marginBottom: 8, color: '#666' }}>
            <strong>任务ID：</strong>
            <span style={{ marginLeft: 8 }}>{record.task_id}</span>
          </p>
          <p style={{ marginBottom: 8, color: '#666' }}>
            <strong>设备：</strong>
            <span style={{ marginLeft: 8 }}>{record.equip_name || record.equip_id}</span>
          </p>
          <p style={{ marginBottom: 8, color: '#666' }}>
            <strong>实验室：</strong>
            <span style={{ marginLeft: 8 }}>{record.lab_name || record.lab_code}</span>
          </p>
          <p style={{ marginBottom: 8, color: '#666' }}>
            <strong>归还数量：</strong>
            <span style={{ marginLeft: 8, fontWeight: 600, color: '#52c41a' }}>{record.borrow_qty}</span>
          </p>
          <div style={{ 
            marginTop: 16, 
            padding: 12, 
            background: '#e6f7ff', 
            border: '1px solid #91d5ff',
            borderRadius: 4
          }}>
            <p style={{ margin: 0, color: '#1890ff', fontSize: 13 }}>
              💡 提示：请确认设备完好无损后再归还
            </p>
          </div>
        </div>
      ),
      okText: '确认归还',
      cancelText: '取消',
      okButtonProps: { type: 'primary' },
      onOk: async () => {
        try {
          await returnBorrow({ 
            borrow_id: record.task_id,
            return_qty: record.borrow_qty, // 归还数量
            code_list: [] // 可以传入具体的设备编码
          });
          message.success('✅ 归还成功，感谢您的使用');
          loadData();
        } catch (error) {
          message.error('❌ 归还失败');
        }
      },
    });
  };

  // 获取状态标签颜色
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
      render: (text) => <Tag color="blue">{text}</Tag>,
    },
    {
      title: '设备名称',
      dataIndex: 'equip_name',
      key: 'equip_name',
      width: 150,
      render: (text) => text || '-',
    },
    {
      title: '设备ID',
      dataIndex: 'equip_id',
      key: 'equip_id',
      width: 120,
    },
    {
      title: '实验室',
      dataIndex: 'lab_name',
      key: 'lab_name',
      width: 120,
      render: (text) => text || '-',
    },
    {
      title: '借记数量',
      dataIndex: 'borrow_qty',
      key: 'borrow_qty',
      width: 100,
      render: (qty) => <strong>{qty}</strong>,
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
      title: '审批人',
      dataIndex: 'approval',
      key: 'approval',
      render: (text) => text || '-',
    },
    {
      title: '创建时间',
      dataIndex: 'ctime',
      key: 'ctime',
      render: (time) => time ? new Date(time * 1000).toLocaleString() : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          {record.task_status === 1 && ( // 待审批状态可以取消
            <Button 
              type="link" 
              danger 
              icon={<CloseOutlined />}
              onClick={() => handleCancel(record)}
            >
              取消
            </Button>
          )}
          {record.task_status === 3 && ( // 已分配状态可以拿走设备
            <Button 
              type="primary"
              size="small"
              icon={<CheckCircleOutlined />}
              onClick={() => handleTakeBorrow(record)}
            >
              拿走设备
            </Button>
          )}
          {record.task_status === 4 && ( // 进行中状态可以归还
            <Button 
              type="primary"
              size="small"
              icon={<RollbackOutlined />}
              onClick={() => handleReturn(record)}
              style={{ background: '#52c41a', borderColor: '#52c41a' }}
            >
              归还设备
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div className="page-container">
      <div className="page-header">
        <h1 className="page-title">我的借记</h1>
        <p className="page-description">
          查看和管理我的借记任务，可以创建新的借记申请
        </p>
      </div>

      <Card style={{ marginBottom: 24 }}>
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <div></div>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setModalVisible(true)}
            size="large"
          >
            创建借记
          </Button>
        </Space>
      </Card>

      <Card>
        <Table
          columns={columns}
          dataSource={dataSource}
          rowKey="task_id"
          loading={loading}
          scroll={{ x: 1200 }}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>

      <Modal
        title={
          <span style={{ fontSize: 18, fontWeight: 600 }}>
            <PlusOutlined style={{ marginRight: 8 }} />
            创建借记任务
          </span>
        }
        visible={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          form.resetFields();
        }}
        onOk={() => form.submit()}
        width={600}
        okText="创建"
        cancelText="取消"
      >
        <Form form={form} onFinish={handleCreate} layout="vertical">
          <Form.Item
            name="equip_id"
            label="设备"
            rules={[{ required: true, message: '请选择设备' }]}
          >
            <Select placeholder="请选择设备" showSearch size="large">
              {equipList.map(equip => (
                <Option key={equip.equip_id} value={equip.equip_id}>
                  {equip.equip_name} ({equip.equip_id})
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="lab_code"
            label="实验室"
            rules={[{ required: true, message: '请选择实验室' }]}
          >
            <Select placeholder="请选择实验室" showSearch size="large">
              {labList.map(lab => (
                <Option key={lab.lab_code} value={lab.lab_code}>
                  {lab.lab_name} ({lab.lab_code})
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="borrow_qty"
            label="借记数量"
            rules={[{ required: true, message: '请输入借记数量' }]}
          >
            <InputNumber
              min={1}
              placeholder="请输入借记数量"
              style={{ width: '100%' }}
              size="large"
            />
          </Form.Item>

          <Form.Item name="reason" label="借记原因">
            <Input.TextArea rows={4} placeholder="请输入借记原因（可选）" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
});

export default MyTasks;
