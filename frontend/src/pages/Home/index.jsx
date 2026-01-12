import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Statistic, Table, Tag } from 'antd';
import {
  UserOutlined,
  InboxOutlined,
  SwapOutlined,
  FileTextOutlined,
} from '@ant-design/icons';
import { observer } from 'mobx-react-lite';
import { useNavigate } from 'react-router-dom';
import authStore from '../../store/authStore';
import { searchInventory } from '../../api/inventory';
import { searchBorrowTask } from '../../api/borrow';
import './index.css';

const Home = observer(() => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState({
    totalInventory: 0,
    myBorrowTasks: 0,
    pendingApproval: 0,
  });
  const [recentBorrows, setRecentBorrows] = useState([]);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      // 加载库存统计
      const inventoryRes = await searchInventory({});
      const totalInventory = inventoryRes.data?.total || 0;

      // 加载我的借记任务
      const myBorrowsRes = await searchBorrowTask({
        operator: authStore.user?.email,
      });
      const myBorrowTasks = myBorrowsRes.data?.total || 0;

      // 如果是教师，加载待审批数量
      let pendingApproval = 0;
      if (authStore.isTeacher) {
        const pendingRes = await searchBorrowTask({
          task_status: 2, // Allocate 状态
        });
        pendingApproval = pendingRes.data?.total || 0;
      }

      setStats({
        totalInventory,
        myBorrowTasks,
        pendingApproval,
      });

      // 加载最近的借记任务
      setRecentBorrows(myBorrowsRes.data?.list?.slice(0, 5) || []);
    } catch (error) {
      console.error('加载数据失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const taskStatusMap = {
    1: { text: '待分配', color: 'default' },
    2: { text: '已分配', color: 'processing' },
    3: { text: '已审批', color: 'success' },
    4: { text: '已完成', color: 'success' },
    5: { text: '已取消', color: 'error' },
    6: { text: '已拒绝', color: 'error' },
  };

  const columns = [
    {
      title: '任务ID',
      dataIndex: 'task_id',
      key: 'task_id',
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
        const statusInfo = taskStatusMap[status] || { text: '未知', color: 'default' };
        return <Tag color={statusInfo.color}>{statusInfo.text}</Tag>;
      },
    },
  ];

  return (
    <div className="home-container">
      <div style={{ 
        marginBottom: 32, 
        padding: '24px 0',
        borderBottom: '2px solid #f0f0f0'
      }}>
        <h2>欢迎回来，{authStore.user?.name}！</h2>
        <p>
          这是实验设备借记管理系统首页，您可以在这里查看系统概览和最近的操作
        </p>
      </div>

      <Row gutter={[24, 24]} style={{ marginBottom: 32 }}>
        <Col xs={24} sm={24} md={8}>
          <Card 
            hoverable
            style={{
              background: 'linear-gradient(135deg, #667eea15 0%, #764ba215 100%)',
              border: '1px solid #667eea30'
            }}
          >
            <Statistic
              title="库存总数"
              value={stats.totalInventory}
              prefix={<InboxOutlined style={{ fontSize: 28 }} />}
              valueStyle={{ 
                color: '#667eea',
                fontSize: 36,
                fontWeight: 'bold'
              }}
              suffix="件"
            />
          </Card>
        </Col>
        <Col xs={24} sm={24} md={8}>
          <Card 
            hoverable
            style={{
              background: 'linear-gradient(135deg, #1890ff15 0%, #096dd915 100%)',
              border: '1px solid #1890ff30'
            }}
          >
            <Statistic
              title="我的借记任务"
              value={stats.myBorrowTasks}
              prefix={<SwapOutlined style={{ fontSize: 28 }} />}
              valueStyle={{ 
                color: '#1890ff',
                fontSize: 36,
                fontWeight: 'bold'
              }}
              suffix="个"
            />
          </Card>
        </Col>
        {authStore.isTeacher && (
          <Col xs={24} sm={24} md={8}>
            <Card 
              hoverable
              style={{
                background: 'linear-gradient(135deg, #ff4d4f15 0%, #cf132215 100%)',
                border: '1px solid #ff4d4f30'
              }}
            >
              <Statistic
                title="待审批"
                value={stats.pendingApproval}
                prefix={<UserOutlined style={{ fontSize: 28 }} />}
                valueStyle={{ 
                  color: '#ff4d4f',
                  fontSize: 36,
                  fontWeight: 'bold'
                }}
                suffix="个"
              />
            </Card>
          </Col>
        )}
      </Row>

      <Card 
        title={
          <span style={{ fontSize: 18, fontWeight: 600 }}>
            <FileTextOutlined style={{ marginRight: 8 }} />
            最近的借记任务
          </span>
        }
        loading={loading}
        extra={
          <a onClick={() => navigate('/borrow/my-tasks')} style={{ fontWeight: 500 }}>
            查看全部 →
          </a>
        }
      >
        <Table
          columns={columns}
          dataSource={recentBorrows}
          rowKey="task_id"
          pagination={false}
          locale={{ emptyText: '暂无借记任务' }}
        />
      </Card>
    </div>
  );
});

export default Home;

