package client

import "net"

// Awkward singleton for ignores
type ignoreSingleton struct {
	// username to ignored ipv4
	ignoreMap map[string][]net.IPAddr
}

func (is *ignoreSingleton) getIgnoreList(username string) []net.IPAddr {
	l, ok := is.ignoreMap[username]
	if ok {
		return l
	}

	ret := make([]net.IPAddr, 1)
	is.ignoreMap[username] = ret
	return ret
}
