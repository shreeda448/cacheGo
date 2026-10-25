package server

import (
	"log"
	"net"
	"syscall"

	"github.com/shreeda448/cacheGo.git/config"
	"github.com/shreeda448/cacheGo.git/core"
)

var numClients int = 0

func RunAsyncTCPServer() error {
	log.Println("starting an asynchronous TCP server on", config.Host, config.Port)
	maxClients := 20000

	// Create EPOLL event object to hold events
	events := make([]syscall.EpollEvent, maxClients)

	// Create a socket
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	//  0 auto detects the protocol as 'tcp'
	if err != nil {
		return err
	}
	defer syscall.Close(serverFD)

	// set the socket operate in a non-blocking mode
	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}

	// bind the ip and port
	myIP := net.ParseIP(config.Host)
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{myIP[0], myIP[1], myIP[2], myIP[3]},
	}); err != nil {
		return err
	}

	// start listening
	if err = syscall.Listen(serverFD, maxClients); err != nil {
		return err
	}

	// async io starts here

	// creating an epoll instance
	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(epollFD)

	// specify the events we want to get hints about
	// and the set the socket on which
	sockServerEvent := syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFD),
	}

	// listen to read events on the server itself
	if err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, &sockServerEvent); err != nil {
		return err
	}

	for {
		// see if any FD is ready for an IO
		nevents, err := syscall.EpollWait(epollFD, events[:], -1)
		if err != nil {
			continue
		}
		for i := 0; i < nevents; i++ {
			// if the socket server itself is ready for an IO
			if int(events[i].Fd) == serverFD {
				// accept the incoming connection from a client
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("err", err)
					continue
				}

				// increase the number of concurrent clients
				numClients++
				syscall.SetNonblock(fd, true)

				// add this new TCP connection to be monitored
				sockClientEvent := syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(fd),
				}

				// listen to I/O events from the client
				if err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, &sockClientEvent); err != nil {
					log.Fatal(err)
				}
			} else {
				comm := core.FDComm{Fd: int(events[i].Fd)}
				cmd, err := readCommand(comm)
				if err != nil {
					syscall.Close(int(events[i].Fd))
					numClients--
					continue
				}
				respond(cmd, comm)
			}
		}
	}
}
