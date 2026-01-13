import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Tag, Button, Image, Empty, Spin, message } from 'antd';
import { ArrowLeftOutlined, ToolOutlined, PictureOutlined } from '@ant-design/icons';
import { observer } from 'mobx-react-lite';
import enumStore from '../../store/enumStore';
import { searchEquip } from '../../api/equip';
import ImageCarousel from '../../components/ImageCarousel';
import './index.css';

const EquipmentDetail = observer(() => {
  const { equipId } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [equipData, setEquipData] = useState(null);

  useEffect(() => {
    loadEquipDetail();
  }, [equipId]);

  const loadEquipDetail = async () => {
    setLoading(true);
    try {
      const res = await searchEquip({});
      const equip = res.data?.list?.find(item => item.equip_id === equipId);
      
      if (equip) {
        setEquipData(equip);
      } else {
        message.error('设备不存在');
        navigate('/equipment');
      }
    } catch (error) {
      message.error('加载失败');
    } finally {
      setLoading(false);
    }
  };

  const handleBack = () => {
    navigate('/equipment');
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px 0' }}>
        <Spin size="large" tip="加载中..." />
      </div>
    );
  }

  if (!equipData) {
    return null;
  }

  const images = equipData.images || [];

  return (
    <div className="equipment-detail-page">
      <div className="page-header">
        <Button 
          icon={<ArrowLeftOutlined />} 
          onClick={handleBack}
          size="large"
          style={{ marginBottom: 16 }}
        >
          返回列表
        </Button>
        <h1 className="page-title">
          <ToolOutlined style={{ marginRight: 8 }} />
          设备详情
        </h1>
      </div>

      <div className="detail-content">
        {/* 图片展示区 */}
        <Card 
          title={
            <span>
              <PictureOutlined style={{ marginRight: 8 }} />
              设备图片
            </span>
          }
          className="image-section"
        >
          {images.length > 0 ? (
            <div className="images-grid">
              {/* 主轮播图 */}
              <div className="main-carousel">
                <ImageCarousel 
                  images={images} 
                  width="100%" 
                  height={400}
                  showControls={true}
                />
              </div>
              
              {/* 缩略图列表 */}
              {images.length > 1 && (
                <div className="thumbnail-list">
                  <Image.PreviewGroup>
                    {images.map((url, index) => (
                      <div key={index} className="thumbnail-item">
                        <Image
                          src={url}
                          alt={`设备图片 ${index + 1}`}
                          width={120}
                          height={90}
                          style={{ 
                            objectFit: 'cover',
                            borderRadius: 4,
                            cursor: 'pointer'
                          }}
                          preview={{
                            mask: '查看'
                          }}
                        />
                      </div>
                    ))}
                  </Image.PreviewGroup>
                </div>
              )}
            </div>
          ) : (
            <Empty 
              image={<PictureOutlined style={{ fontSize: 64, color: '#d9d9d9' }} />}
              description="暂无图片"
              style={{ padding: '60px 0' }}
            />
          )}
        </Card>

        {/* 基本信息 */}
        <Card 
          title={
            <span>
              <ToolOutlined style={{ marginRight: 8 }} />
              基本信息
            </span>
          }
          className="info-section"
        >
          <Descriptions bordered column={2}>
            <Descriptions.Item label="设备ID" span={2}>
              <Tag color="blue" style={{ fontSize: 14 }}>
                {equipData.equip_id}
              </Tag>
            </Descriptions.Item>
            
            <Descriptions.Item label="设备名称" span={2}>
              <strong style={{ fontSize: 16 }}>{equipData.equip_name}</strong>
            </Descriptions.Item>
            
            <Descriptions.Item label="设备分类">
              <Tag color="purple" style={{ fontSize: 14 }}>
                {enumStore.getEnumLabel('EquipCategory', equipData.category_id)}
              </Tag>
            </Descriptions.Item>
            
            <Descriptions.Item label="规格型号">
              {equipData.model || '-'}
            </Descriptions.Item>
            
            <Descriptions.Item label="创建人">
              {equipData.creator || '-'}
            </Descriptions.Item>
            
            <Descriptions.Item label="创建时间">
              {equipData.ctime ? new Date(equipData.ctime * 1000).toLocaleString('zh-CN') : '-'}
            </Descriptions.Item>
            
            <Descriptions.Item label="描述" span={2}>
              <div style={{ whiteSpace: 'pre-wrap', color: '#666' }}>
                {equipData.description || '暂无描述'}
              </div>
            </Descriptions.Item>
          </Descriptions>
        </Card>
      </div>
    </div>
  );
});

export default EquipmentDetail;
