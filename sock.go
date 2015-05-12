package packet

import (
	"net"
	"syscall"
	"time"
)

const (
	Heartbeat  = 10 * time.Second
	packetSize = 8192
	bufferSize = 131072
)

type Conn struct {
	r      *reader
	conn   net.Conn
	raddr  net.Addr
	closed chan struct{}
}

func newConn(r *reader, conn net.Conn, raddr net.Addr) *Conn {
	c := &Conn{
		r:      r,
		conn:   conn,
		raddr:  raddr,
		closed: make(chan struct{}),
	}

	go func() {
		// send heartbeat
		ticker := time.NewTicker(Heartbeat)
		for {
			select {
			case <-c.closed:
				return
			case <-ticker.C:
				c.write(nil)
			}
		}
	}()
	return c
}

func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	if c.raddr == nil {
		return c.RemoteAddr()
	}
	return c.raddr
}

func (c *Conn) Read(b []byte) (int, error) {
	select {
	case <-c.closed:
		return 0, syscall.EINVAL
	default:
		var saddr string
		if c.raddr != nil {
			saddr = c.raddr.String()
		}
		return c.r.read(b, saddr)
	}
}

func (c *Conn) write(b []byte) (int, error) {
	if c.raddr == nil {
		return c.conn.Write(b)
	}
	return c.conn.(net.PacketConn).WriteTo(b, c.raddr)
}

func (c *Conn) Write(b []byte) (int, error) {
	select {
	case <-c.closed:
		return 0, syscall.EINVAL
	default:
		return c.write(b)
	}
}

func (c *Conn) Close() error {
	close(c.closed)
	if c.raddr == nil {
		return c.conn.Close()
	}
	return nil
}

// stub
func (c *Conn) SetDeadline(time.Time) error      { return nil }
func (c *Conn) SetReadDeadline(time.Time) error  { return nil }
func (c *Conn) SetWriteDeadline(time.Time) error { return nil }
