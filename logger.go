package slogger

import (
	"context"
	"io"
	"log"
	"log/syslog"
	"time"

	sl "slogger/syslog"
)

type Logger interface {
	io.Closer

	Alert(ctx context.Context, m string)
	Crit(ctx context.Context, m string)
	Debug(ctx context.Context, m string)
	Emerg(ctx context.Context, m string)
	Err(ctx context.Context, m string)
	Info(ctx context.Context, m string)
	Notice(ctx context.Context, m string)
	Warning(ctx context.Context, m string)
}

type logger struct {
	syslogSender sl.Sender
}

func New(ctx context.Context, syslogProtocol, syslogAddr, syslogTag string,
	bufferSizeMessages int, bufferSendPeriod time.Duration, bufferSendCount int) (Logger, error) {

	// Check syslog connection
	network := syslogProtocol
	if network == sl.SyslogProtocolRELP {
		network = sl.SyslogProtocolTCP
	}
	slog, err := syslog.Dial(network, syslogAddr, syslog.LOG_WARNING|syslog.LOG_DAEMON, syslogTag)
	if err != nil {
		return nil, err
	}
	defer slog.Close()

	// Init logger
	l := new(logger)
	l.syslogSender, err = sl.New(ctx, syslogProtocol, syslogAddr, syslogTag,
		bufferSizeMessages, bufferSendPeriod, bufferSendCount)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (l *logger) Close() error {
	if l.syslogSender == nil {
		return nil
	}
	if err := l.syslogSender.Close(); err != nil {
		return err
	}
	l.syslogSender = nil
	log.Printf("syslog sender gracefully stopped")
	return nil
}

func (l *logger) Alert(ctx context.Context, m string) {
	if err := l.syslogSender.Send(ctx, syslog.LOG_ALERT, m); err != nil {
		log.Printf("%v", err)
	}
}

func (l *logger) Crit(ctx context.Context, m string) {
	if err := l.syslogSender.Send(ctx, syslog.LOG_CRIT, m); err != nil {
		log.Printf("%v", err)
	}
}

func (l *logger) Debug(ctx context.Context, m string) {
	if err := l.syslogSender.Send(ctx, syslog.LOG_DEBUG, m); err != nil {
		log.Printf("%v", err)
	}
}

func (l *logger) Emerg(ctx context.Context, m string) {
	if err := l.syslogSender.Send(ctx, syslog.LOG_EMERG, m); err != nil {
		log.Printf("%v", err)
	}
}

func (l *logger) Err(ctx context.Context, m string) {
	if err := l.syslogSender.Send(ctx, syslog.LOG_ERR, m); err != nil {
		log.Printf("%v", err)
	}
}

func (l *logger) Info(ctx context.Context, m string) {
	if err := l.syslogSender.Send(ctx, syslog.LOG_INFO, m); err != nil {
		log.Printf("%v", err)
	}
}

func (l *logger) Notice(ctx context.Context, m string) {
	if err := l.syslogSender.Send(ctx, syslog.LOG_NOTICE, m); err != nil {
		log.Printf("%v", err)
	}
}

func (l *logger) Warning(ctx context.Context, m string) {
	if err := l.syslogSender.Send(ctx, syslog.LOG_WARNING, m); err != nil {
		log.Printf("%v", err)
	}
}

func (l *logger) Write(ctx context.Context, m string) {
	if err := l.syslogSender.Send(ctx, syslog.LOG_DEBUG, m); err != nil {
		log.Printf("%v", err)
	}
}
