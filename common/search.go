package common

func Search(length int, f func(index int) bool) int {
	for index := 0; index < length; index++ {
		if f(index) {
			return index
		}
	}
	return -1
}
