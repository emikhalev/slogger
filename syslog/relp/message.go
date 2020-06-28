package relp

import (
	"fmt"
	"io"
	"net"
)

type Command string

const (
	CommandOpen   = "open"
	CommandSyslog = "syslog"
	CommandClose  = "close"
	CommandRsp    = "rsp"
	CommandAbort  = "abort"
)

// Message - A single RELP message
type Message struct {
	// The transaction ID that the message was sent in
	Txn int
	// The command that was run. Will be "syslog" pretty much always under normal
	// operation
	Command string
	// The actual message data
	Data string

	// true if the message has been acked
	Acked bool

	// Used internally for acking.
	sourceConnection net.Conn
}

// Send - Sends a message
func (m Message) send(out io.Writer) (nn int, err error) {
	outLength := len([]byte(m.Data))
	outString := fmt.Sprintf("%d %s %d %s\n", m.Txn, m.Command, outLength, m.Data)
	return out.Write([]byte(outString))
}

// Ack - Acknowledges a message
func (m *Message) Ack() (err error) {
	if m.Acked {
		return fmt.Errorf("called Ack on already-acknowledged message %d", m.Txn)
	}

	if m.sourceConnection == nil {
		// If the source connection is gone, we don't need to do any work.
		return nil
	}

	ackMessage := Message{
		Txn:     m.Txn,
		Command: CommandRsp,
		Data:    "200 OK",
	}
	_, err = ackMessage.send(m.sourceConnection)
	if err != nil {
		return err
	}
	m.Acked = true
	return
}
