SLogger

Package to send messages to syslog with buffer and RELP protocol support.<br/>
RELP provide reliable delivery of event messages (https://en.wikipedia.org/wiki/Reliable_Event_Logging_Protocol) 

More information about RELP:<br/>
https://rainer.gerhards.net/2008/03/relp-reliable-event-logging-protocol.html
https://rainer.gerhards.net/2008/04/on-unreliability-of-plain-tcp-syslog.html

Usage example:

	ctx := context.Background()
	l, err := logger.New(ctx, syslog.SyslogProtocolRELP, testSyslogAddrRELP, testSyslogTag, 32, 1100*time.Millisecond, 128)
	if err != nil {
		t.Errorf("cannot init log: %v", err)
	}
	defer l.Close
	l.Err(ctx, mes)