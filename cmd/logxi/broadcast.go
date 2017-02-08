// Package broadcast implements multi-listener broadcast channels.
//
// To create an unbuffered broadcast channel, just declare a Broadcaster:
//
//     var b broadcaster.Broadcaster
//
// To create a buffered broadcast channel with capacity n, call New:
//
//     b := broadcaster.New(n)
//
// To add a listener to a channel, call Listen and read from Ch:
//
//     l := b.Listen()
//     for v := range l.Ch {
//         // ...
//     }
//
//
// To send to the channel, call Send:
//
//     b.Send("Hello world!")
//     v <- l.Ch // returns interface{}("Hello world!")
//
// To remove a listener, call Close.
//
//     l.Close()
//
// To close the broadcast channel, call Close. Any existing or future listeners
// will read from a closed channel:
//
//     b.Close()
//     v, ok <- l.Ch // returns ok == false
package main

import "sync"

// Broadcaster implements a broadcast channel.
// The zero value is a usable unbuffered channel.
type Broadcaster struct {
	m         sync.Mutex
	listeners map[int]chan<- interface{} // lazy init
	nextID    int
	capacity  int
	closed    bool
}

// New returns a new Broadcaster with the given capacity (0 means unbuffered).
func NewBroadcaster(n int) *Broadcaster {
	return &Broadcaster{capacity: n}
}

// Listener implements a listening endpoint for a broadcast channel.
type Listener struct {
	// Ch receives the broadcast messages.
	Ch <-chan interface{}
	b  *Broadcaster
	id int
}

// Send broadcasts a message to the channel.
// Sending on a closed channel causes a runtime panic.
func (b *Broadcaster) Send(v interface{}) {
	b.m.Lock()
	defer b.m.Unlock()
	if b.closed {
		panic("broadcast: send after close")
	}
	for _, l := range b.listeners {
		l <- v
	}
}

// Close closes the channel, disabling the sending of further messages.
func (b *Broadcaster) Close() {
	b.m.Lock()
	defer b.m.Unlock()
	b.closed = true
	for _, l := range b.listeners {
		close(l)
	}
}

// Listen returns a Listener for the broadcast channel.
func (b *Broadcaster) Listen() *Listener {
	b.m.Lock()
	defer b.m.Unlock()
	if b.listeners == nil {
		b.listeners = make(map[int]chan<- interface{})
	}
	for b.listeners[b.nextID] != nil {
		b.nextID++
	}
	ch := make(chan interface{}, b.capacity)
	if b.closed {
		close(ch)
	}
	b.listeners[b.nextID] = ch
	return &Listener{ch, b, b.nextID}
}

// Close closes the Listener, disabling the receival of further messages.
func (l *Listener) Close() {
	l.b.m.Lock()
	defer l.b.m.Unlock()
	delete(l.b.listeners, l.id)
}
