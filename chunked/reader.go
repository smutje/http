package chunked

import (
  "io"
  "fmt"
  "bufio"
  "errors"
)

const (
  header = iota
  body   = iota
  eof    = iota
)

type Reader struct {
  buf     *bufio.Reader
  rest    int
  Error   error
}

func NewReader(rd io.Reader) *Reader{
  buf, ok := rd.(*bufio.Reader)
  if !ok {
    buf = bufio.NewReader(rd)
  }
  return &Reader{buf,0, nil}
}

type InvalidLengthError struct{
  b byte
}

func (e *InvalidLengthError) Error() string {
  return fmt.Sprintf("Invalid hex character: %q", e.b)
}

type InvalidDelimiterError struct {
  e byte
  b byte
}

func (e *InvalidDelimiterError) Error() string {
  return fmt.Sprintf("Invalid delimiter: %q (expected %q)", e.b, e.e)
}

var (
  hex_dec = map[byte]uint64{
    '0' : 0,
    '1' : 1,
    '2' : 2,
    '3' : 3,
    '4' : 4,
    '5' : 5,
    '6' : 6,
    '7' : 7,
    '8' : 8,
    '9' : 9,
    'a' :10,
    'b' :11,
    'c' :12,
    'd' :13,
    'e' :14,
    'f' :15,
    'A' :10,
    'B' :11,
    'C' :12,
    'D' :13,
    'E' :14,
    'F' :15,
  }
  LengthOutOfRange = errors.New("Chunk length out of range")
)

const (
  maxInt = uint64(^uint(0) >> 1) 
)

func (r *Reader) readHeader() (int, map[string]string, error) {
  length := uint64(0)
  for {
    c, err := r.buf.ReadByte()
    if err != nil {
      return 0, nil, err
    }
    if n,ok := hex_dec[c] ; ok {
      length = (length << 4) + n
    }else if c == '\r' || c == ';'  {
      // header terminated
      err = r.buf.UnreadByte()
      if err != nil {
        return 0, nil, err
      }
      break
    }else{
      return 0, nil, &InvalidLengthError{c}
    }
    if length > maxInt {
      return 0, nil, LengthOutOfRange
    }
  }
  _, err := r.buf.ReadString('\n')
  if err != nil {
    return 0,nil,err
  }
  return int(length),nil,nil
}

func (r *Reader) readBody(end bool, buf []byte) error {
  _, err := io.ReadFull( r.buf, buf )
  if err != nil {
    return err
  }
  if end {
    return r.discardCrlf()
  }
  return nil
}

func (r *Reader) discardCrlf() error {
  a, err := r.buf.ReadByte()
  if err != nil {
    return err
  }
  if a != '\r' {
    return &InvalidDelimiterError{'\r',a}
  }
  b, err := r.buf.ReadByte()
  if err != nil {
    return err
  }
  if b != '\n' {
    return &InvalidDelimiterError{'\n',b}
  }
  return nil
}

func max(rest, kappa int) (int, bool){
  if( kappa >= rest ){
    return rest, true
  }
  return kappa, false
}

func (r *Reader) Read(b []byte) (int, error) {
  if r.Error != nil {
    return 0, r.Error
  }
  written := int(0)
  kappa   := len(b)
  for{
    if r.rest == 0 {
      // header
      l, _, err := r.readHeader()
      if err != nil {
        r.Error = err
        return 0,r.Error
      }
      if l == 0 {
        // eof
        r.rest = -1
        r.Error = io.EOF
        return written, r.discardCrlf()
      }
      r.rest = int(l)
    }else{
      // body
      read, end := max(r.rest, kappa)
      r.Error = r.readBody(end, b[written:read+written])
      r.rest  -= read
      kappa   -= read
      written += read
    }
  }
  return written, nil
}
