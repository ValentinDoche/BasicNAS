package smb

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"slices"
	"time"
)

var supportedDialects = []uint16{Dialect210}

const negotiateRequestSize = 36
const negotiateResponseSize = 64
const (
	Dialect202      uint16 = 0x0202
	Dialect210      uint16 = 0x0210
	Dialect300      uint16 = 0x0300
	Dialect302      uint16 = 0x0302
	Dialect311      uint16 = 0x0311
	DialectWildcard uint16 = 0x02FF // utilisé par le SMB1 multi-protocol negotiate
)
const (
	NegotiateSigningEnabled  uint16 = 0x0001
	NegotiateSigningRequired uint16 = 0x0002
)

type NegotiateRequest struct {
	DialectCount           uint16
	SecurityMode           uint16
	Reserved               uint16
	Capabilities           uint32
	ClientGUID             [16]byte
	ClientStartTime        uint64
	NegotiateContextOffset uint32
	NegotiateContextCount  uint16
	Reserved2              uint16
	Dialects               []uint16
}

type NegotiateResponse struct {
	SecurityMode           uint16
	DialectRevision        uint16
	NegotiateContextCount  uint16
	ServerGUID             [16]byte
	Capabilities           uint32
	MaxTransactSize        uint32
	MaxReadSize            uint32
	MaxWriteSize           uint32
	SystemTime             uint64
	ServerStartTime        uint64
	SecurityBufferOffset   uint16
	SecurityBufferLength   uint16
	NegotiateContextOffset uint32
	SecurityBuffer         []byte
}

func (s *Server) handleNegotiate(conn net.Conn, header *Header, msg []byte) error {
	req, err := parseNegotiateRequest(msg)
	if err != nil {
		return fmt.Errorf("[handleNegotiate] failed to parse negotiate : %v", err)
	}
	dialect, err := selectDialect(req.Dialects)
	if err != nil {
		return fmt.Errorf("[handleNegotiate] failed to select dialect : %v", err)
	}
	respHeader := Header{
		Command:   CommandNegotiate,
		Credits:   1,
		Flags:     FlagServerToRedir,
		MessageID: header.MessageID,
	}
	respBody := NegotiateResponse{
		SecurityMode:         NegotiateSigningEnabled,
		DialectRevision:      dialect,
		ServerGUID:           s.guid,
		MaxTransactSize:      65536,
		MaxReadSize:          65536,
		MaxWriteSize:         65536,
		SystemTime:           timeToFiletime(time.Now()),
		SecurityBufferOffset: headerSize + negotiateResponseSize,
	}
	out := append(marshalHeader(respHeader), marshalNegotiateResponse(respBody)...)
	if err := writeSMB2Message(conn, out); err != nil {
		return fmt.Errorf("negotiate: write: %w", err)
	}
	log.Printf("negotiated dialect %#x with %s", dialect, conn.RemoteAddr())
	return nil

}

func parseNegotiateRequest(msg []byte) (*NegotiateRequest, error) {

	if len(msg) < headerSize {
		return nil, fmt.Errorf("[parseNegotiateRequest] message too short: %d bytes", len(msg))
	}
	body := msg[headerSize:]

	if len(body) < negotiateRequestSize {
		return nil, fmt.Errorf("[parseNegotiateRequest] body too short, length is %d", len(body))
	}

	if binary.LittleEndian.Uint16(body[0:2]) != negotiateRequestSize {
		return nil, fmt.Errorf("[parseNegotiateRequest] invalid structure size: % x", body[0:2])
	}

	negotiateRequest := &NegotiateRequest{
		DialectCount: binary.LittleEndian.Uint16(body[2:4]),
		SecurityMode: binary.LittleEndian.Uint16(body[4:6]),
		Reserved:     binary.LittleEndian.Uint16(body[6:8]),
		Capabilities: binary.LittleEndian.Uint32(body[8:12]),
	}
	copy(negotiateRequest.ClientGUID[:], body[12:28])

	if len(body) < negotiateRequestSize+2*int(negotiateRequest.DialectCount) {
		return nil, fmt.Errorf("[parseNegotiateRequest] truncated dialect list: %d dialects announced, %d bytes available", negotiateRequest.DialectCount, len(body))
	}
	negotiateRequest.Dialects = make([]uint16, negotiateRequest.DialectCount)
	supports311 := false
	for i := range negotiateRequest.Dialects {
		offset := negotiateRequestSize + i*2
		negotiateRequest.Dialects[i] = binary.LittleEndian.Uint16(body[offset : offset+2])
		if negotiateRequest.Dialects[i] == Dialect311 {
			supports311 = true
		}
	}
	if supports311 {
		negotiateRequest.NegotiateContextOffset = binary.LittleEndian.Uint32(body[28:32])
		negotiateRequest.NegotiateContextCount = binary.LittleEndian.Uint16(body[32:34])
		negotiateRequest.Reserved2 = binary.LittleEndian.Uint16(body[34:36])
	} else {
		negotiateRequest.ClientStartTime = binary.LittleEndian.Uint64(body[28:36])
	}

	return negotiateRequest, nil

}

func selectDialect(clientDialects []uint16) (uint16, error) {

	for _, supported := range supportedDialects {
		if slices.Contains(clientDialects, supported) {
			return supported, nil
		}
	}
	return 0, fmt.Errorf("[selectDialect] no supported dialect among %#x", clientDialects)
}

func marshalNegotiateResponse(resp NegotiateResponse) []byte {
	buf := make([]byte, negotiateResponseSize)
	binary.LittleEndian.PutUint16(buf[0:2], negotiateResponseSize+1)
	binary.LittleEndian.PutUint16(buf[2:4], resp.SecurityMode)
	binary.LittleEndian.PutUint16(buf[4:6], resp.DialectRevision)
	binary.LittleEndian.PutUint16(buf[6:8], resp.NegotiateContextCount)
	copy(buf[8:24], resp.ServerGUID[:])
	binary.LittleEndian.PutUint32(buf[24:28], resp.Capabilities)
	binary.LittleEndian.PutUint32(buf[28:32], resp.MaxTransactSize)
	binary.LittleEndian.PutUint32(buf[32:36], resp.MaxReadSize)
	binary.LittleEndian.PutUint32(buf[36:40], resp.MaxWriteSize)
	binary.LittleEndian.PutUint64(buf[40:48], resp.SystemTime)
	binary.LittleEndian.PutUint64(buf[48:56], resp.ServerStartTime)
	binary.LittleEndian.PutUint16(buf[56:58], resp.SecurityBufferOffset)
	binary.LittleEndian.PutUint16(buf[58:60], resp.SecurityBufferLength)
	binary.LittleEndian.PutUint32(buf[60:64], resp.NegotiateContextOffset)
	return append(buf, resp.SecurityBuffer...)
}
