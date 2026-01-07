package collection

// 空结构体
var stringExists = struct{}{}

// StringStringSet is the main interface
type StringSet struct {
	// struct为结构体类型的变量
	m map[string]struct{}
}

func NewStringSet(items ...string) *StringSet {
	// 获取StringSet的地址
	s := &StringSet{}
	// 声明map类型的数据结构
	s.m = make(map[string]struct{})
	s.Add(items...)
	return s
}

func (s *StringSet) Add(items ...string) {
	for _, item := range items {
		s.m[item] = stringExists
	}
}

// 添加已存在的元素时，返回false
func (s *StringSet) AddOne(item string) bool {
	if _, ok := s.m[item]; ok {
		return false
	}
	s.m[item] = stringExists
	return true
}

func (s *StringSet) Remove(items ...string) {
	for _, item := range items {
		delete(s.m, item)
	}
}

func (s *StringSet) Contains(item string) bool {
	_, ok := s.m[item]
	return ok
}

func (s *StringSet) Size() int {
	return len(s.m)
}

func (s *StringSet) Clear() {
	s.m = make(map[string]struct{})
}

func (s *StringSet) Equal(other *StringSet) bool {
	// 如果两者Size不相等，就不用比较了
	if s.Size() != other.Size() {
		return false
	}
	// 迭代查询遍历
	for key := range s.m {
		// 只要有一个不存在就返回false
		if !other.Contains(key) {
			return false
		}
	}
	return true
}

func (s *StringSet) IsSubset(other *StringSet) bool {
	if s.Size() > other.Size() {
		return false
	}
	// 迭代遍历
	for key := range s.m {
		if !other.Contains(key) {
			return false
		}
	}
	return true
}

func (s *StringSet) ToSlice() []string {
	results := make([]string, 0)
	for key := range s.m {
		results = append(results, key)
	}
	return results
}

func (s *StringSet) IsEmpty() bool {
	return len(s.m) == 0
}

func (s *StringSet) Copy() *StringSet {
	newStringSet := NewStringSet()
	for key := range s.m {
		newStringSet.Add(key)
	}
	return newStringSet
}

func (s *StringSet) InterSet(other *StringSet) *StringSet { // 交集
	newSet := NewStringSet()
	for key := range other.m {
		if s.Contains(key) {
			newSet.Add(key)
		}
	}
	return newSet
}

func (s *StringSet) InterSlice(other ...string) *StringSet { // 交集
	newSet := NewStringSet()
	for _, key := range other {
		if s.Contains(key) {
			newSet.Add(key)
		}
	}
	return newSet
}

func (s *StringSet) UnionSet(other *StringSet) *StringSet { // 并集
	newSet := other.Copy()
	newSet.Add(s.ToSlice()...)
	return newSet
}

func (s *StringSet) UnionSlice(other ...string) *StringSet { // 并集
	newSet := s.Copy()
	newSet.Add(other...)
	return newSet
}

func (s *StringSet) DiffSet(other *StringSet) *StringSet { // 差集
	newSet := s.Copy()
	newSet.Remove(other.ToSlice()...)
	return newSet
}

func (s *StringSet) DiffSlice(others ...string) *StringSet { // 差集
	newSet := s.Copy()
	newSet.Remove(others...)
	return newSet
}
