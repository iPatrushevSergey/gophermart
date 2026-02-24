package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Address is a custom flag type for the address 'host:port'.
type Address struct {
	Schema string
	Host   string
	Port   int
}

// Set implements flag.Value.
func (a *Address) Set(s string) error {
	if !(strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")) {
		s = "http://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}

	if u.Host == "" {
		return errors.New("host is empty")
	}

	hostName, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}

	a.Schema = u.Scheme
	a.Host = hostName
	a.Port = port

	return nil
}

// String implements flag.Value.
func (a *Address) String() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

// UnmarshalText supports env parsing.
func (a *Address) UnmarshalText(text []byte) error {
	return a.Set(string(text))
}

// URL returns the full URL (schema://host:port).
func (a *Address) URL() string {
	return fmt.Sprintf("%s://%s:%d", a.Schema, a.Host, a.Port)
}

// Duration is a custom flag type for time.Duration.
type Duration struct {
	time.Duration
}

// Set implements flag.Value.
func (d *Duration) Set(s string) error {
	if val, err := strconv.Atoi(s); err == nil {
		d.Duration = time.Duration(val) * time.Second
		return nil
	}
	val, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = val
	return nil
}

// String implements flag.Value.
func (d *Duration) String() string {
	return d.Duration.String()
}

// UnmarshalText supports env parsing.
func (d *Duration) UnmarshalText(text []byte) error {
	return d.Set(string(text))
}

// BCryptCost is bcrypt cost factor (4-31). Validates on parse.
type BCryptCost int

// Set implements flag.Value.
func (b *BCryptCost) Set(s string) error {
	val, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	if val < 4 || val > 31 {
		return fmt.Errorf("bcrypt cost must be 4-31, got %d", val)
	}
	*b = BCryptCost(val)
	return nil
}

// String implements flag.Value.
func (b *BCryptCost) String() string {
	return strconv.Itoa(int(*b))
}

// UnmarshalText supports env parsing.
func (b *BCryptCost) UnmarshalText(text []byte) error {
	return b.Set(string(text))
}
