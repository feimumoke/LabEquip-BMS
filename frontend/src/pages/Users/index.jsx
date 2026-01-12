import React, { useState, useEffect } from 'react';
import { Table, Tag, message } from 'antd';
import { searchUsers } from '../../api/auth';
import { UserRoleText } from '../../utils/auth';

const Users = () => {
  const [loading, setLoading] = useState(false);
  const [dataSource, setDataSource] = useState([]);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const res = await searchUsers({});
      setDataSource(res.data?.list || []);
    } catch (error) {
      message.error('加载失败');
    } finally {
      setLoading(false);
    }
  };

  const getRoleColor = (role) => {
    const colorMap = {
      1: 'red',      // 超级管理员
      2: 'orange',   // 管理员
      3: 'blue',     // 教师
      4: 'default',  // 学生
    };
    return colorMap[role] || 'default';
  };

  const columns = [
    {
      title: '姓名',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: '手机',
      dataIndex: 'phone',
      key: 'phone',
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (role) => (
        <Tag color={getRoleColor(role)}>{UserRoleText[role]}</Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'ctime',
      key: 'ctime',
      render: (time) => new Date(time * 1000).toLocaleString(),
    },
  ];

  return (
    <div>
      <Table
        columns={columns}
        dataSource={dataSource}
        rowKey="email"
        loading={loading}
      />
    </div>
  );
};

export default Users;

