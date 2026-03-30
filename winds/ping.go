package winds

import (
	"encoding/binary"
	"fmt"
	"net"
	"slices"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// Default payload in PING request:
// On macOS ping
// +-----------------------------------+------------+
// | UNIX timestamp second uint32 (BE) |   4 bytes  |
// +-----------------------------------+------------+
// |  UNIX timestamp usec uint32 (BE)  |   4 bytes  |
// +-----------------------------------+------------+
// |      Payload (0x08 ... 0x37)      |  48 bytes  |
// +-----------------------------------+------------+
// On Linux ping
// +-----------------------------------+------------+
// | UNIX timestamp second uint64 (LE) |   8 bytes  |
// +-----------------------------------+------------+
// |  UNIX timestamp usec uint64 (LE)  |   8 bytes  |
// +-----------------------------------+------------+
// |      Payload (0x10 ... 0x37)      |  40 bytes  |
// +-----------------------------------+------------+

var DefaultPingPayloadBase = []byte{
	0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
	0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
	0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27,
	0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
	0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37,
}

type PingResult struct {
	Duration time.Duration
	TTL      int
	Seq      int
	ID       int
	Raw      []byte
	Data     []byte
}

type PingConn struct {
	conn   *icmp.PacketConn
	connV4 *ipv4.PacketConn
	connV6 *ipv6.PacketConn
}

func NewPingConn(network string) (*PingConn, error) {
	if network != NetworkIPv4 && network != NetworkIPv6 {
		m := fmt.Sprintf("invalid network type '%s', pick one of '%s' or '%s'",
			network, NetworkIPv4, NetworkIPv6)
		panic(m)
	}

	listenNetwork := pingerGetNetwork(network)
	conn, err := icmp.ListenPacket(listenNetwork, "")
	if err != nil {
		return nil, err
	}

	pingConn := &PingConn{
		conn:   conn,
		connV4: conn.IPv4PacketConn(),
		connV6: conn.IPv6PacketConn(),
	}

	if pingConn.connV4 != nil {
		if err := pingConn.connV4.SetControlMessage(ipv4.FlagTTL, true); err != nil {
			_ = conn.Close()
			return nil, err
		}
	}

	if pingConn.connV6 != nil {
		if err := pingConn.connV6.SetControlMessage(ipv6.FlagHopLimit, true); err != nil {
			_ = conn.Close()
			return nil, err
		}
	}

	return pingConn, nil
}

func (c *PingConn) Close() error {
	return c.conn.Close()
}

func (c *PingConn) IsIPv6() bool {
	return c.connV6 != nil
}

func (c *PingConn) SetReadDeadline(ddl time.Time) error {
	return c.conn.SetReadDeadline(ddl)
}

func (c *PingConn) ReadFrom(b []byte) (int, int, net.Addr, error) {
	var err error
	var addr net.Addr
	n := 0
	ttl := -1
	if !c.IsIPv6() {
		var cm *ipv4.ControlMessage
		n, cm, addr, err = c.connV4.ReadFrom(b)
		if err != nil {
			return 0, -1, nil, err
		}

		if cm != nil {
			ttl = cm.TTL
		}

	} else {
		var cm *ipv6.ControlMessage
		n, cm, addr, err = c.connV6.ReadFrom(b)
		if err != nil {
			return 0, -1, nil, err
		}

		if cm != nil {
			ttl = cm.HopLimit
		}
	}

	return n, ttl, addr, nil
}

func (c *PingConn) WriteTo(addr net.Addr, b []byte) (int, error) {
	return c.conn.WriteTo(b, addr)
}

func MakePayloadWithTimestampLinux(base []byte, t time.Time) []byte {
	stampSec := t.Unix()
	stampUsec := t.Nanosecond() / 1_000

	payload := make([]byte, len(base)+16)
	binary.LittleEndian.PutUint64(payload, uint64(stampSec))
	binary.LittleEndian.PutUint64(payload[8:], uint64(stampUsec))
	copy(payload[16:], base)
	return payload
}

func MakePayloadWithTimestampMacOS(base []byte, t time.Time) []byte {
	stampSec := t.Unix()
	stampUsec := t.Nanosecond() / 1_000

	payload := make([]byte, len(base)+8)
	binary.BigEndian.PutUint32(payload, uint32(stampSec))
	binary.BigEndian.PutUint32(payload[4:], uint32(stampUsec))
	copy(payload[8:], base)
	return payload
}

type Pinger struct {
	conn    *PingConn
	timeout time.Duration
}

func NewPinger(network string) (*Pinger, error) {
	conn, err := NewPingConn(network)
	if err != nil {
		return nil, err
	}

	sender := &Pinger{
		conn:    conn,
		timeout: 1 * time.Second,
	}
	return sender, nil
}

func (s *Pinger) Close() error {
	return s.conn.Close()
}

func (s *Pinger) Ping(address net.IP, id int, seq int, payload []byte) (*PingResult, error) {
	if (address.To4() == nil) != s.conn.IsIPv6() {
		return nil, net.InvalidAddrError("IP version mismatch")
	}

	echoType := icmp.Type(ipv4.ICMPTypeEcho)
	echoReplyType := icmp.Type(ipv4.ICMPTypeEchoReply)
	if s.conn.IsIPv6() {
		echoType = icmp.Type(ipv6.ICMPTypeEchoRequest)
		echoReplyType = icmp.Type(ipv6.ICMPTypeEchoReply)
	}

	echoRequest := &icmp.Echo{
		ID:  id,
		Seq: seq,
	}

	message := &icmp.Message{
		Type: echoType,
		Code: 0,
		Body: echoRequest,
	}

	recvBuf := make([]byte, 1500)

	timeStart := time.Now()
	echoRequest.Data = MakePayloadWithTimestampLinux(payload, timeStart)

	messageBytes, err := message.Marshal(nil)
	if err != nil {
		return nil, err
	}

	s.conn.IsIPv6()
	remote := pingerMakeAddress(address)
	_, err = s.conn.WriteTo(remote, messageBytes)
	if err != nil {
		return nil, err
	}

	for {
		_ = s.conn.SetReadDeadline(timeStart.Add(s.timeout))
		recvLen, ttl, _, err := s.conn.ReadFrom(recvBuf)
		if err != nil {
			return nil, err
		}

		timeFinished := time.Now()

		protocol := ipv4.ICMPTypeEchoReply.Protocol()
		if s.conn.IsIPv6() {
			protocol = ipv6.ICMPTypeEchoReply.Protocol()
		}

		reply, err := icmp.ParseMessage(protocol, recvBuf[:recvLen])
		if err != nil {
			return nil, err
		}

		if reply.Type != echoReplyType {
			continue
		}

		body, ok := reply.Body.(*icmp.Echo)
		if !ok || body.ID != id || body.Seq != seq {
			continue
		}

		result := &PingResult{
			Duration: timeFinished.Sub(timeStart),
			TTL:      ttl,
			Seq:      body.Seq,
			ID:       body.ID,
			Raw:      slices.Clone(recvBuf[:recvLen]),
			Data:     slices.Clone(body.Data),
		}

		return result, nil
	}
}
