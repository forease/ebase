package ebase

import (
	"net"
	"net/http"
	"reflect"
	"strings"
)

type AuthClients []interface{}

func LoadAuthClients(auths string) AuthClients {
	clients := make(AuthClients, 0)

	ips := strings.Split(auths, ";")
	for _, ip := range ips {
		if ip == "" {
			continue
		}
		_, ipNet, err := net.ParseCIDR(ip)
		if err == nil {
			clients = append(clients, ipNet)
		} else {
			ipHost := net.ParseIP(ip)
			if ipHost != nil {
				clients = append(clients, ipHost)
			}
		}
	}

	return clients
}

func (clients AuthClients) ClientAuthor(ipAddr net.IP) bool {

	for _, client := range clients {
		vv := reflect.TypeOf(client)
		if vv.String() == "*net.IPNet" {
			if client.(*net.IPNet).Contains(ipAddr) {
				return true
			}
		} else if vv.String() == "net.IP" {
			if client.(net.IP).Equal(ipAddr) {
				return true
			}
		}
	}

	return false
}

func GetHost(req *http.Request) string {
	host, _, _ := net.SplitHostPort(req.RemoteAddr)

	return host
}
