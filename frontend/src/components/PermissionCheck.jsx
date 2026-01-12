import { observer } from 'mobx-react-lite';
import { Navigate } from 'react-router-dom';
import authStore from '../store/authStore';

// 权限检查组件
const PermissionCheck = observer(({ children, requireAdmin = false, requireTeacher = false }) => {
  if (!authStore.isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (requireAdmin && !authStore.isAdmin) {
    return <Navigate to="/403" replace />;
  }

  if (requireTeacher && !authStore.isTeacher) {
    return <Navigate to="/403" replace />;
  }

  return children;
});

export default PermissionCheck;

