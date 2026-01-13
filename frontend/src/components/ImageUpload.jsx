import React, { useState } from 'react';
import { Upload, Modal, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import request from '../utils/request';

const ImageUpload = ({ value = [], onChange, maxCount = 10 }) => {
  const [previewVisible, setPreviewVisible] = useState(false);
  const [previewImage, setPreviewImage] = useState('');
  const [previewTitle, setPreviewTitle] = useState('');
  const [fileList, setFileList] = useState(
    value.map((url, index) => ({
      uid: `-${index}`,
      name: `image-${index}`,
      status: 'done',
      url: url,
    }))
  );

  // 获取 base64 用于预览
  const getBase64 = (file) => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.readAsDataURL(file);
      reader.onload = () => resolve(reader.result);
      reader.onerror = (error) => reject(error);
    });
  };

  // 预览图片
  const handlePreview = async (file) => {
    if (!file.url && !file.preview) {
      file.preview = await getBase64(file.originFileObj);
    }

    setPreviewImage(file.url || file.preview);
    setPreviewVisible(true);
    setPreviewTitle(file.name || file.url.substring(file.url.lastIndexOf('/') + 1));
  };

  // 关闭预览
  const handleCancel = () => setPreviewVisible(false);

  // 自定义上传
  const customUpload = async ({ file, onSuccess, onError }) => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('file_name', file.name); // 添加文件名字段

    try {
      const response = await request.post('/apps/common/upload_file', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });

      if (response.retcode === 0) {
        onSuccess(response.data);
        message.success('上传成功');
      } else {
        onError(new Error(response.message || '上传失败'));
        message.error(response.message || '上传失败');
      }
    } catch (error) {
      onError(error);
      message.error('上传失败');
    }
  };

  // 文件列表变化
  const handleChange = ({ fileList: newFileList }) => {
    setFileList(newFileList);

    // 提取所有已上传成功的图片 URL
    const urls = newFileList
      .filter((file) => file.status === 'done')
      .map((file) => file.response?.url || file.url);  // 使用 url 字段

    // 通知父组件
    if (onChange) {
      onChange(urls);
    }
  };

  // 上传前校验
  const beforeUpload = (file) => {
    const isImage = file.type.startsWith('image/');
    const isVideo = file.type.startsWith('video/');
    
    if (!isImage && !isVideo) {
      message.error('只能上传图片或视频文件！');
      return false;
    }

    // 视频文件大小限制更大
    const maxSize = isVideo ? 100 : 10; // 视频100MB，图片10MB
    const isValidSize = file.size / 1024 / 1024 < maxSize;
    
    if (!isValidSize) {
      message.error(`${isVideo ? '视频' : '图片'}大小不能超过 ${maxSize}MB！`);
      return false;
    }

    return true;
  };

  const uploadButton = (
    <div>
      <PlusOutlined />
      <div style={{ marginTop: 8 }}>上传文件</div>
    </div>
  );

  return (
    <>
      <Upload
        listType="picture-card"
        fileList={fileList}
        onPreview={handlePreview}
        onChange={handleChange}
        customRequest={customUpload}
        beforeUpload={beforeUpload}
        maxCount={maxCount}
        accept="image/*,video/*"
      >
        {fileList.length >= maxCount ? null : uploadButton}
      </Upload>
      <Modal
        visible={previewVisible}
        title={previewTitle}
        footer={null}
        onCancel={handleCancel}
      >
        <img alt="preview" style={{ width: '100%' }} src={previewImage} />
      </Modal>
    </>
  );
};

export default ImageUpload;
