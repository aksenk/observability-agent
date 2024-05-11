package core

import (
	"bytes"
	"compress/gzip"
	"io"
)

// isGzipped
// Проверка данных на сжатие форматом gzip.
func isGzipped(data []byte) bool {
	// Проверяем, является ли первый байт 0x1f и второй байт 0x8b, что характерно для gzip формата.
	return len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
}

// unGzip
// Распаковка данных из gzip формата.
func unGzip(inputData []byte) ([]byte, error) {
	gzReader, err := gzip.NewReader(bytes.NewReader(inputData))
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	outputData, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, err
	}

	return outputData, nil
}
