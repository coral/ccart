package caddycfg

import (
	"encoding/json"
	"errors"
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

// JServer used for marshaling
type JServer Server

// Server is the backend config
type Server struct {
	Listen                []string              `json:"listen,omitempty"` //The array of listener addresses to bind to.
	ReadTimeout           string                `json:"read_timeout,omitempty"`
	ReadHeaderTimeout     string                `json:"read_header_timeout,omitempty"`
	WriteTimeout          string                `json:"write_timeout,omitempty"`
	IdleTimeout           string                `json:"idle_timeout,omitempty"`
	MaxHeaderBytes        int64                 `json:"max_header_bytes,omitempty"`
	Routes                []Route               `json:"routes,omitempty"`
	Errors                Errors                `json:"errors,omitempty"`
	TLSConnectionPolicies []TLSConnectionPolicy `json:"tls_connection_policies,omitempty"`
	AutomaticHTTPS        AutomaticHTTPS        `json:"automatic_https,omitempty"`
	MaxRehandles          int                   `json:"max_rehandles,omitempty"`
	StrictSniHost         bool                  `json:"strict_sni_host,omitempty"`
	ExperimentalHTTP3     bool                  `json:"experimental_http3,omitempty"`
	sync.Mutex
}

// `json:",omitempty"`

// AutomaticHTTPS Automatic HTTPS configuration
type AutomaticHTTPS struct {
	Disable bool `json:"disable"`
}

// Route holds the route config
type Route struct {
	Handle []Handle `json:"handle,omitempty"`
	Match  []Match  `json:"match,omitempty"`
}

// Errors ...
type Errors map[string]interface{}

//TLSConnectionPolicy TLS connection policy
type TLSConnectionPolicy map[string]string

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
	Path []string `json:"path,omitempty"`
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

var ErrRouteAlreadyExists = errors.New("route already exists")

// AddRoute adds a route to the server
func (s *Server) AddRoute(newRoute Route) error {
	s.Lock()
	defer s.Unlock()
	for _, r := range s.Routes {
		if reflect.DeepEqual(r, newRoute) {
			return ErrRouteAlreadyExists
		}
	}
	glog.Infof("adding new route: %+v", newRoute)
	s.Routes = append(s.Routes, newRoute)
	return nil
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

// ParseJSON ...
func (s *Server) ParseJSON() ([]byte, error) {
	s.Lock()
	defer s.Unlock()
	// This works because JObject doesn't have a MarshalJSON function associated with it
	return json.Marshal(s)
}
