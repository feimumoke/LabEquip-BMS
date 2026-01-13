import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/es/locale/zh_CN';
import MainLayout from './components/Layout/MainLayout';
import PermissionCheck from './components/PermissionCheck';
import Login from './pages/Login';
import { isAuthenticated } from './utils/auth';
import enumStore from './store/enumStore';
import './styles/common.css';

// 懒加载页面组件
const Home = React.lazy(() => import('./pages/Home'));
const Users = React.lazy(() => import('./pages/Users'));
const Equipment = React.lazy(() => import('./pages/Equipment'));
const EquipmentDetail = React.lazy(() => import('./pages/EquipmentDetail'));
const InventoryList = React.lazy(() => import('./pages/Inventory/List'));
const InventoryManage = React.lazy(() => import('./pages/Inventory/Manage'));
const InventoryTasks = React.lazy(() => import('./pages/Inventory/Tasks'));
const InventoryTransactions = React.lazy(() => import('./pages/Inventory/Transactions'));
const BorrowMyTasks = React.lazy(() => import('./pages/Borrow/MyTasks'));
const BorrowAllTasks = React.lazy(() => import('./pages/Borrow/AllTasks'));
const BorrowApprove = React.lazy(() => import('./pages/Borrow/Approve'));
const EnumTest = React.lazy(() => import('./pages/Test/EnumTest'));

function App() {
  // 应用启动时检查登录状态，如果已登录则加载枚举
  useEffect(() => {
    if (isAuthenticated()) {
      console.log('🔄 User is authenticated, loading enums...');
      enumStore.loadEnums();
    }
  }, []);

  return (
    <ConfigProvider locale={zhCN}>
      <BrowserRouter>
        <React.Suspense fallback={<div>Loading...</div>}>
          <Routes>
            <Route path="/login" element={
              isAuthenticated() ? <Navigate to="/inventory/list" replace /> : <Login />
            } />
            
            <Route path="/" element={
              <PermissionCheck>
                <MainLayout />
              </PermissionCheck>
            }>
              <Route index element={<Navigate to="/inventory/list" replace />} />
              <Route path="users" element={
                <PermissionCheck requireTeacher>
                  <Users />
                </PermissionCheck>
              } />
              <Route path="equipment" element={<Equipment />} />
              <Route path="equipment/:equipId" element={<EquipmentDetail />} />
              <Route path="inventory/list" element={<InventoryList />} />
              <Route path="inventory/manage" element={
                <PermissionCheck requireAdmin>
                  <InventoryManage />
                </PermissionCheck>
              } />
              <Route path="inventory/tasks" element={<InventoryTasks />} />
              <Route path="inventory/transactions" element={<InventoryTransactions />} />
              <Route path="borrow/my-tasks" element={<BorrowMyTasks />} />
              <Route path="borrow/all-tasks" element={
                <PermissionCheck requireTeacher>
                  <BorrowAllTasks />
                </PermissionCheck>
              } />
              <Route path="borrow/approve" element={
                <PermissionCheck requireTeacher>
                  <BorrowApprove />
                </PermissionCheck>
              } />
              <Route path="test/enum" element={<EnumTest />} />
            </Route>
            
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </React.Suspense>
      </BrowserRouter>
    </ConfigProvider>
  );
}

export default App;

