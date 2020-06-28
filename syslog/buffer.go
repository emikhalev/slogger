package syslog

import (
	"context"
	"fmt"
	slog "log/syslog"
	"sync"
)

type messageBuffer struct {
	muLock     sync.RWMutex
	buffer     []*bufferRecord
	head, tail int
}

type bufferRecord struct {
	ctx   context.Context
	ts    string
	level slog.Priority
	value string
}

func newMessageBuffer(size int) *messageBuffer {
	size++
	return &messageBuffer{
		buffer: make([]*bufferRecord, size, size),
		head:   0,
		tail:   0,
	}
}

// reset - reset queue
func (mb *messageBuffer) reset() {
	mb.muLock.Lock()
	defer mb.muLock.Unlock()

	mb.head = 0
	mb.tail = 0
}

// add - add element to queue
func (mb *messageBuffer) add(v *bufferRecord) error {
	if mb.size() == mb.len() {
		return fmt.Errorf("buffer is full")
	}

	mb.muLock.Lock()
	defer mb.muLock.Unlock()

	l := len(mb.buffer)
	mb.buffer[mb.tail] = v
	mb.tail = (mb.tail + 1) % l
	return nil
}

// remove - return element from queue and return it
func (mb *messageBuffer) remove() (*bufferRecord, error) {
	// we need Lock (not just RLock), because if mb.head write (should be atomic with read operation)
	if mb.empty() {
		return nil, fmt.Errorf("buffer if empty")
	}

	mb.muLock.Lock()
	defer mb.muLock.Unlock()

	v := mb.buffer[mb.head]
	mb.head = (mb.head + 1) % len(mb.buffer)
	return v, nil
}

// len - return maximum count of elements in queue
func (mb *messageBuffer) len() int {
	return len(mb.buffer) - 1
}

// size - return count of elements in queue
func (mb *messageBuffer) size() int {
	mb.muLock.RLock()
	defer mb.muLock.RUnlock()

	if mb.head > mb.tail {
		return len(mb.buffer) - mb.head + mb.tail
	} else {
		return mb.tail - mb.head
	}
}

// empty - return true if queue is empty
func (mb *messageBuffer) empty() bool {
	mb.muLock.RLock()
	defer mb.muLock.RUnlock()

	return mb.head == mb.tail
}
