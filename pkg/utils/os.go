package utils

import (
	"bytes"
	"errors"
	"io"
	"os"
)

// ReadFileMax 读取文件（限大小）
func ReadFileMax(p string, max int) (string, bool, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", false, err
	}
	defer f.Close()
	var buf bytes.Buffer
	n, err := io.CopyN(&buf, f, int64(max)+1)
	trunc := n > int64(max)
	if trunc {
		b := buf.Bytes()[:max]
		return string(b) + "\n[...truncated...]", true, nil
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return "", false, err
	}
	return buf.String(), false, nil
}
