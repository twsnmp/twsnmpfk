//go:build windows

package datastore

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// discoverDefaultGateway retrieves the IPv4 default gateway via Windows IP Helper API without executing any external commands.
func discoverDefaultGateway() string {
	var b []byte
	l := uint32(15000)
	flags := uint32(windows.GAA_FLAG_INCLUDE_GATEWAYS | windows.GAA_FLAG_INCLUDE_PREFIX)

	for {
		b = make([]byte, l)
		err := windows.GetAdaptersAddresses(
			windows.AF_INET, // IPv4 only
			flags,
			0,
			(*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])),
			&l,
		)
		if err == nil {
			break
		}
		if err == windows.ERROR_BUFFER_OVERFLOW || err == syscall.ERROR_INSUFFICIENT_BUFFER {
			continue
		}
		return ""
	}

	curr := (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0]))
	for curr != nil {
		// Prefer operational (Up) interfaces
		if curr.OperStatus == windows.IfOperStatusUp {
			gw := curr.FirstGatewayAddress
			for gw != nil {
				ip := gw.Address.IP()
				if ip != nil && ip.To4() != nil && !ip.IsUnspecified() {
					return ip.String()
				}
				gw = gw.Next
			}
		}
		curr = curr.Next
	}

	// Fallback: search all adapters regardless of operational status
	curr = (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0]))
	for curr != nil {
		gw := curr.FirstGatewayAddress
		for gw != nil {
			ip := gw.Address.IP()
			if ip != nil && ip.To4() != nil && !ip.IsUnspecified() {
				return ip.String()
			}
			gw = gw.Next
		}
		curr = curr.Next
	}

	return ""
}
