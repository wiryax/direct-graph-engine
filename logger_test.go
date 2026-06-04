package dge

// import (
// "testing"
// "os"
// )

// func TestLoggerInStdout(t *testing.T) {
// log := NewLogger(os.Stdout)
// log.FlushLog(EventStart, LevelInfo, "", "A")
// }

// func TestLoggerInFile(t *testing.T) {
// fd, err := os.OpenFile("log.txt", os.O_WRONLY, 0644)
// if err != nil {
// t.Fatalf("failed open file %v", err)
// }
// defer fd.Close()

// log := NewLogger(fd)
// log.FlushLog(EventStart, LevelInfo, "", "A")
// log.FlushLog(EventSuccess, LevelInfo, "", "A")
// }
