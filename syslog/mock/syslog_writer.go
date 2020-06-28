package mock

import (
	"log/syslog"
	"sync"
)

type SyslogWriter struct {
	syslog.Writer

	muMessages                                                    sync.RWMutex
	EmergM, AlertM, CritM, ErrM, WarningM, NoticeM, InfoM, DebugM []string
}

// Write sends a log message to the syslog daemon.
func (w *SyslogWriter) Write(b []byte) (int, error) {
	return len(string(b)), nil
}

// Close closes a connection to the syslog daemon.
func (w *SyslogWriter) Close() error {
	return nil
}

// Emerg logs a message with severity LOG_EMERG, ignoring the severity
// passed to New.
func (w *SyslogWriter) Emerg(m string) error {
	w.muMessages.Lock()
	defer w.muMessages.Unlock()

	w.EmergM = append(w.EmergM, m)
	return nil
}

// Alert logs a message with severity LOG_ALERT, ignoring the severity
// passed to New.
func (w *SyslogWriter) Alert(m string) error {
	w.muMessages.Lock()
	defer w.muMessages.Unlock()

	w.AlertM = append(w.AlertM, m)
	return nil
}

// Crit logs a message with severity LOG_CRIT, ignoring the severity
// passed to New.
func (w *SyslogWriter) Crit(m string) error {
	w.muMessages.Lock()
	defer w.muMessages.Unlock()

	w.CritM = append(w.CritM, m)
	return nil
}

// Err logs a message with severity LOG_ERR, ignoring the severity
// passed to New.
func (w *SyslogWriter) Err(m string) error {
	w.muMessages.Lock()
	defer w.muMessages.Unlock()

	w.ErrM = append(w.ErrM, m)
	return nil
}

// Warning logs a message with severity LOG_WARNING, ignoring the
// severity passed to New.
func (w *SyslogWriter) Warning(m string) error {
	w.muMessages.Lock()
	defer w.muMessages.Unlock()

	w.WarningM = append(w.WarningM, m)
	return nil
}

// Notice logs a message with severity LOG_NOTICE, ignoring the
// severity passed to New.
func (w *SyslogWriter) Notice(m string) error {
	w.muMessages.Lock()
	defer w.muMessages.Unlock()

	w.NoticeM = append(w.NoticeM, m)
	return nil
}

// Info logs a message with severity LOG_INFO, ignoring the severity
// passed to New.
func (w *SyslogWriter) Info(m string) error {
	w.muMessages.Lock()
	defer w.muMessages.Unlock()

	w.InfoM = append(w.InfoM, m)
	return nil
}

// Debug logs a message with severity LOG_DEBUG, ignoring the severity
// passed to New.
func (w *SyslogWriter) Debug(m string) error {
	w.muMessages.Lock()
	defer w.muMessages.Unlock()

	w.DebugM = append(w.DebugM, m)
	return nil
}

// Messages - return total message of certain level
func (w *SyslogWriter) Messages(l syslog.Priority) int {
	w.muMessages.RLock()
	defer w.muMessages.RUnlock()

	switch l {
	case syslog.LOG_EMERG:
		return len(w.EmergM)
	case syslog.LOG_ALERT:
		return len(w.AlertM)
	case syslog.LOG_CRIT:
		return len(w.CritM)
	case syslog.LOG_ERR:
		return len(w.ErrM)
	case syslog.LOG_WARNING:
		return len(w.WarningM)
	case syslog.LOG_NOTICE:
		return len(w.NoticeM)
	case syslog.LOG_INFO:
		return len(w.InfoM)
	case syslog.LOG_DEBUG:
		return len(w.DebugM)
	}

	return w.TotalMessages()
}

// EmergM, AlertM, CritM, ErrM, WarningM, NoticeM, InfoM, DebugM
func (w *SyslogWriter) TotalMessages() int {
	w.muMessages.RLock()
	defer w.muMessages.RUnlock()

	return len(w.EmergM) + len(w.AlertM) + len(w.CritM) + len(w.ErrM) + len(w.WarningM) +
		len(w.NoticeM) + len(w.InfoM) + len(w.DebugM)
}

func (w *SyslogWriter) Message(l syslog.Priority, idx int) string {
	if idx < 0 {
		return ""
	}

	w.muMessages.RLock()
	defer w.muMessages.RUnlock()

	var (
		m []string
	)

	switch l {
	case syslog.LOG_EMERG:
		m = w.EmergM
	case syslog.LOG_ALERT:
		m = w.AlertM
	case syslog.LOG_CRIT:
		m = w.CritM
	case syslog.LOG_ERR:
		m = w.ErrM
	case syslog.LOG_WARNING:
		m = w.WarningM
	case syslog.LOG_NOTICE:
		m = w.NoticeM
	case syslog.LOG_INFO:
		m = w.InfoM
	case syslog.LOG_DEBUG:
		m = w.DebugM
	}

	if m != nil && len(m) > idx {
		return m[idx]
	}

	return ""
}
