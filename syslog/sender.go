package syslog

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	slog "log/syslog"
	"sync"
	"time"

	slRelp "slogger/syslog/relp"
)

const (
	SyslogProtocolTCP  = "tcp"
	SyslogProtocolUDP  = "udp"
	SyslogProtocolRELP = "relp"
)

type Sender interface {
	io.Closer
	Send(ctx context.Context, level slog.Priority, v string) error
}

type SyslogWriter interface {
	io.Closer
	Write([]byte) (int, error)
	Emerg(string) error
	Alert(string) error
	Crit(string) error
	Err(string) error
	Warning(string) error
	Notice(string) error
	Info(string) error
	Debug(string) error
}

type dialMethodFunc func(context.Context, string, string, string) (slog SyslogWriter, ok bool)

type syslog struct {
	syslogProtocol, syslogAddr, syslogTag string
	dialMethod                            dialMethodFunc

	bufferSendPeriod time.Duration
	bufferSendCount  int
	syslogBuffer     *messageBuffer
	cancelFunc       context.CancelFunc
	wgSyslogSend     sync.WaitGroup
}

func New(ctx context.Context, syslogProtocol, syslogAddr, syslogTag string,
	bufferSizeMessages int, bufferSendPeriod time.Duration, bufferSendCount int) (Sender, error) {
	// Init sender
	sender := &syslog{
		syslogProtocol: syslogProtocol,
		syslogAddr:     syslogAddr,
		syslogTag:      syslogTag,
		syslogBuffer:   newMessageBuffer(bufferSizeMessages),
		dialMethod:     syslogDial,
	}

	// Start sender goroutine
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	sender.cancelFunc = cancelFunc
	sender.wgSyslogSend.Add(1)
	go sender.syslogSend(cancelCtx, bufferSendPeriod, bufferSendCount)

	return sender, nil
}

func (s *syslog) Close() error {
	if s.cancelFunc == nil {
		return nil
	}
	s.cancelFunc()

	c := make(chan struct{})
	go func() {
		s.wgSyslogSend.Wait()
		c <- struct{}{}
	}()
	select {
	case <-c:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("cannot gracefully stop syslog sender: timeout")
	}
}

// Send - adds message (v inteface{}) to send buffer with level
func (s *syslog) Send(ctx context.Context, level slog.Priority, v string) error {
	if s.syslogBuffer == nil {
		return fmt.Errorf("message buffer not inited")
	}
	if err := s.syslogBuffer.add(&bufferRecord{
		ctx:   ctx,
		ts:    time.Now().UTC().Format(time.RFC3339Nano),
		level: level,
		value: v,
	}); err != nil {
		return fmt.Errorf("cannot add message to syslog buffer: %v", err)
	}

	return nil
}

func (s *syslog) SetDialMethod(dialFunc dialMethodFunc) {
	s.dialMethod = dialFunc
}

func (s *syslog) syslogSend(ctx context.Context, bufferSendPeriod time.Duration, maxRecsToSend int) {
	defer s.wgSyslogSend.Done()

	recs := make([]*bufferRecord, maxRecsToSend, maxRecsToSend)
	tickCh := time.Tick(bufferSendPeriod)

loop:
	for {
		select {
		case <-tickCh:
			i := 0
			for !s.syslogBuffer.empty() && i < maxRecsToSend {
				r, err := s.syslogBuffer.remove()
				if err != nil {
					log.Printf("cannot move remove from syslog buffer: %v", err)
					continue
				}
				recs[i] = r
				i++
			}
			s.toSyslogBulk(ctx, recs[0:i])

		case <-ctx.Done():
			recs := recs[0:0]
			for !s.syslogBuffer.empty() {
				r, err := s.syslogBuffer.remove()
				if err != nil {
					log.Printf("cannot move remove from syslog buffer: %v", err)
					continue
				}
				recs = append(recs, r)
			}
			s.toSyslogBulk(ctx, recs)
			break loop
		}
	}
}

func (s *syslog) toSyslogBulk(ctx context.Context, records []*bufferRecord) {
	if s.syslogProtocol == "" || s.syslogAddr == "" || s.syslogTag == "" || len(records) == 0 {
		return
	}
	if s.dialMethod == nil {
		return
	}
	var (
		slog SyslogWriter
		ok   bool
	)
	for {
		slog, ok = s.dialMethod(ctx, s.syslogProtocol, s.syslogAddr, s.syslogTag)
		if ok {
			break
		}
		time.Sleep(1 * time.Second)
	}
	defer slog.Close()

	for _, r := range records {
		s.toSyslog(r.ctx, slog, r.level, r.value)
	}
}

func (s *syslog) toSyslog(ctx context.Context, sl SyslogWriter, lvl slog.Priority, st string) {
	var (
		err error
	)

	switch lvl {
	case slog.LOG_EMERG:
		err = sl.Emerg(st)
	case slog.LOG_ALERT:
		err = sl.Alert(st)
	case slog.LOG_CRIT:
		err = sl.Crit(st)
	case slog.LOG_ERR:
		err = sl.Err(st)
	case slog.LOG_WARNING:
		err = sl.Warning(st)
	case slog.LOG_NOTICE:
		err = sl.Notice(st)
	case slog.LOG_INFO:
		err = sl.Info(st)
	case slog.LOG_DEBUG:
		err = sl.Debug(st)
	}

	if err != nil {
		log.Printf("cannot send to syslog: %v", err)
	}
}

func syslogDial(ctx context.Context, syslogProtocol, syslogAddr, syslogTag string) (slw SyslogWriter, ok bool) {
	if syslogProtocol == "" || syslogAddr == "" || syslogTag == "" {
		return nil, false
	}
	var (
		err error
	)

	if syslogProtocol == SyslogProtocolRELP {
		slw, err = slRelp.Dial(syslogAddr, slog.LOG_WARNING|slog.LOG_DAEMON, syslogTag, 5*time.Second)
	} else {
		slw, err = slog.Dial(syslogProtocol, syslogAddr, slog.LOG_WARNING|slog.LOG_DAEMON, syslogTag)
	}

	if err != nil {
		log.Printf("cannot open dial to syslog")
		return nil, false
	}
	return slw, true
}
