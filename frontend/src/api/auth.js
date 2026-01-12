import request from '../utils/request';

// 用户注册
export const register = (data) => {
  return request.post('/apps/basic/user/create_user', data);
};

// 用户登录
export const login = (data) => {
  // 后端期望的参数格式: client_type, code, passwd
  return request.post('/apps/basic/user/user_login', {
    client_type: 1, // 默认web端
    code: data.email,
    passwd: data.password,
  });
};

// 查询用户列表
export const searchUsers = (data) => {
  return request.post('/apps/basic/user/search_user', data);
};

