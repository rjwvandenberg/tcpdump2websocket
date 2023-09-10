// Contains the logic for assembling payloads with a length byte first payload stream

package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

type payloadFactory struct{
	webHandle	chan []byte
}

type payloadStream struct{
	net, transport	gopacket.Flow
	reader		tcpreader.ReaderStream
	buffer		[]byte
	webHandle	chan []byte
}

func (p *payloadFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	payload := &payloadStream{
		net:		net,
		transport:	transport,
		reader:		tcpreader.NewReaderStream(),
		buffer:		make([]byte, 0),
		webHandle:	p.webHandle,
	}
	go payload.run()
	return &payload.reader
}

func (p *payloadStream) run() {
	defer p.reader.Close()

	log.Println("Processing stream", p.net, p.transport)
	buffer := bufio.NewReader(&p.reader)

	for {
		payload := make([]byte, 10000)
		if bytesRead, err := buffer.Read(payload); err == io.EOF {
			log.Println("End of stream reached", p.net, p.transport)
			return
		} else if err != nil {
			log.Println("Error reading stream", p.net, p.transport, ":", err)
		} else if bytesRead == 0 {
		} else {
			p.buffer = append(p.buffer, payload[:bytesRead]...)
			for len(p.buffer) > 0 {
				size := int(binary.LittleEndian.Uint16(p.buffer[:2]))
				if size <= 0 {
					p.buffer = make([]byte, 0)
					log.Println("Encountered application data start < 0 size", size, "bytesRead", bytesRead)
				} else if size > len(p.buffer) {
					break
				} else {
					applicationData := p.buffer[:size]
					p.buffer = p.buffer[size:]
					p.webHandle <- applicationData
				}
			}
		}
	}
}
