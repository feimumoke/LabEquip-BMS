import React, { useState, useEffect } from 'react';
import { Table, Button, Modal, Form, Input, Select, message, Card, Tag, Space } from 'antd';
import { PlusOutlined, ToolOutlined, SearchOutlined } from '@ant-design/icons';
import { observer } from 'mobx-react-lite';
import authStore from '../../store/authStore';
import enumStore from '../../store/enumStore';
import { searchEquip, createEquip } from '../../api/equip';

const { Option } = Select;

const Equipment = observer(() => {
  const [loading, setLoading] = useState(false);
  const [dataSource, setDataSource] = useState([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const res = await searchEquip({});
      setDataSource(res.data?.list || []);
    } catch (error) {
      message.error('加载失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async (values) => {
    try {
      await createEquip(values);
      message.success('创建成功');
      setModalVisible(false);
      form.resetFields();
      loadData();
    } catch (error) {
      message.error('创建失败');
    }
  };

  const [searchText, setSearchText] = useState('');

  const columns = [
    {
      title: '设备ID',
      dataIndex: 'equip_id',
      key: 'equip_id',
      width: 180,
      render: (text) => <Tag color="blue">{text}</Tag>,
    },
    {
      title: '设备名称',
      dataIndex: 'equip_name',
      key: 'equip_name',
      render: (text) => (
        <Space>
          <ToolOutlined style={{ color: '#667eea' }} />
          <strong>{text}</strong>
        </Space>
      ),
    },
    {
      title: '分类',
      dataIndex: 'category_id',
      key: 'category_id',
      render: (value) => {
        const label = enumStore.getEnumLabel('EquipCategory', value);
        return <Tag color="purple">{label}</Tag>;
      },
    },
    {
      title: '规格型号',
      dataIndex: 'model',
      key: 'model',
      render: (text) => text || '-',
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (text) => (
        <span style={{ color: '#666' }}>{text || '暂无描述'}</span>
      ),
    },
  ];

  const filteredData = dataSource.filter(item => 
    !searchText || 
    item.equip_name?.toLowerCase().includes(searchText.toLowerCase()) ||
    item.equip_id?.toLowerCase().includes(searchText.toLowerCase()) ||
    item.model?.toLowerCase().includes(searchText.toLowerCase())
  );

  return (
    <div className="page-container">
      <div className="page-header">
        <h1 className="page-title">
          <ToolOutlined style={{ marginRight: 8 }} />
          设备管理
        </h1>
        <p className="page-description">
          管理和查看实验室设备信息，包括设备名称、型号、分类等
        </p>
      </div>

      <Card style={{ marginBottom: 24 }}>
        <Space size="middle" style={{ width: '100%', justifyContent: 'space-between', flexWrap: 'wrap' }}>
          <Input
            placeholder="搜索设备名称、ID或型号"
            prefix={<SearchOutlined />}
            style={{ width: 300 }}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            allowClear
          />
          {authStore.isAdmin && (
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setModalVisible(true)}
              size="large"
            >
              新增设备
            </Button>
          )}
        </Space>
      </Card>

      <Card>
        <Table
          columns={columns}
          dataSource={filteredData}
          rowKey="equip_id"
          loading={loading}
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
            新增设备
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
        <Form form={form} onFinish={handleCreate} layout="vertical" className="modal-form">
          <Form.Item
            name="equip_name"
            label="设备名称"
            rules={[{ required: true, message: '请输入设备名称' }]}
          >
            <Input placeholder="请输入设备名称" size="large" />
          </Form.Item>

          <Form.Item
            name="category_id"
            label="设备分类"
            rules={[{ required: true, message: '请选择设备分类' }]}
          >
            <Select
              placeholder="请选择设备分类"
              size="large"
              showSearch
              optionFilterProp="children"
              filterOption={(input, option) =>
                (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
              }
            >
              {enumStore.getEnumOptions('EquipCategory').map(option => (
                <Option key={option.value} value={option.value} label={option.label}>
                  {option.label}
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item name="model" label="规格型号">
            <Input placeholder="请输入规格型号" size="large" />
          </Form.Item>

          <Form.Item name="description" label="描述">
            <Input.TextArea rows={4} placeholder="请输入设备描述信息" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
});

export default Equipment;

