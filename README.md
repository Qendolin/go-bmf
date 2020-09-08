# AngleCode BMF parsing in Go
 
Supports text, XML and binary formats in version 3.

## API

`bmf.Parse(data []byte) (bmf.Font, error)`  
Parses AngleCode BMF and automatically chooses the correct format


`bmf.ParseText(data []byte) (bmf.Font, error)`  
Parses AngleCode BMF in text format


`bmf.ParseXML(data []byte) (bmf.Font, error)`  
Parses AngleCode BMF in XML format


`bmf.ParseBinary(data []byte) (bmf.Font, error)`  
Parses AngleCode BMF in binary format
