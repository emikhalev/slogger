package syslog

import (
	"strconv"
	"testing"
)

func TestBuffer(t *testing.T) {
	bSize := 10
	mBuf := newMessageBuffer(bSize)

	v := &bufferRecord{
		level: 3,
		value: "Test message",
	}

	// Buffer: []
	for i := 0; i < bSize; i++ {
		vNew := *v
		vNew.value = strconv.Itoa(i)
		if err := mBuf.add(&vNew); err != nil {
			t.Errorf("expect no error, got: %v", err)
		}
	}

	// Buffer: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
	if err := mBuf.add(v); err == nil {
		t.Errorf("expect error, got no error. size: %v, len: %v", mBuf.head, mBuf.tail)
	}

	// Buffer: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
	if v, err := mBuf.remove(); err != nil {
		t.Errorf("remove: expect no error, got: %v", err)
	} else {
		if v.value != "0" {
			t.Errorf("remove: expect value 0, got: %v", v.value)
		}
	}

	// Buffer: [1, 2, 3, 4, 5, 6, 7, 8, 9]
	vNew := *v
	vNew.value = "10"
	if err := mBuf.add(&vNew); err != nil {
		t.Errorf("expect no error, got: %v", err)
	}

	// Buffer: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
	if err := mBuf.add(v); err == nil {
		t.Errorf("expect error, got no error. size: %v, len: %v", mBuf.head, mBuf.tail)
	}

	// Buffer: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
	for i := 0; i < bSize; i++ {
		if v, err := mBuf.remove(); err != nil {
			t.Errorf("remove: expect no error, got: %v", err)
		} else {
			if v.value != strconv.Itoa(i+1) {
				t.Errorf("remove: expect value %d, got: %v", i+1, v.value)
			}
		}
	}

	// Buffer: []
	if _, err := mBuf.remove(); err == nil {
		t.Errorf("remove: expect error, got no error")
	}

	// Buffer: []
	for i := 0; i < bSize; i++ {
		vNew := *v
		vNew.value = strconv.Itoa(i)
		if err := mBuf.add(&vNew); err != nil {
			t.Errorf("expect no error, got: %v", err)
		}
	}

	// Buffer: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
	for i := 0; i < bSize; i++ {
		if v, err := mBuf.remove(); err != nil {
			t.Errorf("remove: expect no error, got: %v", err)
		} else {
			if v.value != strconv.Itoa(i) {
				t.Errorf("remove: expect value %d, got: %v", i, v.value)
			}
		}
	}

	// Buffer: []
	for i := 0; i < bSize/2; i++ {
		if err := mBuf.add(v); err != nil {
			t.Errorf("expect no error, got: %v", err)
		}
	}

	// Buffer: [0-"Test message", 1-"Test message", 2-"Test message", 3-"Test message", 4-"Test message"]
	if _, err := mBuf.remove(); err != nil {
		t.Errorf("remove: expect no error, got: %v", err)
	}

	// Buffer: [1, 2, 3, 4]
	for i := 0; i < bSize/2+1; i++ {
		if err := mBuf.add(v); err != nil {
			t.Errorf("expect no error, got: %v", err)
		}
	}

	// Buffer: [1-"Test message", 2, 3, 4, 5, 6, 7, 8, 9, 10] - all elements value is "Test message"
	for i := 0; i < bSize; i++ {
		if v, err := mBuf.remove(); err != nil {
			t.Errorf("remove: expect no error, got: %v", err)
		} else {
			if v.value != "Test message" {
				t.Errorf("remove: expect value %v, got: %v", "Test message", v.value)
			}
		}
	}
}

func Test_reset(t *testing.T) {
	bSize := 10
	mBuf := newMessageBuffer(bSize)

	if !mBuf.empty() {
		t.Errorf("expect buffer empty got not empty")
	}

	v := &bufferRecord{
		level: 3,
		value: "Test message",
	}
	mBuf.add(v)
	mBuf.add(v)
	mBuf.remove()

	if mBuf.empty() {
		t.Errorf("expect buffer not empty got empty")
	}

	mBuf.reset()
	if !mBuf.empty() {
		t.Errorf("expect buffer empty got not empty")
	}
}

func Test_add(t *testing.T) {
	bSize := 10
	mBuf := newMessageBuffer(bSize)

	v := &bufferRecord{
		level: 3,
		value: "Test message",
	}
	for i := 0; i < bSize; i++ {
		if err := mBuf.add(v); err != nil {
			t.Errorf("expect no error, got: %v", err)
		}
	}

	if err := mBuf.add(v); err == nil {
		t.Errorf("expect error, got no error. size: %v, len: %v", mBuf.head, mBuf.tail)
	}
}

func Test_remove(t *testing.T) {
	bSize := 10
	mBuf := newMessageBuffer(bSize)

	v := &bufferRecord{
		level: 3,
		value: "Test message",
	}
	for i := 0; i < bSize; i++ {
		if err := mBuf.add(v); err != nil {
			t.Errorf("add: expect no error, got: %v", err)
		}
		if i%2 == 0 {
			if _, err := mBuf.remove(); err != nil {
				t.Errorf("remove: expect no error, got: %v", err)
			}
		}
	}
}

func Test_len(t *testing.T) {
	bSize := 10
	mBuf := newMessageBuffer(bSize)

	if mBuf.len() != bSize {
		t.Errorf("len: expect %v, got %v", bSize, mBuf.len())
	}

	v := &bufferRecord{
		level: 3,
		value: "Test message",
	}
	mBuf.add(v)
	mBuf.add(v)
	mBuf.remove()

	if mBuf.len() != bSize {
		t.Errorf("len: expect %v, got %v", bSize, mBuf.len())
	}
}

func Test_size(t *testing.T) {
	bSize := 10
	mBuf := newMessageBuffer(bSize)

	if mBuf.size() != 0 {
		t.Errorf("size: expect %v, got %v", 0, mBuf.size())
	}

	v := &bufferRecord{
		level: 3,
		value: "Test message",
	}
	mBuf.add(v)
	mBuf.add(v)
	mBuf.remove()

	if mBuf.size() != 1 {
		t.Errorf("size: expect %v, got %v", 1, mBuf.size())
	}
}

func Test_empty(t *testing.T) {
	bSize := 10
	mBuf := newMessageBuffer(bSize)

	if !mBuf.empty() {
		t.Errorf("expect buffer empty got not empty")
	}

	v := &bufferRecord{
		level: 3,
		value: "Test message",
	}
	mBuf.add(v)
	mBuf.add(v)
	mBuf.remove()

	if mBuf.empty() {
		t.Errorf("expect buffer not empty got empty")
	}
}
