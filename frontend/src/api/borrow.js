import request from '../utils/request';

// 创建借记任务
export const createBorrow = (data) => {
  return request.post('/apps/bms/borrow/create_borrow', data);
};

// 取消借记任务
export const cancelBorrow = (data) => {
  return request.post('/apps/bms/borrow/cancel_borrow', data);
};

// 拿走借记物品
export const taskBorrow = (data) => {
  return request.post('/apps/bms/borrow/task_borrow', data);
};

// 归还借记物品
export const returnBorrow = (data) => {
  return request.post('/apps/bms/borrow/return_borrow', data);
};

// 查询借记任务
export const searchBorrowTask = (data) => {
  return request.post('/apps/bms/borrow/search_borrow_task', data);
};

// 审批借记任务
export const approveBorrow = (data) => {
  return request.post('/apps/bms/borrow/approve_borrow', data);
};
