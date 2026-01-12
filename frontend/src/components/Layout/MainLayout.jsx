import React, { useState } from 'react';
import { Layout, Menu, Avatar, Dropdown, Space, message } from 'antd';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import {
  UserOutlined,
  ToolOutlined,
  InboxOutlined,
  SwapOutlined,
  FileTextOutlined,
  LogoutOutlined,
} from '@ant-design/icons';
import { observer } from 'mobx-react-lite';
import authStore from '../../store/authStore';
import { UserRoleText } from '../../utils/auth';
import './MainLayout.css';

const { Header, Sider, Content } = Layout;

const MainLayout = observer(() => {
  const navigate = useNavigate();
  const location = useLocation();
  const [collapsed, setCollapsed] = useState(false);

  const handleLogout = () => {
    authStore.logout();
    navigate('/login');
  };

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人信息',
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
    },
  ];

  const handleUserMenuClick = ({ key }) => {
    if (key === 'logout') {
      handleLogout();
    } else if (key === 'profile') {
      // 可以跳转到个人信息页面
      message.info('个人信息功能待开发');
    }
  };

  // 根据权限生成菜单
  const getMenuItems = () => {
    const items = [];

    // 用户管理 - 超级管理员和教师
    if (authStore.isTeacher) {
      items.push({
        key: '/users',
        icon: <UserOutlined />,
        label: '用户管理',
      });
    }

    // 设备管理 - 所有人可查看，管理员可编辑
    items.push({
      key: '/equipment',
      icon: <ToolOutlined />,
      label: '设备管理',
    });

    // 库存管理
    items.push({
      key: '/inventory',
      icon: <InboxOutlined />,
      label: '库存管理',
      children: [
        {
          key: '/inventory/list',
          label: '库存查询',
        },
        ...(authStore.isAdmin
          ? [
              {
                key: '/inventory/manage',
                label: '库存操作',
              },
            ]
          : []),
        {
          key: '/inventory/tasks',
          label: '库存任务',
        },
        {
          key: '/inventory/transactions',
          label: '三级账查询',
        },
      ],
    });

    // 借记管理
    items.push({
      key: '/borrow',
      icon: <SwapOutlined />,
      label: '借记管理',
      children: [
        {
          key: '/borrow/my-tasks',
          label: '我的借记',
        },
        ...(authStore.isTeacher
          ? [
              {
                key: '/borrow/all-tasks',
                label: '所有借记',
              },
              {
                key: '/borrow/approve',
                label: '审批管理',
              },
            ]
          : []),
      ],
    });

    return items;
  };

  return (
    <Layout style={{ minHeight: '100vh', background: '#f5f5f5' }}>
      <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed} theme="dark">
        <div className="logo">
          <FileTextOutlined style={{ fontSize: 24, color: '#fff' }} />
          {!collapsed && <span style={{ marginLeft: 10 }}>实验设备管理</span>}
        </div>
        <Menu
          theme="dark"
          selectedKeys={[location.pathname]}
          mode="inline"
          items={getMenuItems()}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout style={{ background: '#f5f5f5' }}>
        <Header style={{ 
          padding: '0 32px', 
          background: '#fff', 
          display: 'flex', 
          justifyContent: 'space-between', 
          alignItems: 'center',
          height: '72px'
        }}>
          <div style={{ 
            fontSize: 22, 
            fontWeight: 700,
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
            backgroundClip: 'text'
          }}>
            实验设备借记管理系统
          </div>
          <Dropdown menu={{ items: userMenuItems, onClick: handleUserMenuClick }} placement="bottomRight">
            <Space style={{ cursor: 'pointer', padding: '8px 16px', borderRadius: '8px', transition: 'all 0.3s' }}>
              <Avatar icon={<UserOutlined />} size="large" />
              <span style={{ fontSize: 15, fontWeight: 500 }}>
                {authStore.user?.name || authStore.user?.email}
                <span style={{ 
                  marginLeft: 8, 
                  color: '#888', 
                  fontSize: 13,
                  padding: '2px 8px',
                  background: '#f0f0f0',
                  borderRadius: '4px'
                }}>
                  {UserRoleText[authStore.user?.role]}
                </span>
              </span>
            </Space>
          </Dropdown>
        </Header>
        <Content style={{ 
          margin: '16px', 
          padding: 32, 
          background: '#fff', 
          minHeight: 280,
          borderRadius: '12px',
          boxShadow: '0 2px 8px rgba(0, 0, 0, 0.08)'
        }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
});

export default MainLayout;

