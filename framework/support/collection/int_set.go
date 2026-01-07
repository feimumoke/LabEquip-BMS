package collection

// 空结构体
var IntExists = struct{}{}

// IntSet is the main interface
type IntSet struct {
	// struct为结构体类型的变量
	m map[int64]struct{}
}

func NewIntSet(items ...int64) *IntSet {
	// 获取IntSet的地址
	s := &IntSet{}
	// 声明map类型的数据结构
	s.m = make(map[int64]struct{})
	s.Add(items...)
	return s
}

func (s *IntSet) Add(items ...int64) {
	for _, item := range items {
		s.m[item] = IntExists
	}
}

func (s *IntSet) Remove(items ...int64) {
	for _, item := range items {
		delete(s.m, item)
	}
}

func (s *IntSet) Contains(item int64) bool {
	_, ok := s.m[item]
	return ok
}

func (s *IntSet) Size() int {
	return len(s.m)
}

func (s *IntSet) Clear() {
	s.m = make(map[int64]struct{})
}

func (s *IntSet) Equal(other *IntSet) bool {
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

func (s *IntSet) IsSubset(other *IntSet) bool {
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

// ToSlice 在长度小于等于 MaxSortedLength 时会有序返回
//
// 这样做主要是为了解决流量回放时顺序不一致问题
func (s *IntSet) ToSlice() []int64 {
	return s.ToSortedSlice(false)
}

func (s *IntSet) toSlice() []int64 {
	results := make([]int64, 0)
	for key := range s.m {
		results = append(results, key)
	}
	return results
}

func (s *IntSet) ToSortedSlice(isReverse bool) []int64 {
	results := s.toSlice()
	SortInt(results)
	if isReverse {
		ReverseInt(results)
	}
	return results
}

func (s *IntSet) IsEmpty() bool {
	return len(s.m) == 0
}

func (s *IntSet) Copy() *IntSet {
	newIntSet := NewIntSet()
	for key := range s.m {
		newIntSet.Add(key)
	}
	return newIntSet
}

func (s *IntSet) InterSet(other *IntSet) *IntSet { // 交集
	newSet := NewIntSet()
	for key := range s.m {
		if other.Contains(key) {
			newSet.Add(key)
		}
	}
	return newSet
}

func (s *IntSet) UnionSet(other *IntSet) *IntSet { // 并集
	newSet := other.Copy()
	newSet.Add(s.toSlice()...)
	return newSet
}

func (s *IntSet) DiffSet(other *IntSet) *IntSet { // 差集
	newSet := s.Copy()
	newSet.Remove(other.toSlice()...)
	return newSet
}
