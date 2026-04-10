package dns

import (
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

type Client struct {
	Endpoint Endpoint
	conn     net.Conn
}

func NewClient(endpoint Endpoint) *Client {
	c := &Client{
		Endpoint: endpoint,
	}

	return c
}

func (c *Client) Dialed() bool {
	return c.conn != nil
}

func (c *Client) Dial() error {
	conn, err := c.Endpoint.Dial()
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	return err
}

func (c *Client) Query(t Type, name string) (*dnsmessage.Message, error) {
	if c.conn == nil {
		return nil, ErrNotDialed
	}

	query := dnsmessage.Message{
		Questions: []dnsmessage.Question{
			{
				Name:  dnsmessage.MustNewName(name),
				Type:  t,
				Class: dnsmessage.ClassINET,
			},
		},
	}

	queryBytes, err := query.Pack()
	if err != nil {
		return nil, err
	}

	_, err = c.conn.Write(queryBytes)
	if err != nil {
		return nil, err
	}

	respBuf := make([]byte, 2000)
	n, err := c.conn.Read(respBuf)
	if err != nil {
		return nil, err
	}

	var resp dnsmessage.Message
	err = resp.Unpack(respBuf[:n])
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
