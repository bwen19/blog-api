package util

func RemoveDuplicates[T int | int32 | int64 | string](list []T) []T {
	uniqueList := map[T]byte{}
	result := []T{}
	for _, v := range list {
		if _, ok := uniqueList[v]; !ok {
			uniqueList[v] = 0
			result = append(result, v)
		}
	}
	return result
}
