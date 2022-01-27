package object_storage

type ObjectStorage interface {
	Set(key string, data []byte) error
	Get(key string) ([]byte, error)
}
