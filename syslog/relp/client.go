package relp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/syslog"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Client - A client to a RELP server
type Client struct {
	priority syslog.Priority
	tag      string
	hostname string
	raddr    string
	timeout  time.Duration

	mu         sync.Mutex
	connection net.Conn

	nextTxn int
}

const (
	relpVersion  = 0
	relpSoftware = "slogger"
	severityMask = 0x07
	facilityMask = 0xf8
)

// Dial like dial in log/syslog
func Dial(raddr string, priority syslog.Priority, tag string, timeout time.Duration) (*Client, error) {
	if priority < 0 || priority > syslog.LOG_LOCAL7|syslog.LOG_DEBUG {
		return nil, errors.New("log/syslog: invalid priority")
	}

	if tag == "" {
		tag = os.Args[0]
	}
	hostname, _ := os.Hostname()

	c := &Client{
		priority: priority,
		tag:      tag,
		hostname: hostname,
		raddr:    raddr,
		timeout:  timeout,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.connect()
	if err != nil {
		return nil, err
	}
	return c, err

}

// connect makes a connection to the rsyslog server.
// It must be called with w.mu held.
func (c *Client) connect() (err error) {
	if c.connection != nil {
		c.connection.Close()
		c.connection = nil
	}

	c.connection, err = net.DialTimeout("tcp", c.raddr, c.timeout)
	if err != nil {
		return err
	}

	if c.hostname == "" {
		c.hostname = c.connection.LocalAddr().String()
	}

	offer := Message{
		Txn:     1,
		Command: CommandOpen,
		Data:    fmt.Sprintf("relp_version=%d\nrelp_software=%s\ncommands=syslog", relpVersion, relpSoftware),
	}
	offer.send(c.connection)

	offerResponse, err := readMessage(c.connection)
	if err != nil {
		return err
	}

	responseParts := strings.Split(offerResponse.Data, "\n")
	if !strings.HasPrefix(responseParts[0], "200 OK") {
		err = fmt.Errorf("server responded to offer with: %s", responseParts[0])
	} else {
		err = nil
	}

	c.nextTxn = 2

	return nil
}

// SendString - Convenience method which constructs a Message and sends it
func (c *Client) sendString(msg string) (err error) {
	message := Message{
		Txn:     c.nextTxn,
		Command: CommandSyslog,
		Data:    msg,
	}

	err = c.sendMessage(message)

	return err
}

// SendMessage - Sends a message using the client's connection
func (c *Client) sendMessage(msg Message) (err error) {
	c.nextTxn = c.nextTxn + 1
	_, err = msg.send(c.connection)

	ack, err := readMessage(c.connection)
	if err != nil {
		return err
	}
	if ack.Command != CommandRsp {
		return fmt.Errorf("response to txn %d was %s: %s", msg.Txn, ack.Command, ack.Data)
	}
	if ack.Txn != msg.Txn {
		return fmt.Errorf("response txn to %d was %d", msg.Txn, ack.Txn)
	}

	return err
}

// SetDeadline - make the next operation timeout if not completed before the given time
func (c *Client) SetDeadline(t time.Time) error {
	return c.connection.SetDeadline(t)
}

// Close - Closes the connection gracefully
func (c *Client) Close() (err error) {
	closeMessage := Message{
		Txn:     c.nextTxn,
		Command: CommandClose,
	}
	_, err = closeMessage.send(c.connection)

	// no need to lock, because of connection field of just created client and no other goroutines have access to it
	if c.connection != nil {
		err := c.connection.Close()
		c.connection = nil
		return err
	}
	return nil
}

func (c Client) Write(b []byte) (int, error) {
	return c.swrite(b, c.tag)
}

func (c Client) Emerg(m string) error {
	return c.semerg(m, c.tag)
}

func (c Client) Alert(m string) error {
	return c.salert(m, c.tag)
}

func (c Client) Crit(m string) error {
	return c.scrit(m, c.tag)
}

func (c Client) Err(m string) error {
	return c.serr(m, c.tag)
}

func (c Client) Warning(m string) error {
	return c.swarning(m, c.tag)
}

func (c Client) Notice(m string) error {
	return c.snotice(m, c.tag)
}

func (c Client) Info(m string) error {
	return c.sinfo(m, c.tag)
}

func (c Client) Debug(m string) error {
	return c.sdebug(m, c.tag)
}

// Internal client funcs
func (c Client) swrite(b []byte, tag string) (int, error) {
	return c.writeAndRetry(c.priority, tag, string(b))
}

func (c Client) semerg(m, tag string) error {
	_, err := c.writeAndRetry(syslog.LOG_EMERG, tag, m)
	return err
}

func (c Client) salert(m, tag string) error {
	_, err := c.writeAndRetry(syslog.LOG_ALERT, tag, m)
	return err
}

func (c Client) scrit(m, tag string) error {
	_, err := c.writeAndRetry(syslog.LOG_CRIT, tag, m)
	return err
}

func (c Client) serr(m, tag string) error {
	_, err := c.writeAndRetry(syslog.LOG_ERR, tag, m)
	return err
}

func (c Client) swarning(m, tag string) error {
	_, err := c.writeAndRetry(syslog.LOG_WARNING, tag, m)
	return err
}

func (c Client) snotice(m, tag string) error {
	_, err := c.writeAndRetry(syslog.LOG_NOTICE, tag, m)
	return err
}

func (c Client) sinfo(m, tag string) error {
	_, err := c.writeAndRetry(syslog.LOG_INFO, tag, m)
	return err
}

func (c Client) sdebug(m, tag string) error {
	_, err := c.writeAndRetry(syslog.LOG_DEBUG, tag, m)
	return err
}

func readMessage(conn io.Reader) (message Message, err error) {
	reader := bufio.NewReader(conn)

	txn, err := reader.ReadString(' ')
	if err != nil {
		return message, err
	}
	message.Txn, _ = strconv.Atoi(strings.TrimSpace(txn))

	cmd, err := reader.ReadString(' ')
	if err != nil {
		return message, err
	}
	message.Command = strings.TrimSpace(cmd)

	// Check for dataLen == 0
	peekLen, err := reader.Peek(1)
	message.Data = ""
	if string(peekLen[:]) != "0" {
		dataLenS, err := reader.ReadString(' ')
		if err != nil {
			return message, err
		}

		dataLen, err := strconv.Atoi(strings.TrimSpace(dataLenS))
		if err != nil {
			return message, err
		}

		dataBytes := make([]byte, dataLen)
		_, err = io.ReadFull(reader, dataBytes)
		if err != nil {
			return message, err
		}
		message.Data = string(dataBytes[:dataLen])
	}
	return message, err
}

func (c *Client) writeAndRetry(p syslog.Priority, tag, s string) (int, error) {
	pr := (c.priority & facilityMask) | (p & severityMask)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connection != nil {
		if n, err := c.write(pr, tag, s); err == nil {
			return n, err
		}
	}
	if err := c.connect(); err != nil {
		return 0, err
	}
	return c.write(pr, tag, s)
}

func (c *Client) write(p syslog.Priority, tag, msg string) (int, error) {
	m := syslogMessage(p, c.hostname, tag, msg)
	if err := c.sendString(m); err != nil {
		return 0, err
	}
	return len(msg), nil
}

func syslogMessage(p syslog.Priority, hostname, tag, msg string) string {
	if strings.HasSuffix(msg, "\n") {
		msg = strings.TrimSuffix(msg, "\n")
	}

	ts := time.Now().Format(time.RFC3339)
	return fmt.Sprintf("<%d>%s %s %s[%d]: %s",
		p, ts, hostname,
		tag, os.Getpid(), msg)
}
