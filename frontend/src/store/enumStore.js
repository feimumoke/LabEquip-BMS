import { makeAutoObservable } from 'mobx';
import { getEnums } from '../api/enum';

/**
 * 枚举值管理 Store
 * 
 * 功能：
 * 1. 在应用启动时加载所有枚举值
 * 2. 缓存枚举值，避免重复请求
 * 3. 提供便捷的方法获取枚举选项列表
 */
class EnumStore {
  // 所有枚举值的原始数据
  enums = {};
  
  // 是否已加载
  isLoaded = false;
  
  // 是否正在加载
  isLoading = false;

  constructor() {
    makeAutoObservable(this);
  }

  /**
   * 加载所有枚举值
   */
  async loadEnums() {
    if (this.isLoaded || this.isLoading) {
      console.log('⏭️ Enums already loaded or loading, skipping...');
      return;
    }

    this.isLoading = true;
    try {
      console.log('🔄 Loading enums...');
      const res = await getEnums();
      if (res && res.retcode === 0 && res.data) {
        this.enums = res.data;
        this.isLoaded = true;
        console.log('✅ Enums loaded successfully:', this.enums);
        console.log('📋 Available enum keys:', Object.keys(this.enums));
        
        // 特别输出 BorrowTaskStatus 枚举用于调试
        if (this.enums.BorrowTaskStatus) {
          console.log('📊 BorrowTaskStatus enum:', this.enums.BorrowTaskStatus);
        } else {
          console.warn('⚠️ BorrowTaskStatus not found in enums!');
        }
      } else {
        console.error('❌ Invalid enum response:', res);
      }
    } catch (error) {
      console.error('❌ Failed to load enums:', error);
    } finally {
      this.isLoading = false;
    }
  }

  /**
   * 获取指定枚举的原始数据
   * @param {string} enumKey - 枚举的key，如 'EquipCategory'
   * @returns {object} - 枚举的键值对对象
   */
  getEnum(enumKey) {
    if (!this.isLoaded) {
      console.warn(`⚠️ Enums not loaded yet, trying to get '${enumKey}'`);
    }
    const enumData = this.enums[enumKey];
    if (!enumData) {
      console.warn(`⚠️ Enum '${enumKey}' not found. Available keys:`, Object.keys(this.enums));
    }
    return enumData || {};
  }

  /**
   * 获取枚举选项列表（适用于 Ant Design Select 组件）
   * @param {string} enumKey - 枚举的key，如 'EquipCategory'
   * @returns {Array} - 选项数组 [{label: '化学实验设备', value: 2}, ...]
   */
  getEnumOptions(enumKey) {
    const enumData = this.getEnum(enumKey);
    
    // 如果是数组格式（如 CommonEnumYAndN）
    if (Array.isArray(enumData)) {
      return enumData.map(item => ({
        label: item.key,
        value: item.value,
      }));
    }
    
    // 如果是对象格式（如 EquipCategory）
    return Object.entries(enumData).map(([label, value]) => ({
      label,
      value,
    }));
  }

  /**
   * 根据枚举值获取显示名称
   * @param {string} enumKey - 枚举的key
   * @param {number|string} value - 枚举值
   * @returns {string} - 显示名称
   */
  getEnumLabel(enumKey, value) {
    const enumData = this.getEnum(enumKey);
    
    // 如果枚举数据为空，返回原值
    if (!enumData || Object.keys(enumData).length === 0) {
      console.warn(`⚠️ Enum '${enumKey}' is empty, returning value: ${value}`);
      return String(value);
    }
    
    // 如果是数组格式
    if (Array.isArray(enumData)) {
      const item = enumData.find(e => e.value === value);
      if (!item) {
        console.warn(`⚠️ Value ${value} not found in enum '${enumKey}' (array format)`);
      }
      return item ? item.key : String(value);
    }
    
    // 如果是对象格式，查找对应的key
    for (const [label, val] of Object.entries(enumData)) {
      if (val === value) {
        return label;
      }
    }
    
    console.warn(`⚠️ Value ${value} not found in enum '${enumKey}' (object format). Available:`, enumData);
    return String(value);
  }

  /**
   * 检查枚举值是否存在
   * @param {string} enumKey - 枚举的key
   * @param {number|string} value - 枚举值
   * @returns {boolean}
   */
  hasEnumValue(enumKey, value) {
    const enumData = this.getEnum(enumKey);
    
    if (Array.isArray(enumData)) {
      return enumData.some(e => e.value === value);
    }
    
    return Object.values(enumData).includes(value);
  }
}

export default new EnumStore();
