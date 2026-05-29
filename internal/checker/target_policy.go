package checker

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

var errBlockedTarget = errors.New("target resolves to a private or internal address")

// ValidateMonitorTarget validates and normalizes user-provided monitor targets.
// It performs syntax checks and blocks IP literals that are never safe to dial.
func ValidateMonitorTarget(checkType, rawTarget string) (string, error) {
	target := strings.TrimSpace(rawTarget)
	if target == "" {
		return "", fmt.Errorf("target URL is required")
	}

	switch normalizedCheckType(checkType) {
	case "ping":
		return validatePingTarget(target)
	case "http", "links":
		return validateHTTPMonitorTarget(target)
	default:
		return "", fmt.Errorf("unsupported check type: %s", checkType)
	}
}

func normalizedCheckType(checkType string) string {
	if strings.TrimSpace(checkType) == "" {
		return "http"
	}
	return strings.ToLower(strings.TrimSpace(checkType))
}

func validateHTTPMonitorTarget(target string) (string, error) {
	parsed, err := parseHTTPURL(target)
	if err != nil {
		return "", err
	}
	if parsed.User != nil {
		return "", fmt.Errorf("URL credentials are not allowed")
	}
	if err := rejectBlockedHostLiteral(parsed.Hostname()); err != nil {
		return "", err
	}
	return parsed.String(), nil
}

func validatePingTarget(target string) (string, error) {
	if strings.Contains(target, "://") {
		return "", fmt.Errorf("ping target must be host[:port], not a URL")
	}

	host := target
	port := ""
	if h, p, err := net.SplitHostPort(target); err == nil {
		host = h
		port = p
	} else if strings.Count(target, ":") == 1 {
		h, p, found := strings.Cut(target, ":")
		if found {
			host = h
			port = p
		}
	}

	host = strings.Trim(host, "[]")
	if host == "" {
		return "", fmt.Errorf("missing target host")
	}
	if strings.ContainsAny(host, "/?#@") {
		return "", fmt.Errorf("invalid target host")
	}
	if port != "" {
		portNum, err := strconv.Atoi(port)
		if err != nil || portNum < 1 || portNum > 65535 {
			return "", fmt.Errorf("invalid target port")
		}
	}
	if err := rejectBlockedHostLiteral(host); err != nil {
		return "", err
	}
	if port != "" {
		return net.JoinHostPort(host, port), nil
	}
	return host, nil
}

func validateOutboundPingTarget(ctx context.Context, target string) error {
	host := target
	if h, _, err := net.SplitHostPort(target); err == nil {
		host = h
	} else if strings.Count(target, ":") == 1 {
		h, _, found := strings.Cut(target, ":")
		if found {
			host = h
		}
	}
	return validateOutboundHost(ctx, strings.Trim(host, "[]"))
}

func validateOutboundHost(ctx context.Context, host string) error {
	_, err := resolvePublicHost(ctx, host)
	return err
}

func resolvePublicHost(ctx context.Context, host string) ([]netip.Addr, error) {
	if err := rejectBlockedHostLiteral(host); err != nil {
		return nil, err
	}

	resolver := net.DefaultResolver
	addrs, err := resolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("target host has no IP addresses")
	}
	for _, addr := range addrs {
		if isBlockedAddr(addr) {
			return nil, fmt.Errorf("%w: %s", errBlockedTarget, addr.String())
		}
	}
	return addrs, nil
}

func rejectBlockedHostLiteral(host string) error {
	addr, err := netip.ParseAddr(strings.Trim(host, "[]"))
	if err != nil {
		return nil
	}
	if isBlockedAddr(addr) {
		return fmt.Errorf("%w: %s", errBlockedTarget, addr.String())
	}
	return nil
}

func isBlockedAddr(addr netip.Addr) bool {
	if addr.Is4In6() {
		addr = addr.Unmap()
	}
	return addr.IsLoopback() ||
		addr.IsPrivate() ||
		addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast() ||
		addr.IsMulticast() ||
		addr.IsUnspecified() ||
		addr == netip.MustParseAddr("169.254.169.254") ||
		addr == netip.MustParseAddr("100.100.100.200")
}
