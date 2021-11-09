package value_object

// 过滤k-v存储的key的过滤器
type ObjectStorageKeyFilter struct {
	Equal     string
	StartWith string
	EndWith   string
	Contains  string
}
