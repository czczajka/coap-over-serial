package common

import (
	"errors"
	"log"
	"net"
	"time"

	"github.com/tarm/serial"
)

const (
	MTU                = 1024
	SERIAL_BUFFER_SIZE = 2048
)

// SerialPacketConn adapts a serial port to act like a PacketConn for DTLS
type SerialPacketConn struct {
	serialPort *serial.Port
	readBuffer []byte
}

func NewSerialPacketConn(serialPort *serial.Port) *SerialPacketConn {
	log.Println("Enter: NewSerialPacketConn")
	return &SerialPacketConn{
		serialPort: serialPort,
		readBuffer: make([]byte, 2048),
	}
}

// ReadFrom reads data from the serial port and simulates packet behavior
func (s *SerialPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	log.Println("Enter: ReadFrom")
	n, err = s.serialPort.Read(p)
	// Fake address since we don't have real packet addresses in serial
	addr = &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	log.Printf("ReadFrom: read %d bytes", n)
	return n, addr, err
}

// WriteTo writes data to the serial port and simulates packet behavior
func (s *SerialPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	log.Println("Enter: WriteTo")
	n, err = s.serialPort.Write(p)
	log.Printf("WriteTo: wrote %d bytes", n)
	return n, err
}

// Close closes the serial connection
func (s *SerialPacketConn) Close() error {
	log.Println("Enter: Close")
	return s.serialPort.Close()
}

// LocalAddr returns a fake local address (not applicable for serial)
func (s *SerialPacketConn) LocalAddr() net.Addr {
	log.Println("Enter: LocalAddr")
	return &net.UDPAddr{IP: net.IPv4zero, Port: 0}
}

// SetDeadline sets both read and write deadlines
func (s *SerialPacketConn) SetDeadline(t time.Time) error {
	log.Println("Enter: SetDeadline")
	err1 := s.SetReadDeadline(t)
	err2 := s.SetWriteDeadline(t)
	if err1 != nil {
		return err1
	}
	return err2
}

// SetReadDeadline sets a deadline for read operations
func (s *SerialPacketConn) SetReadDeadline(t time.Time) error {
	log.Println("Enter: SetReadDeadline")
	return errors.New("SetReadDeadline not supported for serial connections")
}

// SetWriteDeadline sets a deadline for write operations
func (s *SerialPacketConn) SetWriteDeadline(t time.Time) error {
	log.Println("Enter: SetWriteDeadline")
	return errors.New("SetWriteDeadline not supported for serial connections")
}
