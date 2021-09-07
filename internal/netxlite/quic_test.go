package netxlite

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/google/go-cmp/cmp"
	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-cli/v3/internal/netxlite/errorsx"
	"github.com/ooni/probe-cli/v3/internal/netxlite/mocks"
	"github.com/ooni/probe-cli/v3/internal/netxlite/quicx"
)

func TestQUICDialerQUICGoCannotSplitHostPort(t *testing.T) {
	tlsConfig := &tls.Config{
		ServerName: "www.google.com",
	}
	systemdialer := quicDialerQUICGo{
		QUICListener: &quicListenerStdlib{},
	}
	defer systemdialer.CloseIdleConnections() // just to see it running
	ctx := context.Background()
	sess, err := systemdialer.DialContext(
		ctx, "udp", "a.b.c.d", tlsConfig, &quic.Config{})
	if err == nil || !strings.HasSuffix(err.Error(), "missing port in address") {
		t.Fatal("not the error we expected", err)
	}
	if sess != nil {
		t.Fatal("expected nil sess here")
	}
}

func TestQUICDialerQUICGoInvalidPort(t *testing.T) {
	tlsConfig := &tls.Config{
		ServerName: "www.google.com",
	}
	systemdialer := quicDialerQUICGo{
		QUICListener: &quicListenerStdlib{},
	}
	ctx := context.Background()
	sess, err := systemdialer.DialContext(
		ctx, "udp", "8.8.4.4:xyz", tlsConfig, &quic.Config{})
	if err == nil || !strings.HasSuffix(err.Error(), "invalid syntax") {
		t.Fatal("not the error we expected", err)
	}
	if sess != nil {
		t.Fatal("expected nil sess here")
	}
}

func TestQUICDialerQUICGoInvalidIP(t *testing.T) {
	tlsConfig := &tls.Config{
		ServerName: "www.google.com",
	}
	systemdialer := quicDialerQUICGo{
		QUICListener: &quicListenerStdlib{},
	}
	ctx := context.Background()
	sess, err := systemdialer.DialContext(
		ctx, "udp", "a.b.c.d:0", tlsConfig, &quic.Config{})
	if !errors.Is(err, errInvalidIP) {
		t.Fatal("not the error we expected", err)
	}
	if sess != nil {
		t.Fatal("expected nil sess here")
	}
}

func TestQUICDialerQUICGoCannotListen(t *testing.T) {
	expected := errors.New("mocked error")
	tlsConfig := &tls.Config{
		ServerName: "www.google.com",
	}
	systemdialer := quicDialerQUICGo{
		QUICListener: &mocks.QUICListener{
			MockListen: func(addr *net.UDPAddr) (quicx.UDPLikeConn, error) {
				return nil, expected
			},
		},
	}
	ctx := context.Background()
	sess, err := systemdialer.DialContext(
		ctx, "udp", "8.8.8.8:443", tlsConfig, &quic.Config{})
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected", err)
	}
	if sess != nil {
		t.Fatal("expected nil sess here")
	}
}

func TestQUICDialerQUICGoCannotPerformHandshake(t *testing.T) {
	tlsConfig := &tls.Config{
		ServerName: "dns.google",
	}
	systemdialer := quicDialerQUICGo{
		QUICListener: &quicListenerStdlib{},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // fail immediately
	sess, err := systemdialer.DialContext(
		ctx, "udp", "8.8.8.8:443", tlsConfig, &quic.Config{})
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected", err)
	}
	if sess != nil {
		log.Fatal("expected nil session here")
	}
}

func TestQUICDialerQUICGoWorksAsIntended(t *testing.T) {
	tlsConfig := &tls.Config{
		ServerName: "dns.google",
	}
	systemdialer := quicDialerQUICGo{
		QUICListener: &quicListenerStdlib{},
	}
	ctx := context.Background()
	sess, err := systemdialer.DialContext(
		ctx, "udp", "8.8.8.8:443", tlsConfig, &quic.Config{})
	if err != nil {
		t.Fatal("not the error we expected", err)
	}
	<-sess.HandshakeComplete().Done()
	if err := sess.CloseWithError(0, ""); err != nil {
		t.Fatal(err)
	}
}

func TestQUICDialerQUICGoTLSDefaultsForWeb(t *testing.T) {
	expected := errors.New("mocked error")
	var gotTLSConfig *tls.Config
	tlsConfig := &tls.Config{
		ServerName: "dns.google",
	}
	systemdialer := quicDialerQUICGo{
		QUICListener: &quicListenerStdlib{},
		mockDialEarlyContext: func(ctx context.Context, pconn net.PacketConn,
			remoteAddr net.Addr, host string, tlsConfig *tls.Config,
			quicConfig *quic.Config) (quic.EarlySession, error) {
			gotTLSConfig = tlsConfig
			return nil, expected
		},
	}
	ctx := context.Background()
	sess, err := systemdialer.DialContext(
		ctx, "udp", "8.8.8.8:443", tlsConfig, &quic.Config{})
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected", err)
	}
	if sess != nil {
		t.Fatal("expected nil session here")
	}
	if tlsConfig.RootCAs != nil {
		t.Fatal("tlsConfig.RootCAs should not have been changed")
	}
	if gotTLSConfig.RootCAs != defaultCertPool {
		t.Fatal("invalid gotTLSConfig.RootCAs")
	}
	if tlsConfig.NextProtos != nil {
		t.Fatal("tlsConfig.NextProtos should not have been changed")
	}
	if diff := cmp.Diff(gotTLSConfig.NextProtos, []string{"h3"}); diff != "" {
		t.Fatal("invalid gotTLSConfig.NextProtos", diff)
	}
	if tlsConfig.ServerName != gotTLSConfig.ServerName {
		t.Fatal("the ServerName field must match")
	}
}

func TestQUICDialerQUICGoTLSDefaultsForDoQ(t *testing.T) {
	expected := errors.New("mocked error")
	var gotTLSConfig *tls.Config
	tlsConfig := &tls.Config{
		ServerName: "dns.google",
	}
	systemdialer := quicDialerQUICGo{
		QUICListener: &quicListenerStdlib{},
		mockDialEarlyContext: func(ctx context.Context, pconn net.PacketConn,
			remoteAddr net.Addr, host string, tlsConfig *tls.Config,
			quicConfig *quic.Config) (quic.EarlySession, error) {
			gotTLSConfig = tlsConfig
			return nil, expected
		},
	}
	ctx := context.Background()
	sess, err := systemdialer.DialContext(
		ctx, "udp", "8.8.8.8:8853", tlsConfig, &quic.Config{})
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected", err)
	}
	if sess != nil {
		t.Fatal("expected nil session here")
	}
	if tlsConfig.RootCAs != nil {
		t.Fatal("tlsConfig.RootCAs should not have been changed")
	}
	if gotTLSConfig.RootCAs != defaultCertPool {
		t.Fatal("invalid gotTLSConfig.RootCAs")
	}
	if tlsConfig.NextProtos != nil {
		t.Fatal("tlsConfig.NextProtos should not have been changed")
	}
	if diff := cmp.Diff(gotTLSConfig.NextProtos, []string{"dq"}); diff != "" {
		t.Fatal("invalid gotTLSConfig.NextProtos", diff)
	}
	if tlsConfig.ServerName != gotTLSConfig.ServerName {
		t.Fatal("the ServerName field must match")
	}
}

func TestQUICDialerResolverCloseIdleConnections(t *testing.T) {
	var (
		forDialer   bool
		forResolver bool
	)
	d := &quicDialerResolver{
		Dialer: &mocks.QUICDialer{
			MockCloseIdleConnections: func() {
				forDialer = true
			},
		},
		Resolver: &mocks.Resolver{
			MockCloseIdleConnections: func() {
				forResolver = true
			},
		},
	}
	d.CloseIdleConnections()
	if !forDialer || !forResolver {
		t.Fatal("not called")
	}
}

func TestQUICDialerResolverSuccess(t *testing.T) {
	tlsConfig := &tls.Config{}
	dialer := &quicDialerResolver{
		Resolver: NewResolverSystem(log.Log),
		Dialer: &quicDialerQUICGo{
			QUICListener: &quicListenerStdlib{},
		}}
	sess, err := dialer.DialContext(
		context.Background(), "udp", "www.google.com:443",
		tlsConfig, &quic.Config{})
	if err != nil {
		t.Fatal(err)
	}
	<-sess.HandshakeComplete().Done()
	if err := sess.CloseWithError(0, ""); err != nil {
		t.Fatal(err)
	}
}

func TestQUICDialerResolverNoPort(t *testing.T) {
	tlsConfig := &tls.Config{}
	dialer := &quicDialerResolver{
		Resolver: NewResolverSystem(log.Log),
		Dialer:   &quicDialerQUICGo{}}
	sess, err := dialer.DialContext(
		context.Background(), "udp", "www.google.com",
		tlsConfig, &quic.Config{})
	if err == nil || !strings.HasSuffix(err.Error(), "missing port in address") {
		t.Fatal("not the error we expected")
	}
	if sess != nil {
		t.Fatal("expected a nil sess here")
	}
}

func TestQUICDialerResolverLookupHostAddress(t *testing.T) {
	dialer := &quicDialerResolver{Resolver: &mocks.Resolver{
		MockLookupHost: func(ctx context.Context, domain string) ([]string, error) {
			// We should not arrive here and call this function but if we do then
			// there is going to be an error that fails this test.
			return nil, errors.New("mocked error")
		},
	}}
	addrs, err := dialer.lookupHost(context.Background(), "1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) != 1 || addrs[0] != "1.1.1.1" {
		t.Fatal("not the result we expected")
	}
}

func TestQUICDialerResolverLookupHostFailure(t *testing.T) {
	tlsConfig := &tls.Config{}
	expected := errors.New("mocked error")
	dialer := &quicDialerResolver{Resolver: &mocks.Resolver{
		MockLookupHost: func(ctx context.Context, domain string) ([]string, error) {
			return nil, expected
		},
	}}
	sess, err := dialer.DialContext(
		context.Background(), "udp", "dns.google.com:853",
		tlsConfig, &quic.Config{})
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if sess != nil {
		t.Fatal("expected nil sess")
	}
}

func TestQUICDialerResolverInvalidPort(t *testing.T) {
	// This test allows us to check for the case where every attempt
	// to establish a connection leads to a failure
	tlsConf := &tls.Config{}
	dialer := &quicDialerResolver{
		Resolver: NewResolverSystem(log.Log),
		Dialer: &quicDialerQUICGo{
			QUICListener: &quicListenerStdlib{},
		}}
	sess, err := dialer.DialContext(
		context.Background(), "udp", "www.google.com:0",
		tlsConf, &quic.Config{})
	if err == nil {
		t.Fatal("expected an error here")
	}
	if !strings.HasSuffix(err.Error(), "sendto: invalid argument") &&
		!strings.HasSuffix(err.Error(), "sendto: can't assign requested address") {
		t.Fatal("not the error we expected", err)
	}
	if sess != nil {
		t.Fatal("expected nil sess")
	}
}

func TestQUICDialerResolverApplyTLSDefaults(t *testing.T) {
	expected := errors.New("mocked error")
	var gotTLSConfig *tls.Config
	tlsConfig := &tls.Config{}
	dialer := &quicDialerResolver{
		Resolver: NewResolverSystem(log.Log),
		Dialer: &mocks.QUICDialer{
			MockDialContext: func(ctx context.Context, network, address string,
				tlsConfig *tls.Config, quicConfig *quic.Config) (quic.EarlySession, error) {
				gotTLSConfig = tlsConfig
				return nil, expected
			},
		}}
	sess, err := dialer.DialContext(
		context.Background(), "udp", "www.google.com:443",
		tlsConfig, &quic.Config{})
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected", err)
	}
	if sess != nil {
		t.Fatal("expected nil session here")
	}
	if tlsConfig.ServerName != "" {
		t.Fatal("should not have changed tlsConfig.ServerName")
	}
	if gotTLSConfig.ServerName != "www.google.com" {
		t.Fatal("gotTLSConfig.ServerName has not been set")
	}
}

func TestQUICDialerLoggerCloseIdleConnections(t *testing.T) {
	var forDialer bool
	d := &quicDialerLogger{
		Dialer: &mocks.QUICDialer{
			MockCloseIdleConnections: func() {
				forDialer = true
			},
		},
	}
	d.CloseIdleConnections()
	if !forDialer {
		t.Fatal("not called")
	}
}

func TestQUICDialerLoggerSuccess(t *testing.T) {
	d := &quicDialerLogger{
		Dialer: &mocks.QUICDialer{
			MockDialContext: func(ctx context.Context, network string,
				address string, tlsConfig *tls.Config,
				quicConfig *quic.Config) (quic.EarlySession, error) {
				return &mocks.QUICEarlySession{
					MockCloseWithError: func(
						code quic.ApplicationErrorCode, reason string) error {
						return nil
					},
				}, nil
			},
		},
		Logger: log.Log,
	}
	ctx := context.Background()
	tlsConfig := &tls.Config{}
	quicConfig := &quic.Config{}
	sess, err := d.DialContext(ctx, "udp", "8.8.8.8:443", tlsConfig, quicConfig)
	if err != nil {
		t.Fatal(err)
	}
	if err := sess.CloseWithError(0, ""); err != nil {
		t.Fatal(err)
	}
}

func TestQUICDialerLoggerFailure(t *testing.T) {
	expected := errors.New("mocked error")
	d := &quicDialerLogger{
		Dialer: &mocks.QUICDialer{
			MockDialContext: func(ctx context.Context, network string,
				address string, tlsConfig *tls.Config,
				quicConfig *quic.Config) (quic.EarlySession, error) {
				return nil, expected
			},
		},
		Logger: log.Log,
	}
	ctx := context.Background()
	tlsConfig := &tls.Config{}
	quicConfig := &quic.Config{}
	sess, err := d.DialContext(ctx, "udp", "8.8.8.8:443", tlsConfig, quicConfig)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected", err)
	}
	if sess != nil {
		t.Fatal("expected nil session")
	}
}

func TestNewQUICDialerWithoutResolverChain(t *testing.T) {
	ql := NewQUICListener()
	dlr := NewQUICDialerWithoutResolver(ql, log.Log)
	dlog, okay := dlr.(*quicDialerLogger)
	if !okay {
		t.Fatal("invalid type")
	}
	if dlog.Logger != log.Log {
		t.Fatal("invalid logger")
	}
	dr, okay := dlog.Dialer.(*quicDialerResolver)
	if !okay {
		t.Fatal("invalid type")
	}
	if _, okay := dr.Resolver.(*nullResolver); !okay {
		t.Fatal("invalid resolver type")
	}
	dlog, okay = dr.Dialer.(*quicDialerLogger)
	if !okay {
		t.Fatal("invalid type")
	}
	if dlog.Logger != log.Log {
		t.Fatal("invalid logger")
	}
	ew, okay := dlog.Dialer.(*quicDialerErrWrapper)
	if !okay {
		t.Fatal("invalid type")
	}
	dgo, okay := ew.QUICDialer.(*quicDialerQUICGo)
	if !okay {
		t.Fatal("invalid type")
	}
	if dgo.QUICListener != ql {
		t.Fatal("invalid quic listener")
	}
}

func TestNewSingleUseQUICDialerWorksAsIntended(t *testing.T) {
	sess := &mocks.QUICEarlySession{}
	qd := NewSingleUseQUICDialer(sess)
	outsess, err := qd.DialContext(
		context.Background(), "", "", &tls.Config{}, &quic.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if sess != outsess {
		t.Fatal("invalid outsess")
	}
	for i := 0; i < 4; i++ {
		outsess, err = qd.DialContext(
			context.Background(), "", "", &tls.Config{}, &quic.Config{})
		if !errors.Is(err, ErrNoConnReuse) {
			t.Fatal("not the error we expected", err)
		}
		if outsess != nil {
			t.Fatal("expected nil outconn here")
		}
	}
}

func TestQUICListenerErrWrapper(t *testing.T) {
	t.Run("Listen", func(t *testing.T) {
		t.Run("on success", func(t *testing.T) {
			expectedConn := &mocks.QUICUDPConn{}
			ql := &quicListenerErrWrapper{
				QUICListener: &mocks.QUICListener{
					MockListen: func(addr *net.UDPAddr) (quicx.UDPLikeConn, error) {
						return expectedConn, nil
					},
				},
			}
			conn, err := ql.Listen(&net.UDPAddr{})
			if err != nil {
				t.Fatal(err)
			}
			ewconn := conn.(*quicErrWrapperUDPLikeConn)
			if ewconn.UDPLikeConn != expectedConn {
				t.Fatal("unexpected conn")
			}
		})

		t.Run("on failure", func(t *testing.T) {
			expectedErr := io.EOF
			ql := &quicListenerErrWrapper{
				QUICListener: &mocks.QUICListener{
					MockListen: func(addr *net.UDPAddr) (quicx.UDPLikeConn, error) {
						return nil, expectedErr
					},
				},
			}
			conn, err := ql.Listen(&net.UDPAddr{})
			if err == nil || err.Error() != errorsx.FailureEOFError {
				t.Fatal("unexpected err", err)
			}
			if conn != nil {
				t.Fatal("expected nil conn")
			}
		})
	})
}

func TestQUICErrWrapperUDPLikeConn(t *testing.T) {
	t.Run("ReadFrom", func(t *testing.T) {
		t.Run("on success", func(t *testing.T) {
			expectedAddr := &net.UDPAddr{}
			p := make([]byte, 128)
			conn := &quicErrWrapperUDPLikeConn{
				UDPLikeConn: &mocks.QUICUDPConn{
					MockReadFrom: func(p []byte) (n int, addr net.Addr, err error) {
						return len(p), expectedAddr, nil
					},
				},
			}
			count, addr, err := conn.ReadFrom(p)
			if err != nil {
				t.Fatal(err)
			}
			if count != len(p) {
				t.Fatal("unexpected count")
			}
			if addr != expectedAddr {
				t.Fatal("unexpected addr")
			}
		})

		t.Run("on failure", func(t *testing.T) {
			p := make([]byte, 128)
			expectedErr := io.EOF
			conn := &quicErrWrapperUDPLikeConn{
				UDPLikeConn: &mocks.QUICUDPConn{
					MockReadFrom: func(p []byte) (n int, addr net.Addr, err error) {
						return 0, nil, expectedErr
					},
				},
			}
			count, addr, err := conn.ReadFrom(p)
			if err == nil || err.Error() != errorsx.FailureEOFError {
				t.Fatal("unexpected err", err)
			}
			if count != 0 {
				t.Fatal("unexpected count")
			}
			if addr != nil {
				t.Fatal("unexpected addr")
			}
		})
	})

	t.Run("WriteTo", func(t *testing.T) {
		t.Run("on success", func(t *testing.T) {
			p := make([]byte, 128)
			conn := &quicErrWrapperUDPLikeConn{
				UDPLikeConn: &mocks.QUICUDPConn{
					MockWriteTo: func(p []byte, addr net.Addr) (int, error) {
						return len(p), nil
					},
				},
			}
			count, err := conn.WriteTo(p, &net.UDPAddr{})
			if err != nil {
				t.Fatal(err)
			}
			if count != len(p) {
				t.Fatal("unexpected count")
			}
		})

		t.Run("on failure", func(t *testing.T) {
			p := make([]byte, 128)
			expectedErr := io.EOF
			conn := &quicErrWrapperUDPLikeConn{
				UDPLikeConn: &mocks.QUICUDPConn{
					MockWriteTo: func(p []byte, addr net.Addr) (int, error) {
						return 0, expectedErr
					},
				},
			}
			count, err := conn.WriteTo(p, &net.UDPAddr{})
			if err == nil || err.Error() != errorsx.FailureEOFError {
				t.Fatal("unexpected err", err)
			}
			if count != 0 {
				t.Fatal("unexpected count")
			}
		})
	})
}

func TestQUICDialerErrWrapper(t *testing.T) {
	t.Run("CloseIdleConnections", func(t *testing.T) {
		var called bool
		d := &quicDialerErrWrapper{
			QUICDialer: &mocks.QUICDialer{
				MockCloseIdleConnections: func() {
					called = true
				},
			},
		}
		d.CloseIdleConnections()
		if !called {
			t.Fatal("not called")
		}
	})

	t.Run("DialContext", func(t *testing.T) {
		t.Run("on success", func(t *testing.T) {
			expectedSess := &mocks.QUICEarlySession{}
			d := &quicDialerErrWrapper{
				QUICDialer: &mocks.QUICDialer{
					MockDialContext: func(ctx context.Context, network, address string, tlsConfig *tls.Config, quicConfig *quic.Config) (quic.EarlySession, error) {
						return expectedSess, nil
					},
				},
			}
			ctx := context.Background()
			sess, err := d.DialContext(ctx, "", "", &tls.Config{}, &quic.Config{})
			if err != nil {
				t.Fatal(err)
			}
			if sess != expectedSess {
				t.Fatal("unexpected sess")
			}
		})

		t.Run("on failure", func(t *testing.T) {
			expectedErr := io.EOF
			d := &quicDialerErrWrapper{
				QUICDialer: &mocks.QUICDialer{
					MockDialContext: func(ctx context.Context, network, address string, tlsConfig *tls.Config, quicConfig *quic.Config) (quic.EarlySession, error) {
						return nil, expectedErr
					},
				},
			}
			ctx := context.Background()
			sess, err := d.DialContext(ctx, "", "", &tls.Config{}, &quic.Config{})
			if err == nil || err.Error() != errorsx.FailureEOFError {
				t.Fatal("unexpected err", err)
			}
			if sess != nil {
				t.Fatal("unexpected sess")
			}
		})
	})
}