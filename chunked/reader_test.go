package chunked

import (
  "testing"
  "io"
  "bytes"
  "strings"
  "fmt"
)

func TestReadSimpleHeader(t *testing.T){
  content := "1\r\n"
  rd := NewReader(strings.NewReader(content))
  length, _, err := rd.readHeader()
  if err != nil {
    t.Fatal(err)
  }
  if length != 1 {
    t.Fatalf("Expected length 1, got %d",length)
  }
}
func TestReadMultiByteHeader(t *testing.T){
  content := "cAfFeE\r\n"
  rd := NewReader(strings.NewReader(content))
  length, _, err := rd.readHeader()
  if err != nil {
    t.Fatal(err)
  }
  if length != 13303790 {
    t.Fatalf("Expected length 13303790, got %d",length)
  }
}

func TestReadOverlongHeader(t *testing.T){
  content := fmt.Sprintf("%x\r\n",uint64(maxInt)+1)
  rd := NewReader(strings.NewReader(content))
  _, _, err := rd.readHeader()
  if err != LengthOutOfRange {
    t.Fatalf("Expected a length out of range error, got %q", err)
  }
}



func TestReadBorkedHeader1(t *testing.T){
  content := "orked"
  rd := NewReader(strings.NewReader(content))
  _, _, err := rd.readHeader()
  if err == nil {
    t.Fatal("Expected an error")
  }
}

func TestReadBorkedHeader(t *testing.T){
  content := "1g\r\n"
  rd := NewReader(strings.NewReader(content))
  _, _, err := rd.readHeader()
  if err == nil {
    t.Fatal("Expected an error")
  }
}

func TestReadBody(t *testing.T){
  content := "body\r\n"
  buf := make([]byte,4)
  sr  := strings.NewReader(content)
  rd  := NewReader(sr)
  err := rd.readBody(true, buf)
  if err != nil {
    t.Fatal("Expected an error")
  }
  if bytes.Compare(buf,[]byte{'b','o','d','y'}) != 0 {
    t.Fatalf("Expected 'body', got %i",buf)
  }
  _, err = sr.ReadByte()
  if err != io.EOF {
    t.Fatalf("Expected EOF, got %q",err)
  }
}

func TestRead(t *testing.T){
  content := "4\r\nWiki\r\n5\r\npedia\r\nC\r\n in\r\nchunks.\r\n0\r\n\r\n"
  expected:= "Wikipedia in\r\nchunks."
  result  := new(bytes.Buffer)
  sr      := strings.NewReader(content)
  rd      := NewReader(sr)
  _, err  := io.Copy(result, rd)
  if err != nil {
    t.Fatal(err)
  }
  str     := result.String()
  if str != expected {
    t.Fatalf("Unexpected output: '%q' (expected '%q')", str, expected)
  }
  _, err  = sr.ReadByte()
  if err != io.EOF {
    t.Fatalf("Expected EOF, got %q",err)
  }
}

