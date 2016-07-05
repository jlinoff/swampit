# swampit

This tool sends simple packets of random data in JSON format to a specified IP address at a fixed interval for testing.

The packet layout is a JSON record with three key/value pairs: 'id', 'data' and 'timestamp'. You can add additional data if you want.

The id is the one based sequence number of the record.

The data key/value pair contains 32 bytes of pseudo random ASCII string data. The size of the data can be changed.

The timestamp key/value pair contains the time that the packet was sent.

Here is a very simple example of how to use it.

### First compile it for your system.

```bash
$ go build swampit.go
$ ./swampit -h
```

### Example - In one window start a listener.
You need to start this first to avoid warnings.
```bash
$ timeout 30 nc -l -u -4 localhost:9898
{"id": "1", "data": "g4iy8313", "timestamp": "2016-07-02 00:33:49.844 +0000 UTC"}
{"id": "2", "data": "rf47nmvu", "timestamp": "2016-07-02 00:33:51.848 +0000 UTC"}
{"id": "3", "data": "jgl61cyy", "timestamp": "2016-07-02 00:33:53.85 +0000 UTC"}
{"id": "4", "data": "xm83idt0", "timestamp": "2016-07-02 00:33:55.855 +0000 UTC"}
{"id": "5", "data": "xrda813r", "timestamp": "2016-07-02 00:33:57.86 +0000 UTC"}
```

### Example - In another window start sending stuff.
```bash
$ make
$ ./swampit -p udp -c 10 -i 2 -s 8 -v localhost:9898
2016-07-02 00:33:49.844 +0000 UTC  swampit   88 INFO - {"id": "1", "data": "g4iy8313", "timestamp": "2016-07-02 00:33:49.844 +0000 UTC"}
2016-07-02 00:33:51.849 +0000 UTC  swampit   88 INFO - {"id": "2", "data": "rf47nmvu", "timestamp": "2016-07-02 00:33:51.848 +0000 UTC"}
2016-07-02 00:33:53.85 +0000 UTC   swampit   88 INFO - {"id": "3", "data": "jgl61cyy", "timestamp": "2016-07-02 00:33:53.85 +0000 UTC"}
2016-07-02 00:33:55.855 +0000 UTC  swampit   88 INFO - {"id": "4", "data": "xm83idt0", "timestamp": "2016-07-02 00:33:55.855 +0000 UTC"}
2016-07-02 00:33:57.86 +0000 UTC   swampit   88 INFO - {"id": "5", "data": "xrda813r", "timestamp": "2016-07-02 00:33:57.86 +0000 UTC"}
```
Once it has been started, the records will start showing up on the listener. You will also see them in swampit because the -v option was specified.

### To get additional help use the -h option

```bash
$ ./swampit -h

USAGE
    swampit [OPTIONS] [IPAddress]

DESCRIPTION
    This tool sends simple packets of random data in JSON format to a specified
    IP address at a fixed interval for testing.

    The packet layout is a JSON record with three key/value pairs: 'id', 'data'
    and 'timestamp'. You can additional data if you want.

    The id is the one based sequence number of the record.

    The data key/value pair contains 32 bytes of pseudo random ASCII string
    data. The size of the data can be changed.

    The timestamp key/value pair contains the time that the packet was sent.

OPTIONS
    -c <NUM>, --count <NUM>
                       Send the specified number of records and stop.
                       The default is 10. To run forever set it to 0.

    -d <KEY> <VALUE>, --data <KEY> <VALUE>
                       Send custom data. This can be useful for adding an
                       identifying header. You can add as many of these as you
                       want.

    --dryrun           Do not actually send any data. Useful for testing.

    -h, --help         This help message.

    -i <SECS>, --interval <SECS>
                       The interval between sends. The default is 1 second.
                       Fractional seconds are allowed (ex. -i 2.5).

    -p <PROTOCOL>, --protocol <PROTOCOL>
                       Select the network protocol you want. Tcp is the default.
                       Known network protocols are:
                           "tcp"
                           "tcp4" - IPv4 only
                           "tcp6" - IPv6 only
                           "udp"
                           "udp4" - IPv4 only
                           "udp6" - IPv6 only

    -s <SIZE>, --size <SIZE>
                       The size of the data packet. The default is 32.

    -v, --verbose      Increase the level of verbosity.
                       Set -v to see each record sent.

    -V, --version      Print the program version and exit.

    -x, --exit-on-write-error
                       Exit if a socket write error is encountered. By default
                       this is a warning to allow you time to set up a listener.

EXAMPLES
    $ # Example 1. Get help.
    $ swampit -h

    $ # Example 2. Send 1000 packets to 172.16.98.47:3012
    $ swampit -c 10000 172.16.98.47:3012

    $ # Example 3. Send 100 packets to 172.16.98.47:3012 with data size 1024
    $ swampit -c 100 -s 1024 172.16.98.47:3012

    $ # Example 4. Send 100 packets to 172.16.98.47:3012 with data size 1024
    $ #            Watch them being sent.
    $ swampit -v -c 100 -s 1024 172.16.98.47:3012

    $ # Example 5. You can listen for traffic using tools like nc.
    $ nc -l -k 172.16.98.47 3012

    $ # Example 6. Simple example that should work on all linux and mac
    $ #            platforms. It will send 20 packets and exit.
    $ swampit -c 20 -i 2 127.0.0.1:8989 -p udp &
    $ nc -l -u 127.0.0.1 8989
```

### Finale

I wrote it in Go 1.6.2 and have tested it on CentOS 7.2 and Mac OS X 10.11.5.

Enjoy!
