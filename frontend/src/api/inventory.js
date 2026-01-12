import request from '../utils/request';

// 增加库存
export const createInventory = (data) => {
  return request.post('/apps/bms/inventory/create_equip_inv', data);
};

// 查询库存
export const searchInventory = (data) => {
  return request.post('/apps/bms/inventory/search_equip_inv', data);
};

// 扣减库存
export const decreaseInventory = (data) => {
  return request.post('/apps/bms/inventory/decrease_equip_inv', data);
};

// 查询库存任务
export const searchInventoryTask = (data) => {
  return request.post('/apps/bms/inventory/search_inv_task', data);
};

// 查询三级账
export const searchTransaction = (data) => {
  return request.post('/apps/bms/transaction/search_inv_transaction', data);
};

