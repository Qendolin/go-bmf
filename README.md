# AngelCode BMF parsing in Go
 
Supports parsing text, XML and binary formats in version 3. Also supports writing as XML.

## API

`bmf.Parse(data []byte) (*bmf.Font, error)`  
Parses AngelCode BMF and automatically chooses the correct format


`bmf.ParseText(data []byte) (*bmf.Font, error)`  
Parses AngelCode BMF in text format


`bmf.ParseXML(data []byte) (*bmf.Font, error)`  
Parses AngelCode BMF in XML format


`bmf.ParseBinary(data []byte) (*bmf.Font, error)`  
Parses AngelCode BMF in binary format
