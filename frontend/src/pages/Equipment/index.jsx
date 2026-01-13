import React, { useState, useEffect } from 'react';
import { Table, Button, Modal, Form, Input, Select, message, Card, Tag, Space, Image } from 'antd';
import { PlusOutlined, ToolOutlined, SearchOutlined, EditOutlined, PictureOutlined, EyeOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { observer } from 'mobx-react-lite';
import authStore from '../../store/authStore';
import enumStore from '../../store/enumStore';
import { searchEquip, createEquip, updateEquip } from '../../api/equip';
import ImageUpload from '../../components/ImageUpload';
import ImageCarousel from '../../components/ImageCarousel';

const { Option } = Select;

const Equipment = observer(() => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [dataSource, setDataSource] = useState([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [currentEquip, setCurrentEquip] = useState(null);
  const [imagePreviewVisible, setImagePreviewVisible] = useState(false);
  const [previewImages, setPreviewImages] = useState([]);
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
      if (editMode) {
        await updateEquip({ ...values, equip_id: currentEquip.equip_id });
        message.success('更新成功');
      } else {
        await createEquip(values);
        message.success('创建成功');
      }
      setModalVisible(false);
      setEditMode(false);
      setCurrentEquip(null);
      form.resetFields();
      loadData();
    } catch (error) {
      message.error(editMode ? '更新失败' : '创建失败');
    }
  };

  const handleEdit = (record) => {
    setEditMode(true);
    setCurrentEquip(record);
    form.setFieldsValue({
      equip_name: record.equip_name,
      category_id: record.category_id,
      model: record.model,
      description: record.description,
      image_list: record.images || [],
    });
    setModalVisible(true);
  };

  const handleViewImages = (images) => {
    setPreviewImages(images || []);
    setImagePreviewVisible(true);
  };

  const handleViewDetail = (equipId) => {
    navigate(`/equipment/${equipId}`);
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
    {
      title: '设备图片',
      dataIndex: 'images',
      key: 'images',
      width: 160,
      render: (images, record) => {
        if (!images || images.length === 0) {
          return (
            <div style={{ textAlign: 'center', color: '#999', padding: '8px 0' }}>
              无图片
            </div>
          );
        }
        return (
          <div onClick={() => handleViewDetail(record.equip_id)} style={{ cursor: 'pointer' }}>
            <ImageCarousel 
              images={images} 
              width={120} 
              height={80}
              showControls={true}
            />
          </div>
        );
      },
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            icon={<EyeOutlined />}
            onClick={() => handleViewDetail(record.equip_id)}
          >
            详情
          </Button>
          {authStore.isAdmin && (
            <Button
              type="link"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            >
              编辑
            </Button>
          )}
        </Space>
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
            {editMode ? <EditOutlined style={{ marginRight: 8 }} /> : <PlusOutlined style={{ marginRight: 8 }} />}
            {editMode ? '编辑设备' : '新增设备'}
          </span>
        }
        visible={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          setEditMode(false);
          setCurrentEquip(null);
          form.resetFields();
        }}
        onOk={() => form.submit()}
        width={700}
        okText={editMode ? '更新' : '创建'}
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

          <Form.Item
            name="image_list"
            label="设备图片/视频"
            extra="最多上传10个文件，支持图片（JPG、PNG、GIF、WebP 等，最大10MB）和视频（MP4、AVI、MOV 等，最大100MB）"
          >
            <ImageUpload maxCount={10} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="设备图片"
        visible={imagePreviewVisible}
        onCancel={() => setImagePreviewVisible(false)}
        footer={null}
        width={800}
      >
        {previewImages.length > 0 ? (
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 16 }}>
            <Image.PreviewGroup>
              {previewImages.map((url, index) => (
                <Image
                  key={index}
                  width={200}
                  src={url}
                  alt={`设备图片 ${index + 1}`}
                  style={{ objectFit: 'cover', borderRadius: 8 }}
                />
              ))}
            </Image.PreviewGroup>
          </div>
        ) : (
          <div style={{ textAlign: 'center', padding: '40px 0', color: '#999' }}>
            <PictureOutlined style={{ fontSize: 48, marginBottom: 16 }} />
            <p>暂无图片</p>
          </div>
        )}
      </Modal>
    </div>
  );
});

export default Equipment;

