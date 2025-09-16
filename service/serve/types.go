package serve

import (
	"github.com/up-zero/my-proxy/models"
	"net"
	"sync"
)

type ProxyTask struct {
	models.ProxyBasic
	stopChan chan struct{}
	mu       sync.Mutex

	tcpListener   net.Listener
	tcpActiveConn map[net.Conn]struct{}

	udpListener *net.UDPConn
	udpSessions map[string]net.Conn
}
