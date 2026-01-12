// 用户角色类型
export const UserRole = {
  SUPER_ADMIN: 1,
  ADMIN: 2,
  TEACHER: 3,
  STUDENT: 4,
};

export const UserRoleText = {
  [UserRole.SUPER_ADMIN]: '超级管理员',
  [UserRole.ADMIN]: '管理员',
  [UserRole.TEACHER]: '教师',
  [UserRole.STUDENT]: '学生',
};

// 获取当前用户信息
export const getCurrentUser = () => {
  const userInfo = localStorage.getItem('userInfo');
  if (!userInfo) return null;
  
  try {
    return JSON.parse(userInfo);
  } catch (e) {
    return null;
  }
};

// 检查是否登录
export const isAuthenticated = () => {
  return !!localStorage.getItem('token') && !!getCurrentUser();
};

// 检查是否是管理员或教师
export const isAdminOrTeacher = () => {
  const user = getCurrentUser();
  if (!user) return false;
  
  return [UserRole.SUPER_ADMIN, UserRole.ADMIN, UserRole.TEACHER].includes(user.role);
};

// 检查是否是教师
export const isTeacher = () => {
  const user = getCurrentUser();
  if (!user) return false;
  
  return [UserRole.SUPER_ADMIN, UserRole.ADMIN, UserRole.TEACHER].includes(user.role);
};

// 检查是否是超级管理员
export const isSuperAdmin = () => {
  const user = getCurrentUser();
  if (!user) return false;
  
  return user.role === UserRole.SUPER_ADMIN;
};

// 登出
export const logout = () => {
  localStorage.removeItem('token');
  localStorage.removeItem('userEmail');
  localStorage.removeItem('userInfo');
  window.location.href = '/login';
};

// 保存登录信息
export const saveAuthInfo = (token, userEmail, userInfo) => {
  localStorage.setItem('token', token);
  localStorage.setItem('userEmail', userEmail);
  localStorage.setItem('userInfo', JSON.stringify(userInfo));
};

