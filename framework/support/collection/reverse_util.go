package collection

func Reverse(arr []interface{}) {
	length := len(arr)
	var temp interface{}
	for i := 0; i < length/2; i++ {
		temp = arr[i]
		arr[i] = arr[length-1-i]
		arr[length-1-i] = temp
	}
}

func ReverseInt(arr []int64) {
	length := len(arr)
	var temp int64
	for i := 0; i < length/2; i++ {
		temp = arr[i]
		arr[i] = arr[length-1-i]
		arr[length-1-i] = temp
	}
}

func ReverseUInt(arr []uint64) {
	length := len(arr)
	var temp uint64
	for i := 0; i < length/2; i++ {
		temp = arr[i]
		arr[i] = arr[length-1-i]
		arr[length-1-i] = temp
	}
}

func ReverseString(arr []string) {
	length := len(arr)
	var temp string
	for i := 0; i < length/2; i++ {
		temp = arr[i]
		arr[i] = arr[length-1-i]
		arr[length-1-i] = temp
	}
}
