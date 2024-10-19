package connection

type State int

const (
	// Состояния подключения
	StateHandshake State = iota
	StateRequest
	StateConnecting
	StateForwarding
	StateClosing
)

type Connection struct {
	Fd         int
	State      State
	Buffer     []byte
	PeerFd     int
	AddrType   byte
	DestAddr   string
	DestPort   int
	DnsQueryID uint16
}
