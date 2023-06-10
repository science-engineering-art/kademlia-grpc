package interfaces

type Persistence interface {
	Create(key []byte, data *[]byte) error

	Read(key []byte, start int64, end int64) (data *[]byte, err error)

	Delete(key []byte) error
}
