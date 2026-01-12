import React from 'react';
import { Select } from 'antd';
import { observer } from 'mobx-react-lite';
import enumStore from '../../store/enumStore';

const { Option } = Select;

/**
 * 枚举选择器组件
 * 
 * 用法示例：
 * <EnumSelect 
 *   enumKey="EquipCategory" 
 *   placeholder="请选择设备分类"
 *   value={value}
 *   onChange={onChange}
 * />
 * 
 * @param {string} enumKey - 枚举的key，如 'EquipCategory'
 * @param {string} placeholder - 占位符文本
 * @param {boolean} showSearch - 是否显示搜索框，默认true
 * @param {boolean} allowClear - 是否允许清除，默认false
 * @param {string} size - 尺寸: 'small' | 'middle' | 'large'，默认 'middle'
 * @param {any} value - 选中的值
 * @param {function} onChange - 值变化回调
 * @param {object} ...restProps - 其他 Select 组件支持的属性
 */
const EnumSelect = observer(({
  enumKey,
  placeholder = '请选择',
  showSearch = true,
  allowClear = false,
  size = 'middle',
  value,
  onChange,
  ...restProps
}) => {
  const options = enumStore.getEnumOptions(enumKey);

  return (
    <Select
      placeholder={placeholder}
      showSearch={showSearch}
      allowClear={allowClear}
      size={size}
      value={value}
      onChange={onChange}
      optionFilterProp="children"
      filterOption={(input, option) =>
        (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
      }
      {...restProps}
    >
      {options.map(option => (
        <Option 
          key={option.value} 
          value={option.value} 
          label={option.label}
        >
          {option.label}
        </Option>
      ))}
    </Select>
  );
});

export default EnumSelect;
