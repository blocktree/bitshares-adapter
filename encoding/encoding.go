package encoding

type Marshaller interface {
	Marshal(*Encoder) error
}
