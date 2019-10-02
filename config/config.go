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
	"time"
)

var (
	ErrorRpcConnection           = errors.New("unable to perform rpc call")
	ErrorRpcDeserialize          = errors.New("unable to deserialize rpc response")
	ErrorUnableToSend            = errors.New("unable to send transaction")
	ErrorUnableToValidateAddress = errors.New("unable to validate address")
	ErrorNoHeaderSignature       = errors.New("no signature found in header")
	ErrorSignatureParse          = errors.New("could not parse header signature")
	ErrorUnmarshal               = errors.New("unable to unmarshal object")
	ErrorWrongMessage            = errors.New("signed message is not on known hosts")
	ErrorInvalidPassword         = errors.New("could not decrypt using master password")
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
	Local  *Endpoint
	Server *Endpoint
	Remote *Endpoint
	Config *ssh.ClientConfig
}

func (tunnel *SSHTunnel) Start() error {
	listener, err := net.Listen("tcp", tunnel.Local.String())
	if err != nil {
		return err
	}
	defer listener.Close()
	tunnel.Local.Port = listener.Addr().(*net.TCPAddr).Port
	for {
		errorChan := make(chan error)
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			err := tunnel.forward(conn)
			if err != nil {
				errorChan <- err
				return
			}
		}()
		select {
		case err := <-errorChan:
			return err
		}
	}
}

func (tunnel *SSHTunnel) forward(localConn net.Conn) error {
	serverConn, err := ssh.Dial("tcp", tunnel.Server.String(), tunnel.Config)
	if err != nil {
		return err
	}
	remoteConn, err := serverConn.Dial("tcp", tunnel.Remote.String())
	if err != nil {
		return err
	}
	copyConn := func(writer, reader net.Conn) error {
		_, err = io.Copy(writer, reader)
		if err != nil {
			return err
		}
		return nil
	}
	errorChan := make(chan error)
	go func() {
		err := copyConn(localConn, remoteConn)
		if err != nil {
			errorChan <- err
		}
	}()
	go func() {
		err := copyConn(remoteConn, localConn)
		if err != nil {
			errorChan <- err
		}
	}()
	select {
	case err := <-errorChan:
		return err
	}
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
		Local:  localEndpoint,
		Server: server,
		Remote: NewEndpoint(destination),
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
