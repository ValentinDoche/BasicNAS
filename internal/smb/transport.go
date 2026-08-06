package smb

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const transportPrefixSize = 4

func readStreamProtocolLength(r io.Reader) (uint32, error) {
	var transport [4]byte

	if _, err := io.ReadFull(r, transport[:]); err != nil {
		return 0, err
	}
	if transport[0] != 0 {
		return 0, fmt.Errorf("invalid zero byte: %#x", transport[0])
	}

	length := binary.BigEndian.Uint32(transport[:])
	if length == 0 {
		return 0, errors.New("empty SMB2 message")
	}

	return length, nil
}

func readSMB2Message(r io.Reader) ([]byte, error) {

	streamProtocolLength, err := readStreamProtocolLength(r)
	if err != nil {
		return nil, err
	}

	smb2Message := make([]byte, streamProtocolLength)
	if _, err = io.ReadFull(r, smb2Message); err != nil {
		return nil, err
	}

	return smb2Message, nil
}

func writeSMB2Message(w io.Writer, msg []byte) error {
	if len(msg) > 0xFFFFFF {
		return fmt.Errorf("[writeSMB2Message] message too large: %d bytes", len(msg))
	}
	packet := make([]byte, 4+len(msg))
	binary.BigEndian.PutUint32(packet[0:4], uint32(len(msg)))
	copy(packet[4:], msg)
	_, err := w.Write(packet)
	return err
}
