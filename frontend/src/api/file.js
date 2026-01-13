import request from '../utils/request';

// 批量获取文件信息
export const batchFileInfo = (data) => {
  return request.post('/apps/common/batch_file_info', data);
};

/**
 * 上传文件
 * @param {FormData} formData - 必须包含以下字段：
 *   - file: File对象（必填）
 *   - file_name: 文件名（必填）
 * @example
 * const formData = new FormData();
 * formData.append('file', file);
 * formData.append('file_name', file.name);
 * uploadFile(formData);
 */
export const uploadFile = (formData) => {
  return request.post('/apps/common/upload_file', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
};
