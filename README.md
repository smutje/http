# HTTP helper modules

## Chunked Reader

Reads a http chunked encoded stream.

```go
package main

import(
  "net/dial"
  "net/http"
  "net/http/httputil"
  "github.com/smutje/http/chunked"
)

func main(){
  con    := httputil.NewClientConn(net.Dial("tcp","example.com"))
  _, _ := con.Do(http.NewRequest("GET","/",nil))
  // here it comes
  _, rd  := con.Hijack()
  chunked_reader := chunked.NewReader(rd)
  // assuming that the response is chunked, you can now read it from 
  // the chunked reader
}

```
