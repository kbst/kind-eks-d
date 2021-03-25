/*
This file is here because flags seemed to mostly be in staging/k8s.io/apiserver. Is there a better place for this?
*/
package options

import (
	"context"
	"encoding/csv"
	"net"
	"strings"
)

// IPNetSlice is a flag for comma-separated slices of CIDR addresses
type IPNetSlice []net.IPNet

// String satisfies pflag.Value
func (netSlice IPNetSlice) String() string {
	netStrings := []string{}
	for _, n := range netSlice {
		netStrings = append(netStrings, n.String())
	}
	return strings.Join(netStrings, ",")
}

// Set satisfies pflag.Value
func (netSlice *IPNetSlice) Set(value string) error {
	cidrStrings, err := readAsCSV(value)
	if err != nil {
		return err
	}
	for _, v := range cidrStrings {
		_, n, err := net.ParseCIDR(strings.TrimSpace(v))
		if err != nil {
			return err
		}
		*netSlice = append(*netSlice, *n)
	}
	return nil
}

func readAsCSV(val string) ([]string, error) {
	if val == "" {
		return []string{}, nil
	}
	stringReader := strings.NewReader(val)
	csvReader := csv.NewReader(stringReader)
	return csvReader.Read()
}

// Type satisfies plfag.Value
func (netSlice *IPNetSlice) Type() string {
	return "[]net.IPNet"
}

// ContainsHost checks if all the IPs for a given hostname are in the allowlist
func (netSlice *IPNetSlice) ContainsHost(ctx context.Context, host string) (bool, error) {
	r := net.Resolver{}
	resp, err := r.LookupIPAddr(ctx, host)
	if err != nil {
		return false, err
	}
	for _, host := range resp {
		// reject if any of the IPs for a hostname are not in the allowlist
		if !netSlice.Contains(host.String()) {
			return false, nil
		}
	}
	return true, nil
}

// Contains checks if a given IP is in the allowlist
func (netSlice *IPNetSlice) Contains(ip string) bool {
	// if there are no allowlists, everything is allowed
	if len(*netSlice) == 0 {
		return true
	}
	netIP := net.ParseIP(ip)
	for _, cidr := range *netSlice {
		if cidr.Contains(netIP) {
			return true
		}
	}
	return false
}

// NewIPNetSlice creates a new IPNetSlice for a given list of networks
func NewIPNetSlice(nets ...string) IPNetSlice {
	netSlice := &IPNetSlice{}
	for _, network := range nets {
		netSlice.Set(network)
	}
	return *netSlice
}
