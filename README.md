# AngelCode BMF parsing in Go
[![Go Report Card](https://goreportcard.com/badge/github.com/Qendolin/go-bmf)](https://goreportcard.com/report/github.com/Qendolin/go-bmf)
 
Supports parsing text, XML and binary formats in version 3. Also supports writing as XML.

## API
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Qendolin/go-bmf)](https://pkg.go.dev/github.com/Qendolin/go-bmf)

`bmf.Parse(data []byte) (*bmf.Font, error)`  
Parses AngelCode BMF and automatically chooses the correct format


`bmf.ParseText(data []byte) (*bmf.Font, error)`  
Parses AngelCode BMF in text format


`bmf.ParseXML(data []byte) (*bmf.Font, error)`  
Parses AngelCode BMF in XML format


`bmf.ParseBinary(data []byte) (*bmf.Font, error)`  
Parses AngelCode BMF in binary format

## Issues

If you find any problems please report them. :) 
