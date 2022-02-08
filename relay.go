package main

import (
	"io"
	"net"
)

type Relay struct {
	Bind   string `json:"bind"`
	Target string `json:"target"`
	Proto  string `json:"proto"`
}

func (r *Relay) Start() error {
	if r.Proto == "tcp" {
		listener, err := net.Listen(r.Proto, r.Bind)
		if err != nil {
			return err
		}
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return err
			}
			go forwardTCP(conn, r.Target, r.Proto)
		}
	} else if r.Proto == "udp" {
		proxy, err := NewUDPProxy(r.Target, r.Bind)
		if err != nil {
			return err
		}
		if err := proxy.Start(); err != nil {
			return err
		}
	}
	return nil
}

func forwardTCP(inbound net.Conn, target string, proto string) error {
	outbound, err := net.Dial(proto, target)
	if err != nil {
		return err
	}
	closer := make(chan struct{}, 2)
	go copyTCP(closer, inbound, outbound)
	go copyTCP(closer, outbound, inbound)
	<-closer
	inbound.Close()
	outbound.Close()
	return nil
}

func copyTCP(closer chan struct{}, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	closer <- struct{}{}
}
