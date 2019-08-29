package config

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ErrorNoCoin               = errors.New("coin not available")
	ErrorRpcConnection        = errors.New("unable to perform rpc call")
	ErrorRpcDeserialize       = errors.New("unable to deserialize rpc response")
	ErrorNoAuthMethodProvided = errors.New("missing authorization token")
	ErrorNoRpcUserProvided    = errors.New("missing rpc username")
	ErrorNoRpcPassProvided    = errors.New("missing rpc password")
	ErrorNoRpcPortProvided    = errors.New("missing rpc port")
	ErrorNoHostIPProvided     = errors.New("missing host ip")
	ErrorNoHostUserProvided   = errors.New("missing host user")
	ErrorNoHostPortProvided   = errors.New("missing host port")
	ErrorExternalStatusError  = errors.New("unable to get external source status")
	ErrorNoAmountSpecified    = errors.New("missing amount specification")
	ErrorNoAddressSpecified   = errors.New("missing address specification")
	ErrorNoColdAddress        = errors.New("missing cold address to send")
	ErrorNoExchangeAddress    = errors.New("missing exchange address to send")
	ErrorUnableToSend         = errors.New("unable to send transaction")

	HttpClient = &http.Client{
		Timeout: time.Second * 10,
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
	Log    *log.Logger
}

func (tunnel *SSHTunnel) logf(fmt string, args ...interface{}) {
	if tunnel.Log != nil {
		tunnel.Log.Printf(fmt, args...)
	}
}

func (tunnel *SSHTunnel) Start() error {
	listener, err := net.Listen("tcp", tunnel.Local.String())
	if err != nil {
		return err
	}
	defer listener.Close()
	tunnel.Local.Port = listener.Addr().(*net.TCPAddr).Port
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		tunnel.logf("accepted connection")
		go tunnel.forward(conn)
	}
}

func (tunnel *SSHTunnel) forward(localConn net.Conn) {
	serverConn, err := ssh.Dial("tcp", tunnel.Server.String(), tunnel.Config)
	if err != nil {
		tunnel.logf("server dial error: %s", err)
		return
	}
	tunnel.logf("connected to %s (1 of 2)\n", tunnel.Server.String())
	remoteConn, err := serverConn.Dial("tcp", tunnel.Remote.String())
	if err != nil {
		tunnel.logf("remote dial error: %s", err)
		return
	}
	tunnel.logf("connected to %s (2 of 2)\n", tunnel.Remote.String())
	copyConn := func(writer, reader net.Conn) {
		_, err := io.Copy(writer, reader)
		if err != nil {
			tunnel.logf("io.Copy error: %s", err)
		}
	}
	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)
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
		},
		Local:  localEndpoint,
		Server: server,
		Remote: NewEndpoint(destination),
	}

	return sshTunnel
}

func PrivateKey(pvKeyString string) ssh.AuthMethod {
	encryptionPass := []byte(os.Getenv("KEY_PASSWORD"))
	decrypted, err := decrypt(encryptionPass, pvKeyString)
	if err != nil {
		return nil
	}
	key, err := ssh.ParsePrivateKey([]byte(decrypted))
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

// GlobalResponse is used to wrap all the API responses under the same model.
// Automatically detect if there is an error and return status and code according
func GlobalResponse(result interface{}, err error, c *gin.Context) *gin.Context {
	if err != nil {
		c.JSON(500, gin.H{"message": "Error", "error": err.Error(), "status": -1})
	} else {
		c.JSON(200, gin.H{"data": result, "status": 1})
	}
	return c
}

func decrypt(key []byte, securemess string) (decodedmess string, err error) {
	cipherText, err := base64.StdEncoding.DecodeString(securemess)
	if err != nil {
		return
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	if len(cipherText) < aes.BlockSize {
		err = errors.New("ciphertext block size is too short")
		return
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)
	decodedmess = string(cipherText)
	return
}
