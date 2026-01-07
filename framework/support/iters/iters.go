package iters

import "github.com/ahmetb/go-linq/v3"

type Iters struct {
	innerIter linq.Query
}

func From(source interface{}) Iters {
	return Iters{innerIter: linq.From(source)}
}

func (i Iters) Where(predicate func(interface{}) bool) Iters {
	return Iters{innerIter: i.innerIter.Where(predicate)}
}

func (i Iters) Select(selector func(interface{}) interface{}) Iters {
	return Iters{innerIter: i.innerIter.Select(selector)}
}

func (i Iters) ToMapByKey(result interface{}, keySelector func(interface{}) interface{}) {
	i.innerIter.ToMapBy(result, keySelector, func(i interface{}) interface{} {
		return i
	})
}

func (i Iters) ToMapBy(result interface{}, keySelector func(interface{}) interface{}, valueSelector func(interface{}) interface{}) {
	i.innerIter.ToMapBy(result, keySelector, valueSelector)
}

func (i Iters) Distinct() Iters {
	return Iters{innerIter: i.innerIter.Distinct()}
}

func (i Iters) Union(i2 Iters) Iters {
	return Iters{innerIter: i.innerIter.Union(i2.innerIter)}
}

func (i Iters) Append(item interface{}) Iters {
	return Iters{innerIter: i.innerIter.Append(item)}
}

func (i Iters) Prepend(item interface{}) Iters {
	return Iters{innerIter: i.innerIter.Prepend(item)}
}

func (i Iters) Concat(i2 Iters) Iters {
	return Iters{innerIter: i.innerIter.Concat(i2.innerIter)}
}

func (i Iters) Sort(less func(i, j interface{}) bool) Iters {
	return Iters{innerIter: i.innerIter.Sort(less)}
}

func (i Iters) Reverse() Iters {
	return Iters{innerIter: i.innerIter.Reverse()}
}

func (i Iters) Contains(value interface{}) bool {
	return i.innerIter.Contains(value)
}

func (i Iters) ToSlice(v interface{}) {
	i.innerIter.ToSlice(v)
}

func (i Iters) Max() interface{} {
	return i.innerIter.Max()
}

func (i Iters) Min() interface{} {
	return i.innerIter.Min()
}

func (i Iters) SumInts() int64 {
	return i.innerIter.SumInts()
}

func (i Iters) SumFloats() float64 {
	return i.innerIter.SumFloats()
}

func (i Iters) Count() int {
	return i.innerIter.Count()
}

func (i Iters) CountWith(predicate func(interface{}) bool) int {
	return i.innerIter.CountWith(predicate)
}

func (i Iters) Average() float64 {
	return i.innerIter.Average()
}

func (i Iters) All(predicate func(interface{}) bool) bool {
	return i.innerIter.All(predicate)
}

func (i Iters) AnyWith(predicate func(interface{}) bool) bool {
	return i.innerIter.AnyWith(predicate)
}

func (i Iters) Intersect(i2 Iters) Iters {
	return Iters{innerIter: i.innerIter.Intersect(i2.innerIter)}
}

func (i Iters) Except(i2 Iters) Iters {
	return Iters{innerIter: i.innerIter.Except(i2.innerIter)}
}
