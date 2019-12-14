package caddycfg

import (
	"reflect"
	"sync"

	"github.com/golang/glog"
)

const (
	// FileServer for static fileserver mode
	FileServer = "file_server"
	// ReverseProxy for proxy mode
	ReverseProxy = "reverse_proxy"
)

// Config holds the mains tructure
type Config struct {
	Apps Apps `json:"apps,omitempty"`
}

// Apps holds our applications
type Apps struct {
	HTTP HTTP `json:"http,omitempty"`
}

// HTTP holds the named servers
type HTTP struct {
	Servers Servers `json:"servers"`
}

// Servers ...
type Servers map[string]Server

// Server is the backend config
type Server struct {
	Listen         []string       `json:"listen,omitempty"`
	Routes         []Route        `json:"routes,omitempty"`
	AutomaticHTTPS AutomaticHTTPS `json:"automatic_https,omitempty"`
	sync.Mutex
}

// AutomaticHTTPS Automatic HTTPS configuration
type AutomaticHTTPS struct {
	Disable bool `json:"disable"`
}

// Route holds the route config
type Route struct {
	Handle []Handle `json:"handle,omitempty"`
	Match  []Match  `json:"match,omitempty"`
}

// Handle config
type Handle struct {
	Handler   string     `json:"handler,omitempty"`
	Root      string     `json:"root,omitempty"`
	Upstreams []Upstream `json:"upstreams,omitempty"`
}

// Upstream config
type Upstream struct {
	Dial string `json:"dial,omitempty"`
}

// Match holds the hostname config
type Match struct {
	Host []string `json:"host,omitempty"`
}

// New returns a caddy config
func New() Config {
	return Config{
		Apps: Apps{
			HTTP: HTTP{
				Servers: Servers{},
			},
		},
	}
}

// AddRoute adds a route to the server
func (s *Server) AddRoute(newRoute Route) {
	s.Lock()
	defer s.Unlock()
	for _, r := range s.Routes {
		if reflect.DeepEqual(r, newRoute) {
			glog.Info("route already exists")
			return
		}
	}
	glog.Infof("adding new route: %+v", newRoute)
	s.Routes = append(s.Routes, newRoute)
}

// DeleteRoute removes a route from the tree
func (s *Server) DeleteRoute(oldRoute Route) {
	s.Lock()
	defer s.Unlock()
	for i, r := range s.Routes {
		if reflect.DeepEqual(r, oldRoute) {
			s.Routes = append(s.Routes[:i], s.Routes[i+1:]...)
			return
		}
	}
	glog.Error("route for deletion not found")
}
