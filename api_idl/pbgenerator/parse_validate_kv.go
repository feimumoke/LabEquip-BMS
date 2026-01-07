package pbgenerator

type ParamKeyWord = string

const (
	// int64
	IntGTE ParamKeyWord = "gte" // 大于等于，这个原生不支持，由代码实现转换，不存在到内部的映射
	IntLTE ParamKeyWord = "lte" // 小于等于，这个原生不支持，由代码实现转换
	IntGT  ParamKeyWord = "gt"  // 大于
	IntLT  ParamKeyWord = "lt"  // 小于

	// string
	StringMinLen   ParamKeyWord = "min_len"   // 最小长度
	StringMaxLen   ParamKeyWord = "max_len"   // 最大长度
	StringLen      ParamKeyWord = "len"       // 固定长度
	StringRegex    ParamKeyWord = "regex"     // 正则
	StringNotEmpty ParamKeyWord = "not_empty" // 非空

	// repeated
	RepeatedMinCount ParamKeyWord = "min_count" //数组最小长度
	RepeatedMaxCount ParamKeyWord = "max_count" //数组最大长度

	Default ParamKeyWord = "default" // 默认值支持

)

// 内部校验参数关键字
type ParamKeyWordInternal = string

const (
	IntGTInternal            ParamKeyWordInternal = "int_gt"             //大于
	IntLTInternal            ParamKeyWordInternal = "int_lt"             //小于
	StringRegexInternal      ParamKeyWordInternal = "regex"              //正则
	StringMinLenInternal     ParamKeyWordInternal = "length_gt"          // //最小长度
	StringMaxLenInternal     ParamKeyWordInternal = "length_lt"          //最大长度
	StringLenInternal        ParamKeyWordInternal = "length_eq"          //固定长度
	StringNotEmptyInternal   ParamKeyWordInternal = "string_not_empty"   //字符串不能为空
	RepeatedMinCountInternal ParamKeyWordInternal = "repeated_count_min" //数组最小长度
	RepeatedMaxCountInternal ParamKeyWordInternal = "repeated_count_max" //数组最大长度

)

// 不同类型合法的参数
var legalParamsMap = map[ParamType][]ParamKeyWord{
	Int64:    {IntGTE, IntLTE, IntGT, IntLT},
	String:   {StringMinLen, StringMaxLen, StringLen, StringRegex, StringNotEmpty},
	Repeated: {RepeatedMinCount, RepeatedMaxCount},
}

// 合法参数到内部合法参数的映射
var legalParamsToInternalMap = map[ParamKeyWord]ParamKeyWordInternal{
	StringMinLen:     StringMinLenInternal,
	StringMaxLen:     StringMaxLenInternal,
	StringLen:        StringLenInternal,
	StringRegex:      StringRegexInternal,
	StringNotEmpty:   StringNotEmptyInternal,
	IntGT:            IntGTInternal,
	IntLT:            IntLTInternal,
	RepeatedMinCount: RepeatedMinCountInternal,
	RepeatedMaxCount: RepeatedMaxCountInternal,
}
