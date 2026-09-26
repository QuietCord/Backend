package util

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"io"
)

// InflateSettings decompresses blobs from the Quiet/Vencord client (fflate deflateSync → zlib wrapper).
func InflateSettings(data []byte) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewReader(data))
	if err == nil {
		defer zr.Close()
		return io.ReadAll(zr)
	}

	r := flate.NewReader(bytes.NewReader(data))
	defer r.Close()
	return io.ReadAll(r)
}

// DeflateSettings compresses with zlib for octet-stream settings uploads.
func DeflateSettings(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
