package lib

// Clone flat map
func CloneMap(v map[string]string) (res map[string]string) {
	res = make(map[string]string, len(v))
	for k, v := range v {
		res[k] = v
	}
	return
}
