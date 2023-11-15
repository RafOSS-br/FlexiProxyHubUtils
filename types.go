package utils

import (
)

type ProxyMode int

const (
	Normal ProxyMode = iota
	AsynchronousReport
)
// PROXY MODE, UPSTREAM ADDR, HOSTNAME, PATH = BOOL
type HostnameAndPortPerPathPerProxyModeAndUpstreamAddr map[int]map[string]map[string]map[string]bool
// UPSTREAM ADDR, HOSTNAME, PATH = BOOL
type HostnameAndPortPerPathAndUpstreamAddr map[string]map[string]map[string]bool
// HOSTNAME, PATH = BOOL
type HostnameAndPortPerPath map[string]map[string]bool

type ProxyToList string


type Listener struct {
	Address string
	Port    string
}

// "routes": [{"host": "localhost", "port": "8080", "path": ["/teste.txt"]}]
type Routes struct {
	Host string   `json:"host"`
	Path []string `json:"path"`
}

type Router struct {
	Routes  []Routes  `json:"routes"`
	Mode    ProxyMode `json:"mode"`
	ProxyTo string    `json:"proxy_to"`
}

type Configuration struct {
	LogLevel          string
	VisibleHeaders    string
	HeaderToReplicate Headers
	BodyMaxLen        int
	Proxy             []Router
	Listener          Listener
}

type Headers []string