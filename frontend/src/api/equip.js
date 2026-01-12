import request from '../utils/request';

// 创建设备
export const createEquip = (data) => {
  return request.post('/apps/basic/equip/create_equip', data);
};

// 查询设备
export const searchEquip = (data) => {
  return request.post('/apps/basic/equip/search_equip', data);
};

// 查询实验室
export const searchLab = (data) => {
  return request.post('/apps/basic/lab/search_lab', data);
};

