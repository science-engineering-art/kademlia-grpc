package interfaces

type Persistence interface {
	Create(key []byte, data *[]byte) error

	Read(key []byte, start int32, end int32) (data *[]byte, err error)

	Delete(key []byte) error
}
