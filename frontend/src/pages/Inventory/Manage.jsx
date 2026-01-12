import React, { useState, useEffect } from 'react';
import { Card, Form, Select, InputNumber, Input, Button, message, Tabs } from 'antd';
import { PlusOutlined, MinusOutlined } from '@ant-design/icons';
import { createInventory, decreaseInventory } from '../../api/inventory';
import { searchEquip, searchLab } from '../../api/equip';

const { Option } = Select;
const { TextArea } = Input;

const InventoryManage = () => {
  const [increaseForm] = Form.useForm();
  const [decreaseForm] = Form.useForm();
  const [equipList, setEquipList] = useState([]);
  const [labList, setLabList] = useState([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadEquipList();
    loadLabList();
  }, []);

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

  const handleIncrease = async (values) => {
    setLoading(true);
    try {
      await createInventory({
        equip_id: values.equip_id,
        lab_code: values.lab_code,
        count: values.count,
        description: values.description,
      });
      message.success('增加库存成功');
      increaseForm.resetFields();
    } catch (error) {
      message.error('增加库存失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDecrease = async (values) => {
    setLoading(true);
    try {
      await decreaseInventory({
        equip_id: values.equip_id,
        lab_code: values.lab_code,
        count: values.count,
        reason: values.reason,
      });
      message.success('扣减库存成功');
      decreaseForm.resetFields();
    } catch (error) {
      message.error('扣减库存失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Tabs defaultActiveKey="1">
        <Tabs.TabPane tab={<span><PlusOutlined />增加库存</span>} key="1">
          <Card>
            <Form
              form={increaseForm}
              onFinish={handleIncrease}
              layout="vertical"
              style={{ maxWidth: 600 }}
            >
              <Form.Item
                name="equip_id"
                label="设备"
                rules={[{ required: true, message: '请选择设备' }]}
              >
                <Select placeholder="请选择设备" showSearch optionFilterProp="children">
                  {equipList.map((equip) => (
                    <Option key={equip.equip_id} value={equip.equip_id}>
                      {equip.equip_name}
                    </Option>
                  ))}
                </Select>
              </Form.Item>

              <Form.Item
                name="lab_code"
                label="实验室"
                rules={[{ required: true, message: '请选择实验室' }]}
              >
                <Select placeholder="请选择实验室">
                  {labList.map((lab) => (
                    <Option key={lab.lab_code} value={lab.lab_code}>
                      {lab.lab_name}
                    </Option>
                  ))}
                </Select>
              </Form.Item>

              <Form.Item
                name="count"
                label="数量"
                rules={[{ required: true, message: '请输入数量' }]}
              >
                <InputNumber min={1} style={{ width: '100%' }} placeholder="请输入数量" />
              </Form.Item>

              <Form.Item name="description" label="备注">
                <TextArea rows={3} placeholder="请输入备注" />
              </Form.Item>

              <Form.Item>
                <Button type="primary" htmlType="submit" loading={loading} icon={<PlusOutlined />}>
                  增加库存
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </Tabs.TabPane>

        <Tabs.TabPane tab={<span><MinusOutlined />扣减库存</span>} key="2">
          <Card>
            <Form
              form={decreaseForm}
              onFinish={handleDecrease}
              layout="vertical"
              style={{ maxWidth: 600 }}
            >
              <Form.Item
                name="equip_id"
                label="设备"
                rules={[{ required: true, message: '请选择设备' }]}
              >
                <Select placeholder="请选择设备" showSearch optionFilterProp="children">
                  {equipList.map((equip) => (
                    <Option key={equip.equip_id} value={equip.equip_id}>
                      {equip.equip_name}
                    </Option>
                  ))}
                </Select>
              </Form.Item>

              <Form.Item
                name="lab_code"
                label="实验室"
                rules={[{ required: true, message: '请选择实验室' }]}
              >
                <Select placeholder="请选择实验室">
                  {labList.map((lab) => (
                    <Option key={lab.lab_code} value={lab.lab_code}>
                      {lab.lab_name}
                    </Option>
                  ))}
                </Select>
              </Form.Item>

              <Form.Item
                name="count"
                label="数量"
                rules={[{ required: true, message: '请输入数量' }]}
              >
                <InputNumber min={1} style={{ width: '100%' }} placeholder="请输入数量" />
              </Form.Item>

              <Form.Item
                name="reason"
                label="扣减原因"
                rules={[{ required: true, message: '请输入扣减原因' }]}
              >
                <TextArea rows={3} placeholder="请输入扣减原因" />
              </Form.Item>

              <Form.Item>
                <Button type="primary" danger htmlType="submit" loading={loading} icon={<MinusOutlined />}>
                  扣减库存
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </Tabs.TabPane>
      </Tabs>
    </div>
  );
};

export default InventoryManage;

