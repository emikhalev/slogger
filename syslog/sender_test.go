package syslog

import (
	"context"
	"fmt"
	slog "log/syslog"
	"strconv"
	"testing"
	"time"

	"slogger/syslog/mock"
)

type logMessage struct {
	Level        slog.Priority `json:"level"`
	Msg          string        `json:"msg"`
	ParentSpanID string        `json:"parent_span_id"`
	TraceID      string        `json:"trace_id"`
	SpanID       string        `json:"span_id"`
	TS           string        `json:"ts"`
}

func TestSyslog_Close(t *testing.T) {
	ctx := context.Background()
	mockWriter := &mock.SyslogWriter{}
	bufSize := 1024 * 1024

	s, err := New(ctx, "1", "2", "3", bufSize, 100*time.Second, bufSize/2)
	if err != nil {
		t.Errorf("cannot create syslog sender: %v", err)
	}
	sl := s.(*syslog)
	sl.SetDialMethod(func(context.Context, string, string, string) (SyslogWriter, bool) {
		return mockWriter, true
	})

	for i := 0; i < bufSize; i++ {
		s.Send(ctx, slog.LOG_ERR, "")
	}

	totalMessages := mockWriter.TotalMessages()
	if totalMessages > 0 {
		t.Errorf("expect mockWriter messages = 0, got: %d", totalMessages)
	}
	s.Close()

	totalMessages = mockWriter.TotalMessages()
	if totalMessages > bufSize {
		t.Errorf("expect mockWriter messages = %d, got: %d", bufSize, totalMessages)
	}
}

func TestSyslog_Send(t *testing.T) {
	ctx := context.Background()
	mockWriter := &mock.SyslogWriter{}

	bufSize := 32 // should be even number for this test

	s, err := New(ctx, "1", "2", "3", bufSize, 10*time.Millisecond, bufSize/2)
	if err != nil {
		t.Errorf("cannot create syslog sender: %v", err)
	}
	sl := s.(*syslog)
	sl.SetDialMethod(func(context.Context, string, string, string) (SyslogWriter, bool) {
		return mockWriter, true
	})

	for i := 0; i < bufSize+2; i++ {
		s.Send(ctx, slog.LOG_ERR, strconv.Itoa(i))
	}

	// Check 0 messages in mock
	totalMessages := mockWriter.TotalMessages()
	if totalMessages > 0 {
		t.Errorf("expect mockWriter messages = 0, got: %d", totalMessages)
	}

	time.Sleep(15 * time.Millisecond)
	// Check 16 messages in mock
	totalMessages = mockWriter.TotalMessages()
	if totalMessages != bufSize/2 {
		t.Errorf("expect mockWriter messages = %d, got: %d", bufSize/2, totalMessages)
	}

	time.Sleep(15 * time.Millisecond)
	// Check 32 messages in mock
	totalMessages = mockWriter.TotalMessages()
	if totalMessages != bufSize {
		t.Errorf("expect mockWriter messages = %d, got: %d", bufSize, totalMessages)
	}

}

func TestSyslog_toSyslog(t *testing.T) {
	ctx := context.Background()
	mockWriter := &mock.SyslogWriter{}
	s := &syslog{}
	f := "Error on retrieving status by period from %s to %s"
	v := []interface{}{"01-02-1900", "02-02-1900"}
	m := fmt.Sprintf(f, v...)

	s.toSyslog(ctx, mockWriter, slog.LOG_ERR, m)
	s.toSyslog(ctx, mockWriter, slog.LOG_WARNING, m)

	if mockWriter.TotalMessages() != 2 {
		t.Errorf("expect mockWriter messages = %d, got: %d", 2, mockWriter.TotalMessages())
		return
	}

	cnt := mockWriter.Messages(slog.LOG_WARNING)
	if cnt != 1 {
		t.Errorf("expect mockWriter Warning messages count = %d, got: %d", 1, cnt)
		return
	}

	cnt = mockWriter.Messages(slog.LOG_ERR)
	if cnt != 1 {
		t.Errorf("expect mockWriter Error messages count = %d, got: %d", 1, cnt)
		return
	}
}
