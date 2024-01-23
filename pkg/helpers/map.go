package helpers

type KeyGetter interface {
	GetKey() string
}

func ConvertToMap[T KeyGetter](input []T) map[string]T {
	res := make(map[string]T)
	for _, file := range input {
		res[file.GetKey()] = file
	}
	return res
}
