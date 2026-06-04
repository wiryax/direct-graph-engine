package dge

import "fmt"

type Connector struct {
	connectors map[string]any
}

func NewConnector(connectors map[string]any) *Connector {
	return &Connector{
		connectors: connectors,
	}
}

func (c *Connector) GetConnector(key string) (any, error) {
	con, ok := c.connectors[key]
	if !ok {
		return nil, fmt.Errorf("couldn't find any connector with key %s", key)
	}
	return con, nil
}

func (c *Connector) SetConnector(key string, conn any) {
	c.connectors[key] = conn
}
