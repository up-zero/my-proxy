package serve

import (
	"github.com/up-zero/my-proxy/models"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
)

type ProxyTask struct {
	models.ProxyBasic
	stopChan chan struct{}
	mu       sync.Mutex
	capture  *CaptureHub
	// statistic
	bytesIn    atomic.Int64 // client -> target
	bytesOut   atomic.Int64 // target -> client
	TrafficIn  int64        `json:"traffic_in"`
	TrafficOut int64        `json:"traffic_out"`

	tcpListener   net.Listener
	tcpActiveConn map[net.Conn]struct{}

	udpListener *net.UDPConn
	udpSessions map[string]net.Conn

	httpServer *http.Server
}
