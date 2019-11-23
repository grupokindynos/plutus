package config

import (
	"errors"
	"fmt"
	"github.com/grupokindynos/common/aes"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrorRpcConnection           = errors.New("unable to perform rpc call")
	ErrorRpcDeserialize          = errors.New("unable to deserialize rpc response")
	ErrorUnableToSend            = errors.New("unable to send transaction")
	ErrorUnableToValidateAddress = errors.New("unable to validate address")
	ErrorUnmarshal               = errors.New("unable to unmarshal object")
	HttpClient                   = &http.Client{
		Timeout: time.Second * 5,
	}
)

type Endpoint struct {
	Host string
	Port int
	User string
}

func NewEndpoint(s string) *Endpoint {
	endpoint := &Endpoint{
		Host: s,
	}
	if parts := strings.Split(endpoint.Host, "@"); len(parts) > 1 {
		endpoint.User = parts[0]
		endpoint.Host = parts[1]
	}
	if parts := strings.Split(endpoint.Host, ":"); len(parts) > 1 {
		endpoint.Host = parts[0]
		endpoint.Port, _ = strconv.Atoi(parts[1])
	}
	return endpoint
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

type SSHTunnel struct {
	Local     *Endpoint
	Server    *Endpoint
	Remote    *Endpoint
	Config    *ssh.ClientConfig
	CloseChan chan interface{}
}

func (tunnel *SSHTunnel) Start(mainWg *sync.WaitGroup) error {
	listener, err := net.Listen("tcp", tunnel.Local.String())
	if err != nil {
		return err
	}
	mainWg.Done()
	tunnel.Local.Port = listener.Addr().(*net.TCPAddr).Port
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go tunnel.forward(conn, &wg)
		wg.Wait()
		break
	}
	err = listener.Close()
	if err != nil {
		return err
	}
	return nil
}

func (tunnel *SSHTunnel) forward(localConn net.Conn, wg *sync.WaitGroup) {
	serverConn, err := ssh.Dial("tcp", tunnel.Server.String(), tunnel.Config)
	if err != nil {
		return
	}
	remoteConn, err := serverConn.Dial("tcp", tunnel.Remote.String())
	if err != nil {
		return
	}
	copyConn := func(writer, reader net.Conn) {
		_, _ = io.Copy(writer, reader)
	}
	go func() {
		copyConn(localConn, remoteConn)
	}()
	go func() {
		copyConn(remoteConn, localConn)
	}()
	<-tunnel.CloseChan
	_ = localConn.Close()
	_ = serverConn.Close()
	_ = remoteConn.Close()
	wg.Done()
}

func (tunnel *SSHTunnel) Close() error {
	tunnel.CloseChan <- struct{}{}
	return nil
}

func NewSSHTunnel(tunnel string, auth ssh.AuthMethod, destination string) *SSHTunnel {
	// A random port will be chosen for us.
	localEndpoint := NewEndpoint("localhost")
	server := NewEndpoint(tunnel)

	sshTunnel := &SSHTunnel{
		Config: &ssh.ClientConfig{
			User: server.User,
			Auth: []ssh.AuthMethod{auth},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
			Timeout: time.Second * 5,
		},
		Local:     localEndpoint,
		Server:    server,
		Remote:    NewEndpoint(destination),
		CloseChan: make(chan interface{}),
	}

	return sshTunnel
}

func PrivateKey(pvKeyString string) ssh.AuthMethod {
	encryptionPass := []byte(os.Getenv("KEY_PASSWORD"))
	decrypted, err := aes.Decrypt(encryptionPass, pvKeyString)
	if err != nil {
		return nil
	}
	key, err := ssh.ParsePrivateKey([]byte(decrypted))
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}
