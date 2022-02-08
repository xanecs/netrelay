package main

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

type UDPChannel struct {
	lastActivity       time.Time
	mutex              sync.Mutex
	outgoingConnection *net.UDPConn
	incomingAddr       *net.UDPAddr
	incomingConnection *net.UDPConn
}

func NewUDPChannel(incomingConnection *net.UDPConn, incomingAddr *net.UDPAddr, targetAddr *net.UDPAddr, notifyClosed chan *UDPChannel) (*UDPChannel, error) {
	conn, err := net.DialUDP("udp", nil, targetAddr)
	if err != nil {
		log.Println("error opening outgoing connectn: %v", err)
		return nil, err
	}
	channel := &UDPChannel{
		incomingAddr:       incomingAddr,
		incomingConnection: incomingConnection,
		outgoingConnection: conn,
		lastActivity:       time.Now(),
	}

	go channel.Forward()
	go channel.Timeout(notifyClosed)
	return channel, nil
}

func (c *UDPChannel) Send(data []byte) (int, error) {
	c.mutex.Lock()
	c.lastActivity = time.Now()
	c.mutex.Unlock()
	return c.outgoingConnection.Write(data)
}

func (c *UDPChannel) Forward() {
	buf := make([]byte, 2048)
	for {
		n, _, err := c.outgoingConnection.ReadFromUDP(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Printf("ending listener on %v", c.outgoingConnection.LocalAddr())
				return
			}
			log.Print(err)
			continue
		}
		if _, err := c.incomingConnection.WriteToUDP(buf[0:n], c.incomingAddr); err != nil {
			log.Print(err)
			continue
		}
		c.mutex.Lock()
		c.lastActivity = time.Now()
		c.mutex.Unlock()
	}
}

func (c *UDPChannel) Timeout(notify chan *UDPChannel) {
	c.mutex.Lock()
	timeout := c.lastActivity.Add(30 * time.Second)
	c.mutex.Unlock()
	now := time.Now()
	for timeout.After(now) {
		time.Sleep(timeout.Sub(now))
		c.mutex.Lock()
		timeout = c.lastActivity.Add(30 * time.Second)
		c.mutex.Unlock()
		now = time.Now()
	}
	notify <- c
}

func (c *UDPChannel) Close() {
	c.outgoingConnection.Close()
}

type UDPProxy struct {
	channels   map[string]*UDPChannel
	mutex      sync.Mutex
	targetAddr *net.UDPAddr
	bindAddr   *net.UDPAddr
}

func NewUDPProxy(target string, bind string) (*UDPProxy, error) {
	targetAddr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		return nil, err
	}
	bindAddr, err := net.ResolveUDPAddr("udp", bind)
	if err != nil {
		return nil, err
	}
	return &UDPProxy{
		channels:   make(map[string]*UDPChannel),
		targetAddr: targetAddr,
		bindAddr:   bindAddr,
	}, nil
}

func (p *UDPProxy) Start() error {
	incomingConn, err := net.ListenUDP("udp", p.bindAddr)
	if err != nil {
		log.Printf("could not listen on udp: %v", err)
		return err
	}
	buf := make([]byte, 2048)
	notifyClosed := make(chan *UDPChannel)
	go p.close(notifyClosed)
	for {
		n, addr, err := incomingConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("could not read from incoming udp: %v", err)
		}
		addrStr := addr.String()
		p.mutex.Lock()
		channel, ok := p.channels[addrStr]
		if !ok {
			channel, err = NewUDPChannel(incomingConn, addr, p.targetAddr, notifyClosed)
			if err != nil {
				p.mutex.Unlock()
				continue
			}
			p.channels[addrStr] = channel
		}
		channel.Send(buf[0:n])
		p.mutex.Unlock()
	}
}

func (p *UDPProxy) close(notify chan *UDPChannel) {
	for channel := range notify {
		addr := channel.incomingAddr
		log.Printf("removing channel %v", addr)
		p.mutex.Lock()
		channel.Close()
		delete(p.channels, addr.String())
		p.mutex.Unlock()
	}
}
