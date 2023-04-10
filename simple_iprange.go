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
		octets := strings.Split(ipRange, ".")
		if len(octets) == 4 {
			firstIPStr, lastIPStr := "", ""
			valid := true

			for i, octet := range octets {
				if strings.Contains(octet, "-") {
					limits := strings.Split(octet, "-")
					if len(limits) == 2 {
						lower, err := strconv.Atoi(limits[0])
						if err != nil || lower < 0 || lower > 255 {
							valid = false
							break
						}

						upper, err := strconv.Atoi(limits[1])
						if err != nil || upper < 0 || upper > 255 || upper < lower {
							valid = false
							break
						}

						if i == 0 {
							firstIPStr += fmt.Sprintf("%d.", lower)
							lastIPStr += fmt.Sprintf("%d.", upper)
						} else {
							firstIPStr += fmt.Sprintf("%d.", lower)
							lastIPStr += fmt.Sprintf("%d.", upper)
						}
					} else {
						valid = false
						break
					}
				} else {
					num, err := strconv.Atoi(octet)
					if err != nil || num < 0 || num > 255 {
						valid = false
						break
					}
					if i == 0 {
						firstIPStr += fmt.Sprintf("%d.", num)
						lastIPStr += fmt.Sprintf("%d.", num)
					} else {
						firstIPStr += fmt.Sprintf("%d.", num)
						lastIPStr += fmt.Sprintf("%d.", num)
					}
				}
			}

			if valid {
				firstIP := net.ParseIP(firstIPStr[:len(firstIPStr)-1])
				lastIP := net.ParseIP(lastIPStr[:len(lastIPStr)-1])

				if firstIP != nil && lastIP != nil {
					return &IPRange{FirstIP: firstIP, LastIP: lastIP}, nil
				}
			}
		}
	} else if strings.Contains(ipRange, "/") {
		_, ipNet, err := net.ParseCIDR(ipRange)
		if err == nil {
			firstIP, lastIP := networkEndpoints(ipNet)
			return &IPRange{FirstIP: firstIP, LastIP: lastIP}, nil
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
		ip := net.ParseIP(ipRange)
		if ip != nil {
			return &IPRange{FirstIP: ip, LastIP: ip}, nil
		}
	}
	return nil, errors.New("Invalid IP range")
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
