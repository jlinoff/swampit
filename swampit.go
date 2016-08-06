package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"path"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type kvpair struct {
	Key   string
	Value string
}

type options struct {
	IPAddr           string
	Count            int
	Interval         float64
	Size             int
	Verbose          int
	Protocol         string
	KvPairs          []kvpair
	Dryrun           bool
	ExitOnWriteError bool
}

func main() {
	opts := getOpts()
	var conn net.Conn
	var err error
	if opts.Dryrun == false {
		conn, err = net.Dial(opts.Protocol, opts.IPAddr)
		if err != nil {
			Err("Could not connect: %v", err)
		}
		defer conn.Close()
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			fmt.Println("\n^C interrupt, exiting.")
			conn.Close()
			os.Exit(1)
		}()
	}
	send(opts, conn)
}

// Send the data.
func send(opts options, conn net.Conn) {
	kvjs := createKvPairsJSON(opts)
	rand.Seed(time.Now().UnixNano())
	var alphabet = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	ms := time.Duration(opts.Interval * 1000)
	d := time.Duration(ms * time.Millisecond)
	var jstr string
	i := 0
	for true {
		i++
		// Create the random data string.
		b := make([]rune, opts.Size)
		for j := range b {
			b[j] = alphabet[rand.Intn(len(alphabet))]
		}
		data := string(b)

		// Create the time stamp.
		ts := time.Now().UTC().Truncate(time.Millisecond).String()

		// Create the JSON string.
		//common := fmt.Sprintf("\"id\": \"%v\", \"timestamp\": \"%v\", \"data\": \"%v\"", i, ts, data)
		common := fmt.Sprintf("\"id\": \"%v\", \"data\": \"%v\", \"timestamp\": \"%v\"", i, data, ts)
		if len(kvjs) == 0 {
			jstr = fmt.Sprintf("{%v}", common)
		} else {
			jstr = fmt.Sprintf("{%v, %v}", common, kvjs)
		}
		if opts.Verbose > 0 {
			Info("%v", jstr)
		}
		jstr += "\n"

		// Send it.
		var n int
		var err error
		if opts.Dryrun == false {
			n, err = conn.Write([]byte(jstr))
			if err != nil {
				if opts.ExitOnWriteError {
					Err("%v", err)
				} else {
					Warn("%v", err)
				}
			}
		} else {
			n = len(jstr)
		}
		if opts.Verbose > 1 {
			Info("wrote %v bytes", n)
		}

		// Wait the appropriate interval unless we are done.
		if opts.Count > 0 && i >= opts.Count {
			break
		}
		time.Sleep(d)
	}
}

// Create the kvpairs JSON string
func createKvPairsJSON(opts options) string {
	kvjs := ""
	if len(opts.KvPairs) > 0 {
		for _, r := range opts.KvPairs {
			kvp := fmt.Sprintf("\"%v\": \"%v\"", r.Key, r.Value)
			if len(kvjs) > 0 {
				kvjs += ", "
			}
			kvjs += kvp
		}
	}
	return kvjs
}

// Get the command line options.
func getOpts() options {
	opts := options{
		Count:            10, // run 10 times
		Interval:         1.0,
		Size:             32,
		Protocol:         "tcp",
		Verbose:          0,
		ExitOnWriteError: false,
	}
	vp := []string{
		"tcp", "tcp4", "tcp6",
		"udp", "udp4", "udp6"}
	sort.Strings(vp)
	for i := 1; i < len(os.Args); i++ {
		opt := os.Args[i]
		switch opt {
		case "-c", "--count":
			opts.Count = getWholeNumberArg(&i, 0)
		case "-d", "--date":
			key := getNextArg(&i)
			val := getNextArg(&i)
			rec := kvpair{Key: key, Value: val}
			opts.KvPairs = append(opts.KvPairs, rec)
		case "--dryrun":
			opts.Dryrun = true
		case "-h", "--help":
			help()
			os.Exit(0)
		case "-i", "--interval":
			opts.Interval = getFloat64Arg(&i, 0)
		case "-p", "--protocol":
			p := getNextArg(&i)
			f := sort.SearchStrings(vp, p)
			if f >= len(vp) || vp[f] != p {
				Err("Invalid protocol specified '%v': %v", p, vp)
			}
			opts.Protocol = p
		case "-P", "--Protocol":
			// Same as -p with no error checking
			opts.Protocol = getNextArg(&i)
		case "-s", "--size":
			opts.Size = getWholeNumberArg(&i, 1)
		case "-x", "-exit-on-write-error":
			opts.ExitOnWriteError = true
		case "-v", "--version:":
			opts.Verbose++
		case "-V", "--version":
			base := path.Base(os.Args[0])
			fmt.Printf("%s 0.1.0\n", base)
			os.Exit(0)
		default:
			if strings.HasPrefix(opt, "-") {
				Err("Unrecognized option '%v'", opt)
			}
			if len(opts.IPAddr) > 0 {
				Err("Attempted to define too many IP addresses: %v", opt)
			}
			checkIPAddr(opt)
			opts.IPAddr = opt
		}
	}
	return opts
}

// Check IP address.
func checkIPAddr(opt string) {
	// IPv4: h1.h2.h3.h4:<port>
	// IPv6: [h1:h2:...:hn]:<port>
	// If the IP address contains a period, then treat it as IPv4,
	// otherwise treat it as IPv6.
	reIPv4 := regexp.MustCompile(`^(\d+\.\d+\.\d+\.\d+):(\d+)$`)
	reIPv6 := regexp.MustCompile(`^\[([:0-9a-fA-F]+)\]:(\d+)$`)
	reHost := regexp.MustCompile(`^([a-zA-Z-0-9\-\_]+):(\d+)$`)

	if reIPv4.MatchString(opt) {
		// Nominally IPv4, now check the address syntax explicitly.
		flds := reIPv4.FindStringSubmatch(opt)
		addr := flds[1]                  // ip address
		port, _ := strconv.Atoi(flds[2]) // port
		ip := net.ParseIP(addr)
		if ip == nil {
			Err("invalid IP address '%v'", addr)
		}
		if port < 1 || port > 65535 {
			Err("invalid IP address port '%v'", opt)
		}
	} else if reIPv6.MatchString(opt) {
		// Nominally IPv4, now check the address syntax explicitly.
		flds := reIPv6.FindStringSubmatch(opt)
		addr := flds[1]                  // ip address
		port, _ := strconv.Atoi(flds[2]) // port
		ip := net.ParseIP(addr)
		if ip == nil {
			Err("invalid IP address '%v'", addr)
		}
		if port < 1 || port > 65535 {
			Err("invalid IP address port '%v'", opt)
		}
	} else if reHost.MatchString(opt) {
		flds := reHost.FindStringSubmatch(opt)
		port, _ := strconv.Atoi(flds[2]) // port
		if port < 1 || port > 65535 {
			Err("invalid IP address port '%v'", opt)
		}
	} else {
		Err("invalid IP address '%v'", opt)
	}
}

// Get the next argument.
func getNextArg(i *int) string {
	*i++
	if *i >= len(os.Args) {
		Err("missing argument for %s", os.Args[*i-1])
	}
	return os.Args[*i]
}

// Get whole number [0..inf].
func getWholeNumberArg(i *int, minval int) int {
	opt := os.Args[*i]
	arg := getNextArg(i)
	val, err := strconv.Atoi(arg)
	if err != nil {
		Err("invalid number for '%v': '%v'", opt, arg)
	}
	if val < minval {
		// 0 means forever
		Err("values less than %v not allowed for '%v': %v", minval, opt, arg)
	}
	return val
}

// Get float64 arg [0..inf].
func getFloat64Arg(i *int, minval float64) float64 {
	opt := os.Args[*i]
	arg := getNextArg(i)
	val, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		Err("invalid number for '%v': '%v'", opt, arg)
	}
	if val < minval {
		// 0 means forever
		Err("values less than %v not allowed for '%v': %v", minval, opt, arg)
	}
	return val
}

// Print the help.
func help() {
	base := path.Base(os.Args[0])
	msg := `
USAGE
    %[1]v [OPTIONS] [IPAddress]

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
    $ %[1]v -h

    $ # Example 2. Send 1000 packets to 172.16.98.47:3012
    $ %[1]v -c 10000 172.16.98.47:3012

    $ # Example 3. Send 100 packets to 172.16.98.47:3012 with data size 1024
    $ %[1]v -c 100 -s 1024 172.16.98.47:3012

    $ # Example 4. Send 100 packets to 172.16.98.47:3012 with data size 1024
    $ #            Watch them being sent.
    $ %[1]v -v -c 100 -s 1024 172.16.98.47:3012

    $ # Example 5. You can listen for traffic using tools like nc.
    $ nc -l -k 172.16.98.47 3012

    $ # Example 6. Simple example that should work on all linux and mac
    $ #            platforms. It will send 20 packets and exit.
    $ %[1]v -c 20 -i 2 127.0.0.1:8989 -p udp &
    $ nc -l -u 127.0.0.1 8989

`
	fmt.Printf(msg, base)
}

// Warn reports a warning message to stdout.
// Called just like fmt.Printf.
func Warn(f string, a ...interface{}) {
	Base(os.Stdout, "WARNING", fmt.Sprintf(f, a...), 2, false)
}

// Info reports an informational message to stdout.
// Called just like fmt.Printf.
func Info(f string, a ...interface{}) {
	Base(os.Stdout, "INFO", fmt.Sprintf(f, a...), 2, false)
}

// Err reports an Error message to stderr and exits.
// Called just like fmt.Printf.
func Err(f string, a ...interface{}) {
	Base(os.Stderr, "ERROR", fmt.Sprintf(f, a...), 2, false)
	os.Exit(1)
}

// Base is the basis for the messages.
func Base(fp *os.File, p string, s string, level int, sf bool) {
	pc, fname, lineno, _ := runtime.Caller(level)
	fname = fname[0 : len(fname)-3]
	//t := time.Now().UTC().Truncate(time.Millisecond).String()
	t := time.Now().UTC().Truncate(time.Millisecond).Format("2006-01-02 15:05:05.000 MST")
	if sf { // show function name
		fct := runtime.FuncForPC(pc).Name()
		fmt.Fprintf(fp, "%-28s %s %s %d %s - %s\n", t, path.Base(fname), fct, lineno, p, s)
	} else {
		fmt.Fprintf(fp, "%-28v %v %4d %s - %s\n", t, path.Base(fname), lineno, p, s)
	}
}
