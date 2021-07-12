# AngelCode BMF parsing in Go
[![Go Report Card](https://goreportcard.com/badge/github.com/Qendolin/go-bmf)](https://goreportcard.com/report/github.com/Qendolin/go-bmf)
 
Supports parsing and serializing text, XML and binary formats in version 3.

## API
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Qendolin/go-bmf)](https://pkg.go.dev/github.com/Qendolin/go-bmf)

`bmf.Parse(src io.Reader) (*bmf.Font, error)`  
Parses AngelCode BMF and automatically chooses the correct format


`bmf.ParseText(src io.Reader) (*bmf.Font, error)`  
Parses AngelCode BMF in text format


`bmf.ParseXML(src io.Reader) (*bmf.Font, error)`  
Parses AngelCode BMF in XML format


`bmf.ParseBinary(src io.Reader) (*bmf.Font, error)`  
Parses AngelCode BMF in binary format


`bmf.SerializeBinary(fnt *bmf.Font, dst io.Writer) error`  
Serializes AngelCode BMF in binary format


`bmf.SerializeText(fnt *bmf.Font, dst io.Writer) error`  
Serializes AngelCode BMF in text format

## Issues

If you find any problems please report them. :) 
