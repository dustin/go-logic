// Package logic provides tools for interacting with Saleae Logic files.
package logic

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Frame represents a value read from a logic data export.
type SerialFrame struct {
	Offset     time.Duration
	Value      byte
	ParityErr  bool
	FramingErr bool
}

// SerialCSVReader emits records read from a Logic trace export.
// Attempts are made to parse the default format, but it's really kind
// of inconsistent.  Best to use the hex format.
type SerialCSVReader struct {
	c *csv.Reader
}

// Next returns the next SerialFrame from a stream.
func (s *SerialCSVReader) Next() (SerialFrame, error) {
	rv := SerialFrame{}
	row, err := s.c.Read()
	if err != nil {
		return rv, err
	}
	if len(row) != 4 {
		return rv, fmt.Errorf("expected 4 columns, got %v", row)
	}
	rv.Offset, err = time.ParseDuration(row[0] + "s")
	if err != nil {
		return rv, err
	}
	if len(row[1]) == 1 {
		rv.Value = row[1][0]
	} else if row[1] == "COMMA" {
		rv.Value = ','
	} else if strings.HasPrefix(row[1], "0x") {
		i, err := strconv.ParseInt(row[1][2:], 16, 16)
		if err != nil {
			return rv, fmt.Errorf("error parsing %v, %v", row, err)
		}
		rv.Value = byte(i)
	} else {
		c := row[1]
		switch c[0] {
		case '"':
			rv.Value = '"'
		case '\'':
			if c[1] == ' ' {
				rv.Value = ' '
			} else {
				i, err := strconv.ParseInt(c[1:len(c)-1], 10, 16)
				if err != nil {
					return rv, fmt.Errorf("error parsing %v, %v", row, err)
				}
				rv.Value = byte(i)
			}
		case '\\':
			switch c[1] {
			case 't':
				rv.Value = '\t'
			case 'r':
				rv.Value = '\r'
			case 'n':
				rv.Value = '\n'
			default:
				return rv, fmt.Errorf("unhandled escape code: %v", c[1])
			}
		default:
			return rv, fmt.Errorf("unhandled value in %v: %q", row, c)
		}
	}
	// TODO: handle the other columns
	return rv, nil
}

func (s *SerialCSVReader) Read(b []byte) (int, error) {
	f, err := s.Next()
	if err != nil {
		return 0, err
	}
	b[0] = f.Value
	return 1, nil
}

// NewSerialCSVReader reads CSV from the input representing a serial
// stream of bytes.
func NewSerialCSVReader(r io.Reader) (*SerialCSVReader, error) {
	c := csv.NewReader(r)
	hdr, err := c.Read()
	if err != nil {
		return nil, err
	}
	if len(hdr) != 4 {
		return nil, fmt.Errorf("Expected 4 columns, got %v", hdr)
	}
	return &SerialCSVReader{c}, nil
}
