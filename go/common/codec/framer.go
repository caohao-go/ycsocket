package codec

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
)

var (
	MaxFrameSize = 10 * 1024 * 1024 // max size of frame.
)

var (
	ErrHeadOverflowsUint16 = errors.New("head len overflows uint16")
	ErrHeadOverflowsUint32 = errors.New("total len overflows uint32")
	ErrFrameTooLarge       = errors.New("length of frame is larger than MaxFrameSize")
)

const (
	// FrameTypeNormal mainly used in a secure environment, such as a private LAN.
	FrameTypeNormal = 0
	FrameHeadLen    = 10 // total length of frame head
	ProtocolVersion = 1  // protocol version  v1

	// FrameTypeSignature mainly used for intercommunication between
	// data in the same cloud through the bastion host.
	FrameTypeSignature = 1  // signature frame
	SignFrameHeadLen   = 47 // total length of sign frame head
	SignVersion        = 1  // version of signature frame

	// FrameTypeEncrypt mainly used for android, iphone and
	// other terminal devices to directly access tenant data,
	// to prevent data from being hijacked.
	FrameTypeEncrypt    = 2  // encrypt frame
	EncryptFrameHeadLen = 15 // total length of encrypt frame head
	EncryptVersion      = 1  // version of encrypt frame

	ProtocolTypeRPC  = 1 // protocol type rpc
	ProtocolTypeHTTP = 2 // protocol type http
)

// FrameHead is head of the rpc frame.
type FrameHead struct {
	FrameType uint8  // type of the frame 0-default
	Version   uint8  // version of protocol
	HeaderLen uint16 // header length
	TotalLen  uint32 // total length
	Reserved  uint16
}

func NewFrameHead() *FrameHead {
	return &FrameHead{
		FrameType: FrameTypeNormal,
		Version:   ProtocolVersion,
	}
}

// Extract extracts field values of the FrameHead from the buffer.
func (h *FrameHead) Extract(buf []byte) {
	h.FrameType = buf[0]
	h.Version = buf[1]
	h.HeaderLen = binary.BigEndian.Uint16(buf[2:4])
	h.TotalLen = binary.BigEndian.Uint32(buf[4:8])
	h.Reserved = binary.BigEndian.Uint16(buf[8:10])
}

// Construct constructs bytes body for the whole frame.
func (h *FrameHead) Construct(header, body []byte) ([]byte, error) {
	headerLen := len(header)
	if headerLen > math.MaxUint16 {
		return nil, ErrHeadOverflowsUint16
	}
	totalLen := int64(FrameHeadLen) + int64(headerLen) + int64(len(body))
	if totalLen > int64(MaxFrameSize) {
		return nil, ErrFrameTooLarge
	}
	if totalLen > math.MaxUint32 {
		return nil, ErrHeadOverflowsUint32
	}

	// construct the buffer
	buf := make([]byte, totalLen)
	buf[0] = h.FrameType
	buf[1] = h.Version
	binary.BigEndian.PutUint16(buf[2:4], uint16(headerLen))
	binary.BigEndian.PutUint32(buf[4:8], uint32(totalLen))
	binary.BigEndian.PutUint16(buf[8:10], h.Reserved)
	copy(buf[FrameHeadLen:FrameHeadLen+headerLen], header)
	copy(buf[FrameHeadLen+headerLen:], body)
	return buf, nil
}

// SignFrameHead is head of signature of frame
type SignFrameHead struct {
	FrameType    uint8  // type of the frame 1-signature frame
	Version      uint8  // version of signature frame
	ProtocolType uint8  // 1-rpc 2-http
	TotalLen     uint32 // total length
	Tenant       uint64 // tenant hash
	Sign         []byte // signature 32 个字节，md5(tenant_id+secret+frame)
}

func NewSignFrameHead() *SignFrameHead {
	return &SignFrameHead{
		FrameType:    uint8(FrameTypeSignature),
		Version:      SignVersion,
		ProtocolType: ProtocolTypeRPC,
	}
}

// Extract extracts field values of the FrameHead from the buffer.
func (h *SignFrameHead) Extract(buf []byte) {
	h.FrameType = buf[0]
	h.Version = buf[1]
	h.ProtocolType = buf[2]
	h.TotalLen = binary.BigEndian.Uint32(buf[3:7])
	h.Tenant = binary.BigEndian.Uint64(buf[7:15])
	h.Sign = buf[15:47]
}

// Construct constructs bytes body for the whole frame.
func (h *SignFrameHead) Construct(tenant uint64, token, frameBody []byte) ([]byte, error) {
	h.Tenant = tenant

	h.TotalLen = uint32(SignFrameHeadLen) + uint32(len(frameBody))
	if h.TotalLen > uint32(MaxFrameSize) {
		return nil, ErrFrameTooLarge
	}

	if h.TotalLen > math.MaxUint32 {
		return nil, ErrHeadOverflowsUint32
	}

	// construct the buffer
	buf := make([]byte, h.TotalLen)
	buf[0] = h.FrameType
	buf[1] = h.Version
	buf[2] = h.ProtocolType
	binary.BigEndian.PutUint32(buf[3:7], h.TotalLen)
	binary.BigEndian.PutUint64(buf[7:15], h.Tenant)

	sumBytes := md5.Sum(append(token, frameBody...))
	signature := make([]byte, hex.EncodedLen(len(sumBytes[:])))
	_ = hex.Encode(signature, sumBytes[:])
	copy(buf[15:47], signature)

	copy(buf[SignFrameHeadLen:], frameBody)
	return buf, nil
}

// EncryptFrameHead 加密帧
type EncryptFrameHead struct {
	FrameType    uint8  // type of the frame 2-encrypt frame
	Version      uint8  // version of encrypt frame
	ProtocolType uint8  // 1-rpc 2-http
	TotalLen     uint32 // total length
	Tenant       uint64 // tenant hash
}

func NewEncryptFrameHead() *EncryptFrameHead {
	return &EncryptFrameHead{
		FrameType:    uint8(FrameTypeEncrypt),
		Version:      EncryptVersion,
		ProtocolType: ProtocolTypeRPC,
	}
}

// Extract extracts field values of the FrameHead from the buffer.
func (h *EncryptFrameHead) Extract(buf []byte) {
	h.FrameType = buf[0]
	h.Version = buf[1]
	h.ProtocolType = buf[2]
	h.TotalLen = binary.BigEndian.Uint32(buf[3:7])
	h.Tenant = binary.BigEndian.Uint64(buf[7:15])
}

// Construct constructs bytes body for the whole frame.
func (h *EncryptFrameHead) Construct(tenant uint64, token, frameBody []byte) ([]byte, error) {
	h.Tenant = tenant

	encryptFrameBody, err := aesEncrypt(frameBody, token)
	if err != nil {
		return nil, err
	}

	h.TotalLen = EncryptFrameHeadLen + uint32(len(encryptFrameBody))
	if h.TotalLen > uint32(MaxFrameSize) {
		return nil, ErrFrameTooLarge
	}

	if h.TotalLen > math.MaxUint32 {
		return nil, ErrHeadOverflowsUint32
	}

	// construct the buffer
	buf := make([]byte, h.TotalLen)
	buf[0] = h.FrameType
	buf[1] = h.Version
	buf[2] = h.ProtocolType
	binary.BigEndian.PutUint32(buf[3:7], h.TotalLen)
	binary.BigEndian.PutUint64(buf[7:15], h.Tenant)

	copy(buf[EncryptFrameHeadLen:], encryptFrameBody)
	return buf, nil
}

func aesEncrypt(orig, key []byte) (buf []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("AES Encrypt Panic: %v", e)
		}
	}()

	block, err := aes.NewCipher(key)
	if block == nil {
		return nil, fmt.Errorf("AES Encrypt NewCipher Error: %v", err)
	}

	blockSize := block.BlockSize()
	orig = pkcs7Padding(orig, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	cryted := make([]byte, len(orig))
	blockMode.CryptBlocks(cryted, orig)

	buf = make([]byte, base64.StdEncoding.EncodedLen(len(cryted)))
	base64.StdEncoding.Encode(buf, cryted)
	return buf, nil
}

// 补码
func pkcs7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
