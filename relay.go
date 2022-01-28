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
		go forward(conn, r.Target, r.Proto)
	}
}

func forward(inbound net.Conn, target string, proto string) error {
	outbound, err := net.Dial(proto, target)
	if err != nil {
		return err
	}
	closer := make(chan struct{}, 2)
	go copy(closer, inbound, outbound)
	go copy(closer, outbound, inbound)
	<-closer
	inbound.Close()
	outbound.Close()
	return nil
}

func copy(closer chan struct{}, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	closer <- struct{}{}
}
