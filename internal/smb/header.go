package smb

import (
	"encoding/binary"
	"fmt"
)

const headerSize = 64
const FlagServerToRedir uint32 = 0x00000001

var protocolId = [4]byte{0xFE, 'S', 'M', 'B'}

type Header struct {
	ProtocolId    [4]byte
	StructureSize uint16
	CreditCharge  uint16
	Status        uint32
	Command       uint16
	Credits       uint16
	Flags         uint32
	NextCommand   uint32
	MessageID     uint64
	Reserved      uint32
	TreeID        uint32
	SessionID     uint64
	Signature     [16]byte
}

func parseHeader(msg []byte) (*Header, error) {
	if len(msg) < headerSize {
		return nil, fmt.Errorf("[parseHeader] header is too short, length is %d", len(msg))
	}

	if [4]byte(msg[0:4]) != protocolId {
		return nil, fmt.Errorf("[parseHeader] invalid protocol id: % x", msg[0:4])
	}

	if binary.LittleEndian.Uint16(msg[4:6]) != headerSize {
		return nil, fmt.Errorf("[parseHeader] invalid structure size: % x", msg[4:6])
	}

	header := &Header{
		ProtocolId:    [4]byte(msg[0:4]),
		StructureSize: binary.LittleEndian.Uint16(msg[4:6]),
		CreditCharge:  binary.LittleEndian.Uint16(msg[6:8]),
		Status:        binary.LittleEndian.Uint32(msg[8:12]),
		Command:       binary.LittleEndian.Uint16(msg[12:14]),
		Credits:       binary.LittleEndian.Uint16(msg[14:16]),
		Flags:         binary.LittleEndian.Uint32(msg[16:20]),
		NextCommand:   binary.LittleEndian.Uint32(msg[20:24]),
		MessageID:     binary.LittleEndian.Uint64(msg[24:32]),
		Reserved:      binary.LittleEndian.Uint32(msg[32:36]),
		TreeID:        binary.LittleEndian.Uint32(msg[36:40]),
		SessionID:     binary.LittleEndian.Uint64(msg[40:48]),
		Signature:     [16]byte{},
	}
	copy(header.Signature[:], msg[48:64])

	return header, nil
}

func marshalHeader(header Header) []byte {

	buf := make([]byte, headerSize)
	copy(buf[0:4], protocolId[:])
	binary.LittleEndian.PutUint16(buf[4:6], headerSize)
	binary.LittleEndian.PutUint16(buf[6:8], header.CreditCharge)
	binary.LittleEndian.PutUint32(buf[8:12], header.Status)
	binary.LittleEndian.PutUint16(buf[12:14], header.Command)
	binary.LittleEndian.PutUint16(buf[14:16], header.Credits)
	binary.LittleEndian.PutUint32(buf[16:20], header.Flags)
	binary.LittleEndian.PutUint32(buf[20:24], header.NextCommand)
	binary.LittleEndian.PutUint64(buf[24:32], header.MessageID)
	binary.LittleEndian.PutUint32(buf[32:36], header.Reserved)
	binary.LittleEndian.PutUint32(buf[36:40], header.TreeID)
	binary.LittleEndian.PutUint64(buf[40:48], header.SessionID)
	copy(buf[48:64], header.Signature[:])

	return buf

}
