// tcpdump2websocket forwards tcp network streams to a websocket

package main

import (
	"flag"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pfring"
	"github.com/google/gopacket/tcpassembly"
)

func main() {
	var (
		adapter = flag.String("i", "enp0s25", "Network Interface Name")
		filter 	= flag.String("f", "tcp and net 127.0.0.1/8", "BPF filter")
	)

	if ring, err := pfring.NewRing(*adapter, 65565, pfring.FlagPromisc); err != nil {
		log.Panicln("Could not create pfring adapter handle.", err)
	} else if err := ring.SetBPFFilter(*filter); err != nil {
		log.Panicln("Could not set BPF filter", err)
	} else if err := ring.SetSocketMode(pfring.ReadOnly); err != nil {
		log.Panicln("Could not set read only on socket.", err)
	} else if err := ring.SetDirection(pfring.ReceiveOnly); err != nil {
		log.Panicln("Could not set receive only on socket", err)
	} else if err := ring.Enable(); err != nil {
		log.Panicln("Could not enable pfring", err)
	} else {
		defer ring.Close()

		packetSource := gopacket.NewPacketSource(ring, layers.LinkTypeEthernet)
		packets := packetSource.Packets()

		ticker := time.Tick(time.Minute)

		streamServer := NewStreamServer()

		tcpassembly.DefaultAssemblerOptions.MaxBufferedPagesPerConnection = 100
		streamFactory := &payloadFactory{
			webHandle:	streamServer.webHandle,
		}
		streamPool := tcpassembly.NewStreamPool(streamFactory)
		assembler := tcpassembly.NewAssembler(streamPool)

		streamServer.start()

		for {
			select {
			case packet := <-packets:
				networkflow := packet.NetworkLayer().NetworkFlow()
				tcp := packet.TransportLayer().(*layers.TCP)
				assembler.Assemble(networkflow, tcp)
			case <-ticker:
				assembler.FlushOlderThan(time.Now().Add(time.Minute * -2))
			}

		}
	}
}
