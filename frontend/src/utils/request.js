import axios from 'axios';
import { message } from 'antd';

// 创建 axios 实例
const request = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    // 从 localStorage 获取 token
    const token = localStorage.getItem('token');
    const userEmail = localStorage.getItem('userEmail');
    
    // 白名单接口：不需要 token
    const whiteListUrls = [
      '/apps/basic/user/user_login',
      '/apps/basic/user/create_user',
      '/apps/common/enums',
    ];
    
    const isWhiteList = whiteListUrls.some(url => config.url.includes(url));
    
    // 只有非白名单接口且有 token 时才添加认证头
    if (!isWhiteList && token && userEmail) {
      config.headers['Authorization'] = `Bearer ${token}`;
      config.headers['X-User-Email'] = userEmail;
    }
    
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    const res = response.data;
    
    // 统一处理后端返回的错误
    // retcode !== 0 表示业务错误（包括负数）
    if (res.retcode !== undefined && res.retcode !== null && res.retcode !== 0) {
      const errorMsg = res.message || res.msg || res.errmsg || '请求失败';
      
      // 立即弹窗提示错误信息
      message.error(errorMsg, 3); // 显示3秒
      
      // 特殊错误码处理
      // 401: Token 过期或未授权
      if (res.retcode === 401 || res.retcode === 10001) {
        localStorage.removeItem('token');
        localStorage.removeItem('userEmail');
        localStorage.removeItem('userInfo');
        setTimeout(() => {
          window.location.href = '/login';
        }, 1500);
      }
      
      // 403: 权限不足
      if (res.retcode === 403 || res.retcode === 10003) {
        setTimeout(() => {
          window.location.href = '/inventory/list';
        }, 1500);
      }
      
      // 返回 rejected promise，中断后续处理
      const error = new Error(errorMsg);
      error.response = response;
      error.retcode = res.retcode;
      return Promise.reject(error);
    }
    
    // retcode === 0 表示成功，返回完整响应数据
    return res;
  },
  (error) => {
    // HTTP 状态码错误处理
    if (error.response) {
      const { status, data } = error.response;
      let errorMsg = '请求失败';
      
      if (status === 401) {
        errorMsg = '登录已过期，请重新登录';
        localStorage.removeItem('token');
        localStorage.removeItem('userEmail');
        localStorage.removeItem('userInfo');
        setTimeout(() => {
          window.location.href = '/login';
        }, 1500);
      } else if (status === 403) {
        errorMsg = '没有权限访问';
      } else if (status === 404) {
        errorMsg = '请求的资源不存在';
      } else if (status === 500) {
        errorMsg = data?.message || data?.msg || '服务器内部错误';
      } else if (status >= 400 && status < 500) {
        errorMsg = data?.message || data?.msg || '请求参数错误';
      } else if (status >= 500) {
        errorMsg = data?.message || data?.msg || '服务器错误';
      }
      
      message.error(errorMsg);
      return Promise.reject(new Error(errorMsg));
    } else if (error.request) {
      // 请求已发出但没有收到响应
      const errorMsg = '网络错误，请检查网络连接或后端服务是否启动';
      message.error(errorMsg);
      return Promise.reject(new Error(errorMsg));
    } else {
      // 请求配置出错
      const errorMsg = error.message || '请求配置错误';
      message.error(errorMsg);
      return Promise.reject(error);
    }
  }
);

export default request;

