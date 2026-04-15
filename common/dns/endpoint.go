package dns

import (
	"fmt"
	"net"
)

type Endpoint interface {
	Address() string
	FullAddress() string
	Dial() (net.Conn, error)
}

type UDPEndpoint struct {
	IP   net.IP
	Port int
}

func NewUDPEndpoint(ip net.IP, port int) Endpoint {
	e := &UDPEndpoint{
		IP:   ip,
		Port: port,
	}

	return e
}

func ToUDPEndpoints(ips []net.IP, port int) []Endpoint {
	endpoints := make([]Endpoint, len(ips))
	for i, ip := range ips {
		endpoints[i] = NewUDPEndpoint(ip, port)
	}

	return endpoints
}

func (e *UDPEndpoint) Address() string {
	return e.IP.String()
}

func (e *UDPEndpoint) FullAddress() string {
	base := net.JoinHostPort(e.IP.String(), fmt.Sprintf("%d", e.Port))
	return base + "/udp"
}

func (e *UDPEndpoint) String() string {
	return e.FullAddress()
}

func (e *UDPEndpoint) Dial() (net.Conn, error) {
	raddr := &net.UDPAddr{
		IP:   e.IP,
		Port: e.Port,
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

type TCPEndpoint struct {
	IP   net.IP
	Port int
}

func NewTCPEndpoint(ip net.IP, port int) Endpoint {
	e := &TCPEndpoint{
		IP:   ip,
		Port: port,
	}

	return e
}

func ToTCPEndpoints(ips []net.IP, port int) []Endpoint {
	endpoints := make([]Endpoint, len(ips))
	for i, ip := range ips {
		endpoints[i] = NewTCPEndpoint(ip, port)
	}

	return endpoints
}

func (e *TCPEndpoint) Address() string {
	return e.IP.String()
}

func (e *TCPEndpoint) FullAddress() string {
	base := net.JoinHostPort(e.IP.String(), fmt.Sprintf("%d", e.Port))
	return base + "/tcp"
}

func (e *TCPEndpoint) String() string {
	return e.FullAddress()
}

func (e *TCPEndpoint) Dial() (net.Conn, error) {
	raddr := &net.TCPAddr{
		IP:   e.IP,
		Port: e.Port,
	}

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

type UNIXEndpoint struct {
	Path string
}

func NewUNIXEndpoint(path string) Endpoint {
	e := &UNIXEndpoint{
		Path: path,
	}

	return e
}

func ToUNIXEndpoints(paths []string) []Endpoint {
	endpoints := make([]Endpoint, len(paths))
	for i, path := range paths {
		endpoints[i] = NewUNIXEndpoint(path)
	}

	return endpoints
}

func (e *UNIXEndpoint) Address() string {
	return e.Path
}

func (e *UNIXEndpoint) FullAddress() string {
	return e.Path + "/unix"
}

func (e *UNIXEndpoint) String() string {
	return e.FullAddress()
}

func (e *UNIXEndpoint) Dial() (net.Conn, error) {
	raddr := &net.UnixAddr{
		Name: e.Path,
		Net:  "unix",
	}

	conn, err := net.DialUnix("unix", nil, raddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
