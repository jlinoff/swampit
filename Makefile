all: swampit

swampit: swampit.go
	go build $<

# Simple test.
# Run make test01-recv in one window and make test01-send in another.
test01-send: swampit
	./swampit -p udp4 -i 2 -s 32 -v 127.0.0.1:8989

test01-recv: swampit
	timeout 30 nc -4 -l -u 127.0.0.1 8989

# Key/value test.
test02-send: swampit
	./swampit -p udp4 -i 2 -s 32 -v -d Key Value 127.0.0.1:8989

test02-recv: swampit
	timeout 30 nc -4 -l -u 127.0.0.1 8989

# Hostname test.
test03-send: swampit
	./swampit -p udp4 -i 2 -s 32 -v localhost:8989

test03-recv: swampit
	timeout 30 nc -4 -l -u localhost 8989
