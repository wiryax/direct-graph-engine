package connector

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"os"

	dge "github.com/wiryax/dage/core"
)

type CSVReaderConnector struct {
	fileName string
	conn     *csv.Reader
	fd       *os.File
}

func NewCSVReaderConnector(fileName string) *CSVReaderConnector {
	return &CSVReaderConnector{
		fileName: fileName,
	}
}

func (c *CSVReaderConnector) Validate(gCtx *dge.GraphContext) error {
	fd, err := os.OpenFile(c.fileName, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	buffReader := bufio.NewReaderSize(fd, 64*1024)

	c.conn = csv.NewReader(buffReader)
	c.conn.ReuseRecord = true
	return nil
}
func (c *CSVReaderConnector) Acquire(conn any) any {
	return c.conn
}
func (c *CSVReaderConnector) Release() error {
	return c.fd.Close()
}

type CSVWriterConnector struct {
	data   *bytes.Buffer
	column []string
	conn   *csv.Writer
}

func NewCSVWriterConnector(column []string) *CSVWriterConnector {
	return &CSVWriterConnector{
		column: column,
	}
}

func (c *CSVWriterConnector) GetData() []byte {
	return c.data.Bytes()
}

func (c *CSVWriterConnector) Validate(gCtx *dge.GraphContext) error {
	data := make([]byte, 0, 128)
	c.data = bytes.NewBuffer(data)
	c.conn = csv.NewWriter(c.data)
	return nil
}
func (c *CSVWriterConnector) Acquire(_ any) any {
	return c.conn
}
func (c *CSVWriterConnector) Release() error {
	c.conn.Flush()
	return nil
}
