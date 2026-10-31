package signals

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"time"
)

func Encode(data any) (string, error) {
	var jbuf bytes.Buffer
	enc := json.NewEncoder(&jbuf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(data); err != nil {
		return "", err
	}

	j := bytes.TrimRight(jbuf.Bytes(), "\n")

	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.DefaultCompression)
	if err != nil {
		return "", err
	}
	gz.OS = 3
	gz.ModTime = time.Now()
	if _, err := gz.Write(j); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(buf.Bytes()), nil
}
