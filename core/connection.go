package dge

type Connection interface {
	Acquirer
	Validate(*GraphContext) error
	Release() error
}

type Acquirer interface {
	Acquire(any) any
}
