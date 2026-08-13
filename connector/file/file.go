package file

import (
	"os"

	dge "github.com/wiryax/dage/core"
)

type FileConnector struct {
	Name string
	Flag int
	Mode os.FileMode
	conn *os.File
}

func (f *FileConnector) Acquire(_ any) any {
	return f.conn
}
func (f *FileConnector) Validate(_ *dge.GraphContext) error {
	file, err := os.OpenFile(f.Name, 688, f.Mode)
	f.conn = file
	return err
}
func (f *FileConnector) Release() error {
	return f.conn.Close()
}
