package cmd

import (
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTripBinarySer(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.bin")
	data := &serializableTestData{A: []byte{0xde, 0xad}, B: 3}
	err := writeSerializedBinary(file, data)
	require.NoError(t, err)

	hasGzip, err := hasGzipHeader(file)
	require.NoError(t, err)
	require.False(t, hasGzip)

	var result serializableTestData
	err = loadSerializedBinary(file, &result)
	require.NoError(t, err)
	require.EqualValues(t, data, &result)
}

func TestRoundTripBinarySerWithGzip(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.bin.gz")
	data := &serializableTestData{A: []byte{0xde, 0xad}, B: 3}
	err := writeSerializedBinary(file, data)
	require.NoError(t, err)

	hasGzip, err := hasGzipHeader(file)
	require.NoError(t, err)
	require.True(t, hasGzip)

	var result serializableTestData
	err = loadSerializedBinary(file, &result)
	require.NoError(t, err)
	require.EqualValues(t, data, &result)
}

func hasGzipHeader(filename string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer file.Close()

	header := make([]byte, 2)
	_, err = file.Read(header)
	if err != nil {
		return false, err
	}

	// Gzip header magic numbers: 1F 8B
	return header[0] == 0x1F && header[1] == 0x8B, nil
}

type serializableTestData struct {
	A []byte
	B uint8
}

func (s *serializableTestData) Serialize(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, uint64(len(s.A))); err != nil {
		return err
	}
	if _, err := w.Write(s.A); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, s.B); err != nil {
		return err
	}
	return nil
}

func (s *serializableTestData) Deserialize(in io.Reader) error {
	var lenA uint64
	if err := binary.Read(in, binary.BigEndian, &lenA); err != nil {
		return err
	}
	s.A = make([]byte, lenA)
	if _, err := io.ReadFull(in, s.A); err != nil {
		return err
	}
	if err := binary.Read(in, binary.BigEndian, &s.B); err != nil {
		return err
	}
	return nil
}
