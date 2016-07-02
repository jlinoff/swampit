# swampit

This tool sends simple packets of random data in JSON format to a specified IP address at a fixed interval for testing.

The packet layout is a JSON record with three key/value pairs: 'id', 'data' and 'timestamp'. You can add additional data if you want.

The id is the one based sequence number of the record.

The data key/value pair contains 32 bytes of pseudo random ASCII string data. The size of the data can be changed.

The timestamp key/value pair contains the time that the packet was sent.
