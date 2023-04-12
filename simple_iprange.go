package simple_iprange

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type IPRange struct {
	FirstIP net.IP
	LastIP  net.IP
}

type IPRangeList []*IPRange

func (ir *IPRange) String() string {
	return fmt.Sprintf("{%s %s}", ir.FirstIP, ir.LastIP)
}

func Parse(ipRange string) (*IPRange, error) {
	if strings.Contains(ipRange, "-") {
		ips := strings.Split(ipRange, "-")
		if len(ips) == 2 {
			octets1 := strings.Split(ips[0], ".")
			octets2 := strings.Split(ips[1], ".")

			if len(octets1) == 4 && len(octets2) == 4 {
				firstIPStr, lastIPStr := "", ""
				valid := true

				for i := 0; i < 4; i++ {
					lower1, err1 := strconv.Atoi(octets1[i])
					lower2, err2 := strconv.Atoi(octets2[i])

					if err1 != nil || err2 != nil || lower1 < 0 || lower1 > 255 || lower2 < 0 || lower2 > 255 || lower1 > lower2 {
						valid = false
						break
					}

					firstIPStr += fmt.Sprintf("%d.", lower1)
					lastIPStr += fmt.Sprintf("%d.", lower2)
				}

				if valid {
					firstIP := net.ParseIP(firstIPStr[:len(firstIPStr)-1])
					lastIP := net.ParseIP(lastIPStr[:len(lastIPStr)-1])

					if firstIP != nil && lastIP != nil {
						return &IPRange{FirstIP: firstIP, LastIP: lastIP}, nil
					}
				}
			}
		}
	} else if strings.Contains(ipRange, "*") {
		firstIP := strings.ReplaceAll(ipRange, "*", "0")
		lastIP := strings.ReplaceAll(ipRange, "*", "255")

		first := net.ParseIP(firstIP)
		last := net.ParseIP(lastIP)

		if first == nil || last == nil {
			return nil, errors.New("Invalid IP range with wildcard")
		}

		return &IPRange{FirstIP: first, LastIP: last}, nil
	} else {
		// Try to parse as a single IP address
		ip := net.ParseIP(ipRange)
		if ip != nil {
			return &IPRange{FirstIP: ip, LastIP: ip}, nil
		}

		// Try to parse as a CIDR format
		ip, ipNet, err := net.ParseCIDR(ipRange)
		if err == nil {
			firstIP := ip
			lastIP := getLastIP(ipNet)
			return &IPRange{FirstIP: firstIP, LastIP: lastIP}, nil
		}
	}
	return nil, errors.New("Invalid IP range")
}
func getLastIP(ipNet *net.IPNet) net.IP {
	ip := make(net.IP, len(ipNet.IP))
	for i := 0; i < len(ip); i++ {
		ip[i] = ipNet.IP[i] | ^ipNet.Mask[i]
	}
	return ip
}

// formatCompleteIP takes a given IP address and the last octet as an integer and returns a complete IP address
func formatCompleteIP(ip string, lastOctet int) string {
	octets := strings.Split(ip, ".")
	octets[len(octets)-1] = strconv.Itoa(lastOctet)
	return strings.Join(octets, ".")
}

func ParseList(ipRanges string) (IPRangeList, error) {
	ipRangeList := []*IPRange{}
	ipRanges = strings.ReplaceAll(ipRanges, "\n", ",")
	ipRanges = strings.ReplaceAll(ipRanges, " ", "")
	ranges := strings.Split(ipRanges, ",")
	for _, ipRange := range ranges {
		ipRange = strings.TrimSpace(ipRange)
		if ipRange == "" {
			continue
		}
		ipr, err := Parse(ipRange)
		if err != nil {
			return nil, fmt.Errorf("Invalid IP rangelist: %v", err)
		}
		ipRangeList = append(ipRangeList, ipr)
	}
	return ipRangeList, nil
}

func networkEndpoints(ipNet *net.IPNet) (net.IP, net.IP) {
	firstIP := ipNet.IP.Mask(ipNet.Mask)
	lastIP := make(net.IP, len(firstIP))
	copy(lastIP, firstIP)
	for i := 0; i < len(firstIP); i++ {
		lastIP[i] |= ^ipNet.Mask[i]
	}
	return firstIP, lastIP
}

// 创建一个 Expander 接口，包含 Expand 方法
type Expander interface {
	Expand() ([]string, error)
}

// 为 IPRange 类型实现 Expand 方法
func (ir *IPRange) Expand() ([]string, error) {
	var ips []string
	ip := ir.FirstIP

	for {
		ips = append(ips, ip.String())

		if ip.Equal(ir.LastIP) {
			break
		}

		if err := inc(ip); err != nil {
			return nil, err
		}
	}

	return ips, nil
}

// 为 IPRangeList 类型实现 Expand 方法
func (list IPRangeList) Expand() ([]string, error) {
	var ips []string
	for _, ir := range list {
		rangeIPs, err := ir.Expand()
		if err != nil {
			return nil, err
		}
		ips = append(ips, rangeIPs...)
	}
	return ips, nil
}

func inc(ip net.IP) error {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
	return nil
}
