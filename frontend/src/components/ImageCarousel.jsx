import React, { useState } from 'react';
import { Carousel, Image, Empty } from 'antd';
import { LeftOutlined, RightOutlined, PictureOutlined } from '@ant-design/icons';
import './ImageCarousel.css';

const ImageCarousel = ({ images = [], width = 200, height = 150, showControls = true }) => {
  const [currentIndex, setCurrentIndex] = useState(0);
  const carouselRef = React.useRef(null);

  if (!images || images.length === 0) {
    return (
      <div style={{ 
        width, 
        height, 
        display: 'flex', 
        alignItems: 'center', 
        justifyContent: 'center',
        backgroundColor: '#f5f5f5',
        borderRadius: 8
      }}>
        <Empty 
          image={<PictureOutlined style={{ fontSize: 32, color: '#d9d9d9' }} />}
          description="暂无图片"
          imageStyle={{ height: 40 }}
        />
      </div>
    );
  }

  const handlePrev = (e) => {
    e.stopPropagation();
    carouselRef.current?.prev();
  };

  const handleNext = (e) => {
    e.stopPropagation();
    carouselRef.current?.next();
  };

  const handleAfterChange = (current) => {
    setCurrentIndex(current);
  };

  return (
    <div className="image-carousel-container" style={{ width, height, position: 'relative' }}>
      <Carousel 
        ref={carouselRef}
        dots={false}
        afterChange={handleAfterChange}
        autoplay={false}
      >
        {images.map((url, index) => (
          <div key={index} className="carousel-image-wrapper">
            <Image
              src={url}
              alt={`设备图片 ${index + 1}`}
              width={width}
              height={height}
              style={{ 
                objectFit: 'cover',
                borderRadius: 8
              }}
              preview={{
                mask: '查看大图'
              }}
            />
          </div>
        ))}
      </Carousel>
      
      {showControls && images.length > 1 && (
        <>
          <button 
            className="carousel-control carousel-control-prev"
            onClick={handlePrev}
            aria-label="上一张"
          >
            <LeftOutlined />
          </button>
          <button 
            className="carousel-control carousel-control-next"
            onClick={handleNext}
            aria-label="下一张"
          >
            <RightOutlined />
          </button>
          <div className="carousel-indicator">
            {currentIndex + 1} / {images.length}
          </div>
        </>
      )}
    </div>
  );
};

export default ImageCarousel;
