package connection

const (
	// Состояния подключения
	StateHandshake = iota
	StateRequest
	StateConnecting
	StateForwarding
	StateClosing
)

type Connection struct {
	Fd         int
	State      int
	Buffer     []byte
	PeerFd     int
	AddrType   byte
	DestAddr   string
	DestPort   int
	DnsQueryID uint16
}
