package encoding

type TransactionMarshaller interface {
	MarshalTransaction(*Encoder) error
}
