Capture tcp packets from a (mirrored) connection and reassemble them, then forward the packets onto a websocket for use in other applications.  
This implementation looks for packets with a (byte-length,data) payload. It forwards the packets onto port 2222 with route /endpoint.

[GoPacket](https://github.com/google/gopacket) handles all the network logic, so check it out if you want to implement something similar.

## Usage
```
<command> i <interface> f <bpf filter>
```

## Structure
main.go  contains cli settings and program setup  
streamserver.go  contains the websocket and packet forwarding logic  
streamfactory.go  contains logic to filter packets of interest