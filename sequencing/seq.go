package sequencing

type Sequence interface {
	GetName() string
	GetContents() []byte
}

type SequenceComponent interface {
	GetOutput() <-chan Sequence
}
