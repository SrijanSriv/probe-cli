package netxlite

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/ooni/probe-cli/v3/internal/netxlite/errorsx"
	"golang.org/x/net/idna"
)

// Resolver performs domain name resolutions.
type Resolver interface {
	// LookupHost behaves like net.Resolver.LookupHost.
	LookupHost(ctx context.Context, hostname string) (addrs []string, err error)

	// Network returns the resolver type (e.g., system, dot, doh).
	Network() string

	// Address returns the resolver address (e.g., 8.8.8.8:53).
	Address() string

	// CloseIdleConnections closes idle connections, if any.
	CloseIdleConnections()
}

// NewResolverSystem creates a new resolver using system
// facilities for resolving domain names (e.g., getaddrinfo).
func NewResolverSystem(logger Logger) Resolver {
	return &resolverIDNA{
		Resolver: &resolverLogger{
			Resolver: &resolverShortCircuitIPAddr{
				Resolver: &resolverErrWrapper{
					Resolver: &resolverSystem{},
				},
			},
			Logger: logger,
		},
	}
}

// resolverSystem is the system resolver.
type resolverSystem struct {
	testableTimeout    time.Duration
	testableLookupHost func(ctx context.Context, domain string) ([]string, error)
}

var _ Resolver = &resolverSystem{}

// LookupHost implements Resolver.LookupHost.
func (r *resolverSystem) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	// This code forces adding a shorter timeout to the domain name
	// resolutions when using the system resolver. We have seen cases
	// in which such a timeout becomes too large. One such case is
	// described in https://github.com/ooni/probe/issues/1726.
	addrsch, errch := make(chan []string, 1), make(chan error, 1)
	ctx, cancel := context.WithTimeout(ctx, r.timeout())
	defer cancel()
	go func() {
		addrs, err := r.lookupHost()(ctx, hostname)
		if err != nil {
			errch <- err
			return
		}
		addrsch <- addrs
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case addrs := <-addrsch:
		return addrs, nil
	case err := <-errch:
		return nil, err
	}
}

func (r *resolverSystem) timeout() time.Duration {
	if r.testableTimeout > 0 {
		return r.testableTimeout
	}
	return 15 * time.Second
}

func (r *resolverSystem) lookupHost() func(ctx context.Context, domain string) ([]string, error) {
	if r.testableLookupHost != nil {
		return r.testableLookupHost
	}
	return net.DefaultResolver.LookupHost
}

// Network implements Resolver.Network.
func (r *resolverSystem) Network() string {
	return "system"
}

// Address implements Resolver.Address.
func (r *resolverSystem) Address() string {
	return ""
}

// CloseIdleConnections implements Resolver.CloseIdleConnections.
func (r *resolverSystem) CloseIdleConnections() {
	// nothing
}

// DefaultResolver is the resolver we use by default.
var DefaultResolver = &resolverSystem{}

// resolverLogger is a resolver that emits events
type resolverLogger struct {
	Resolver
	Logger Logger
}

var _ Resolver = &resolverLogger{}

// LookupHost returns the IP addresses of a host
func (r *resolverLogger) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	r.Logger.Debugf("resolve %s...", hostname)
	start := time.Now()
	addrs, err := r.Resolver.LookupHost(ctx, hostname)
	elapsed := time.Since(start)
	if err != nil {
		r.Logger.Debugf("resolve %s... %s in %s", hostname, err, elapsed)
		return nil, err
	}
	r.Logger.Debugf("resolve %s... %+v in %s", hostname, addrs, elapsed)
	return addrs, nil
}

// resolverIDNA supports resolving Internationalized Domain Names.
//
// See RFC3492 for more information.
type resolverIDNA struct {
	Resolver
}

// LookupHost implements Resolver.LookupHost.
func (r *resolverIDNA) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	host, err := idna.ToASCII(hostname)
	if err != nil {
		return nil, err
	}
	return r.Resolver.LookupHost(ctx, host)
}

// resolverShortCircuitIPAddr recognizes when the input hostname is an
// IP address and returns it immediately to the caller.
type resolverShortCircuitIPAddr struct {
	Resolver
}

// LookupHost implements Resolver.LookupHost.
func (r *resolverShortCircuitIPAddr) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	if net.ParseIP(hostname) != nil {
		return []string{hostname}, nil
	}
	return r.Resolver.LookupHost(ctx, hostname)
}

// ErrNoResolver indicates you are using a dialer without a resolver.
var ErrNoResolver = errors.New("no configured resolver")

// nullResolver is a resolver that is not capable of resolving
// domain names to IP addresses and always returns ErrNoResolver.
type nullResolver struct{}

// LookupHost implements Resolver.LookupHost.
func (r *nullResolver) LookupHost(ctx context.Context, hostname string) (addrs []string, err error) {
	return nil, ErrNoResolver
}

// Network implements Resolver.Network.
func (r *nullResolver) Network() string {
	return "null"
}

// Address implements Resolver.Address.
func (r *nullResolver) Address() string {
	return ""
}

// CloseIdleConnections implements Resolver.CloseIdleConnections.
func (r *nullResolver) CloseIdleConnections() {
	// nothing
}

// resolverErrWrapper is a Resolver that knows about wrapping errors.
type resolverErrWrapper struct {
	Resolver
}

var _ Resolver = &resolverErrWrapper{}

// LookupHost implements Resolver.LookupHost
func (r *resolverErrWrapper) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	addrs, err := r.Resolver.LookupHost(ctx, hostname)
	if err != nil {
		return nil, &errorsx.ErrWrapper{
			Failure:    errorsx.ClassifyResolverError(err),
			Operation:  errorsx.ResolveOperation,
			WrappedErr: err,
		}
	}
	return addrs, nil
}