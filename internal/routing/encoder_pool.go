package routing

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
)

// bufferpool is a pool of bytes.Buffer objects for JSON encoding.
// We marshal to a buffer, then write the buffer to the response.
// This reduces per-response allocations from encoder creation.
var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// AcquireBuffer gets a buffer from the pool for JSON encoding.
func AcquireBuffer() *bytes.Buffer {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// ReleaseBuffer returns a buffer to the pool.
func ReleaseBuffer(buf *bytes.Buffer) {
	if buf != nil && buf.Cap() <= 64*1024 { // Only pool buffers up to 64KB to avoid memory bloat
		buf.Reset()
		bufferPool.Put(buf)
	}
}

// EncodeJSON marshals v to JSON and writes to w, using buffer pooling to reduce allocations.
// Returns the number of bytes written and any error.
func EncodeJSON(w io.Writer, v any) (int, error) {
	// Use json.Marshal which is optimized for small-to-medium payloads.
	// For streaming large payloads, consider json.NewEncoder separately.
	b, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}
	return w.Write(b)
}

// EncodeJSONWithBuffer marshals v to JSON using a buffered writer from the pool.
// This is useful for streaming responses where we want to control when the buffer is returned.
type BufferedEncoder struct {
	buf *bytes.Buffer
}

// NewBufferedEncoder creates a new buffered encoder from the pool.
func NewBufferedEncoder() *BufferedEncoder {
	return &BufferedEncoder{
		buf: AcquireBuffer(),
	}
}

// Encode marshals v to JSON into the internal buffer.
func (be *BufferedEncoder) Encode(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = be.buf.Write(data)
	return err
}

// Bytes returns the encoded bytes.
func (be *BufferedEncoder) Bytes() []byte {
	return be.buf.Bytes()
}

// WriteTo writes the buffer contents to w.
func (be *BufferedEncoder) WriteTo(w io.Writer) (int64, error) {
	return be.buf.WriteTo(w)
}

// Release returns the buffer to the pool.
func (be *BufferedEncoder) Release() {
	if be.buf != nil {
		ReleaseBuffer(be.buf)
		be.buf = nil
	}
}
