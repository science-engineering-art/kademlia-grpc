package interfaces

type Persistence interface {
	Create(key []byte, data *[]byte) error

	Read(key []byte) (data *[]byte, err error)

	Delete(key []byte) error
}
