package file

import (
	"os"

	dge "github.com/wiryax/dage/core"
)

type FileConnector struct {
	Name string
	perm os.FileMode
	flag int
	conn *os.File
}

func NewFileConnector(name string, flag int, perm os.FileMode) *FileConnector {
	return &FileConnector{
		Name: name,
		perm: perm,
		flag: flag,
	}
}

func (f *FileConnector) Acquire(_ any) any {
	return f.conn
}
func (f *FileConnector) Validate(_ *dge.GraphContext) error {
	file, err := os.OpenFile(f.Name, f.flag, f.perm)
	f.conn = file
	return err
}
func (f *FileConnector) Release() error {
	return f.conn.Close()
}
