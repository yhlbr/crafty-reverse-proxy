package main

import (
	"fmt"
	"io"
	"log"
	"net"
    "go.xela.tech/abuseipdb"
)

func checkIpBlock(ip string) bool {
	for _, blockedIp := range getConfig().Blocks {
		if blockedIp == ip {
			return true
		}
	}
	for _, allowedIp := range getConfig().Allows {
		if allowedIp == ip {
			return false
		}
	}

	// Check with AbuseIPDB
	if getConfig().AbuseIPDBKey != "" {
		client := abuseipdb.NewClient(getConfig().AbuseIPDBKey)
		report, err := client.Check(ip)
		if err != nil {
			log.Printf("Error checking IP with AbuseIPDB: %s", err)
			return false
		}
		if report.Data.AbuseConfidenceScore >= 30 {
			log.Printf("Blocked connection from %s due to high abuse confidence score (%d)", ip, report.Data.AbuseConfidenceScore)
			addIpBlock(ip)
			return true
		} else {
			log.Printf("Allowed connection from %s with abuse confidence score (%d)", ip, report.Data.AbuseConfidenceScore)
			addIpAllow(ip)
			return false
		}
	}
	return false
}

func handleClient(client net.Conn, target string, server ServerType, protocol string) {
	remoteAddr := client.RemoteAddr().(*net.TCPAddr)
	log.Printf("Connection from %s\n", remoteAddr.IP.String())

	// Check if the ip is blocked using GetConfig().Blocks
	if checkIpBlock(remoteAddr.IP.String()) {
		log.Printf("Blocked connection from %s\n", remoteAddr.IP.String())
		client.Close()
		return
	}

	incrementPlayerCount(server)
	defer decrementPlayerCount(server)
	serverConnection, err := net.Dial(protocol, target)
	if err != nil {
		println("Server not up and running: " + err.Error() + "\n")
		startMcServer(server)
		serverConnection = awaitForServerStart(protocol, target)
		if serverConnection == nil {
			client.Close()
			return
		}
	}

	defer serverConnection.Close()
	defer client.Close()

	go func() {
		_, err := io.Copy(client, serverConnection)
		log.Printf("User connected!\n")
		if err != nil {
			log.Printf("Error copying from server to client: %s", err)
		}
	}()

	_, err = io.Copy(serverConnection, client)
	if err != nil {
		log.Printf("Error copying from client to server: %s", err)
	}
}

func handleMainServer(server ServerType) {
	listenAddr := server.ExternalIp + ":" + server.ExternalPort
	targetAddr := server.InternalIp + ":" + server.InternalPort

	listener, err := net.Listen(server.Protocol, listenAddr)
	if err != nil {
		log.Fatalf("Error starting "+server.Protocol+" server: %s\n", err)
	}
	defer func() {
		listener.Close()
		println("Listener closed for external port: " + server.ExternalPort + "\n")
	}()

	fmt.Printf(server.Protocol+" reverse proxy running on %s, forwarding to %s\n", listenAddr, targetAddr)

	for {
		client, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %s", err)
			continue
		}

		go handleClient(client, targetAddr, server, server.Protocol)
	}
}

func handleSubServers(subServer OthersType, server ServerType) {
	listenAddr := subServer.ExternalIp + ":" + subServer.ExternalPort
	targetAddr := subServer.InternalIp + ":" + subServer.InternalPort

	listener, err := net.Listen(subServer.Protocol, listenAddr)
	if err != nil {
		log.Fatalf("Error starting "+server.Protocol+" server: %s", err)
	}
	defer listener.Close()

	fmt.Printf(subServer.Protocol+" reverse proxy running on %s, forwarding to %s\n", listenAddr, targetAddr)

	for {
		client, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %s", err)
			continue
		}

		go handleClient(client, targetAddr, server, subServer.Protocol)
	}
}

func handleServer(server ServerType) {
	handleMainServer(server)

	for _, subServer := range server.Others {
		handleSubServers(subServer, server)
	}
}
