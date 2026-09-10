package main

// Originally from: https://github.com/dustin/go-broadcast/blob/71439988bd91/broadcaster.go
// used under MIT license

type broadcaster struct {
	input chan any
	reg   chan chan<- any
	unreg chan chan<- any

	outputs map[chan<- any]bool
}

// The Broadcaster interface describes the main entry points to
// broadcasters.
type Broadcaster interface {
	// Register a new channel to receive broadcasts
	Register(chan<- any)
	// Unregister a channel so that it no longer receives broadcasts.
	Unregister(chan<- any)
	// Shut this broadcaster down.
	Close() error
	// Submit a new object to all subscribers
	Submit(any)
	// Try Submit a new object to all subscribers return false if input chan is fill
	TrySubmit(any) bool
}

func (b *broadcaster) broadcast(m any) {
	for ch := range b.outputs {
		ch <- m
	}
}

func (b *broadcaster) run() {
	for {
		select {
		case m := <-b.input:
			b.broadcast(m)
		case ch, ok := <-b.reg:
			if ok {
				b.outputs[ch] = true
			} else {
				return
			}
		case ch := <-b.unreg:
			delete(b.outputs, ch)
		}
	}
}

// NewBroadcaster creates a new broadcaster with the given input
// channel buffer length.
func NewBroadcaster(buflen int) Broadcaster {
	b := &broadcaster{
		input:   make(chan any, buflen),
		reg:     make(chan chan<- any),
		unreg:   make(chan chan<- any),
		outputs: make(map[chan<- any]bool),
	}

	go b.run()

	return b
}

func (b *broadcaster) Register(newch chan<- any) {
	b.reg <- newch
}

func (b *broadcaster) Unregister(newch chan<- any) {
	b.unreg <- newch
}

func (b *broadcaster) Close() error {
	close(b.reg)
	close(b.unreg)
	return nil
}

// Submit an item to be broadcast to all listeners.
func (b *broadcaster) Submit(m any) {
	if b != nil {
		b.input <- m
	}
}

// TrySubmit attempts to submit an item to be broadcast, returning
// true iff it the item was broadcast, else false.
func (b *broadcaster) TrySubmit(m any) bool {
	if b == nil {
		return false
	}
	select {
	case b.input <- m:
		return true
	default:
		return false
	}
}
