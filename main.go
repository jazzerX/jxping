package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: ./jxping addr")

	}

	addr := os.Args[1]

	ip, err := net.ResolveIPAddr("ip4", addr)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	ip4conn := conn.IPv4PacketConn()

	ip4conn.SetTTL(64)
	ip4conn.SetControlMessage(ipv4.FlagTTL, true)

	for i := 0; i < 4; i++ {
		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   os.Getegid() & 0xffff,
				Seq:  i,
				Data: make([]byte, 56),
			},
		}

		binMsg, err := msg.Marshal(nil)
		start := time.Now()

		packetConn := ip4conn.PacketConn

		if _, err := packetConn.WriteTo(binMsg, &net.IPAddr{IP: net.ParseIP(ip.String())}); err != nil {
			log.Fatal(err)
		}

		reply := make([]byte, 256)

		err = ip4conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			log.Fatal(err)
		}
		n, cm, _, err := ip4conn.ReadFrom(reply)

		if err != nil && cm == nil {
			log.Fatal(err)
		}

		duration := time.Since(start)

		parsedReply, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), reply[:n])

		if err != nil {
			log.Fatal(err)
		}

		if parsedReply.Type == ipv4.ICMPTypeEchoReply {
			echoReply, ok := msg.Body.(*icmp.Echo)
			if !ok {
				log.Fatal("invalid ICMP Echo Reply message")
			}

			fmt.Printf("%d bytes from %s: seq = %d ttl = %d time = %v\n", n, addr, echoReply.Seq, cm.TTL, duration)
		}
	}
}
