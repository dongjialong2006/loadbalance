package types

type Conn interface {
	Done()
	Close()
	AddRef()
	Enable() bool
	Weight() int
	CheckHealth() error
}
