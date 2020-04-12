package websocket

import (
	"net"
	"time"
)

// Batch represents a batch of websocket messages that, when written with
// Conn.WriteBatch, are pipelined into a single network packet.
type Batch struct {
	bconn  batchConn
	nmsgs  int
	wsconn *Conn
}

// Reset resets the batch.
func (b *Batch) Reset() {
	b.nmsgs = 0
	b.bconn.buf = b.bconn.buf[:0]
}

// Len returns number of messages in the batch.
func (b *Batch) Len() int {
	return b.nmsgs
}

// WriteMessage writes a message to the batch. It is safe to modify the
// contents of the argument after WriteMessage returns but not before.
func (b *Batch) WriteMessage(messageType int, data []byte) error {
	if b.wsconn == nil {
		b.wsconn = newConn(&b.bconn, true, 0, 4096, nil, nil, nil)
	}
	if err := b.wsconn.WriteMessage(messageType, data); err != nil {
		return err
	}
	b.nmsgs++
	return nil
}

// WriteBatch write the given batch to Conn. The batch messages will be
// applied sequentially. The batch is reset after a successful write.
func (c *Conn) WriteBatch(b *Batch) error {
	if b.nmsgs == 0 {
		return nil
	}
	if b.bconn.dldirty {
		if err := c.conn.SetWriteDeadline(b.bconn.dl); err != nil {
			return err
		}
		b.bconn.dldirty = false
	}
	_, err := c.conn.Write(b.bconn.buf)
	return err
}

type batchConn struct {
	dldirty bool
	dl      time.Time
	buf     []byte
}

var _ net.Conn = &batchConn{}

func (c *batchConn) Close() error {
	panic("not supported")
}
func (c *batchConn) LocalAddr() net.Addr {
	panic("not supported")
}
func (c *batchConn) RemoteAddr() net.Addr {
	panic("not supported")
}
func (c *batchConn) Read(p []byte) (int, error) {
	panic("not supported")
}
func (c *batchConn) SetReadDeadline(deadline time.Time) error {
	panic("not supported")
}
func (c *batchConn) SetDeadline(deadline time.Time) error {
	panic("not supported")
}
func (c *batchConn) SetWriteDeadline(deadline time.Time) error {
	c.dldirty = true
	c.dl = deadline
	return nil
}
func (c *batchConn) Write(p []byte) (int, error) {
	c.buf = append(c.buf, p...)
	return len(p), nil
}
