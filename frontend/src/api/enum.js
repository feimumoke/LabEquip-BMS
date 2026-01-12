import request from '../utils/request';

// 获取所有枚举值
export const getEnums = () => {
  return request.post('/apps/common/enums', {});
};
