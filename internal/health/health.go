package health

import (
	"bufio"
	"context"
	"net"
	"os"
	"sync"
	"time"
)

// DefaultSocketPath is the default path for the Unix health-check socket.
// Socket is Unix/Linux-only; not supported on native Windows.
// Default is for single-process/container use; set HEALTH_SOCKET for multiple instances on same host.
const DefaultSocketPath = "/tmp/purgebot-health.sock"

// SocketServer listens on a Unix stream socket; each accepted connection
// receives one line: "ok" or "not ready" based on isReady().
type SocketServer struct {
	path    string
	isReady func() bool
	ln      net.Listener
	once    sync.Once
}

// NewSocketServer creates a Unix socket health server. Start with Run().
func NewSocketServer(path string, isReady func() bool) *SocketServer {
	return &SocketServer{path: path, isReady: isReady}
}

// Run removes any stale socket file, then listens and serves in a goroutine.
// Returns immediately; the accept loop runs in a goroutine so the caller does not block.
func (s *SocketServer) Run() {
	_ = os.Remove(s.path) // ignore error; stale socket from previous run or hard kill
	var lc net.ListenConfig
	ln, err := lc.Listen(context.Background(), "unix", s.path)
	if err != nil {
		return
	}
	s.ln = ln
	go s.serve()
}

func (s *SocketServer) serve() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		ok := s.isReady != nil && s.isReady()
		resp := "not ready\n"
		if ok {
			resp = "ok\n"
		}
		_, _ = conn.Write([]byte(resp))
		_ = conn.Close()
	}
}

// Shutdown closes the listener. Safe to call multiple times.
func (s *SocketServer) Shutdown(ctx context.Context) error {
	var err error
	s.once.Do(func() {
		if s.ln != nil {
			err = s.ln.Close()
		}
	})
	return err
}

// Ping connects to the Unix socket at path, reads one line, and returns true
// if the response is "ok". Timeout is 2 seconds (must be less than Docker healthcheck timeout).
func Ping(path string) bool {
	if path == "" {
		path = DefaultSocketPath
	}
	dialer := &net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.DialContext(context.Background(), "unix", path)
	if err != nil {
		return false
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		return false
	}
	return scanner.Text() == "ok"
}
