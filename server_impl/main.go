package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

func main() {
	SSLKEYLOGFILE := os.Getenv("SSLKEYLOGFILE")
	QLOGDIR := os.Getenv("QLOGDIR")
	LOGS := os.Getenv("LOGS")
	TESTCASE := os.Getenv("TESTCASE")
	WWW := os.Getenv("WWW")
	CERTS := os.Getenv("CERTS")
	IP := os.Getenv("IP")
	PORT := os.Getenv("PORT")

	lf, err := os.Create(LOGS + "/log.txt")
	if err != nil {
		panic(err)
	}
	defer lf.Close()
	log.SetOutput(lf)

	if err = os.MkdirAll(QLOGDIR, 0755); err != nil {
		log.Fatal(err)
	}
	qlf, err := os.Create(QLOGDIR + "log.qlog")
	if err != nil {
		log.Fatal(err)
	}
	defer qlf.Close()
	quicConf := &quic.Config{
		Tracer: func(ctx context.Context, lp logging.Perspective, connID quic.ConnectionID) *logging.ConnectionTracer {
			return qlog.NewConnectionTracer(NewBufferedWriteCloser(bufio.NewWriter(qlf), qlf), lp, connID)
		},
		RequireAddressValidation: func(net.Addr) bool { return true },
		Versions:                 []quic.VersionNumber{quic.Version1},
	}

	keyLogWriter, err := os.Create(SSLKEYLOGFILE)
	if err != nil {
		log.Fatal(err)
	}
	defer keyLogWriter.Close()

	certFile := CERTS + "/cert.pem"
	keyFile := CERTS + "/priv.key"
	//cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	//if err != nil {
	//	log.Fatal(err)
	//}

	tlsConf := &tls.Config{
		//Certificates: []tls.Certificate{cert},
		KeyLogWriter: keyLogWriter,
	}

	switch TESTCASE {
	case "handshake", "retry", "transfer", "multihandshake":
		handler := http.FileServer(http.Dir(WWW))
		server := http3.Server{
			Handler:    handler,
			Addr:       IP + ":" + PORT,
			QuicConfig: quicConf,
			TLSConfig:  tlsConf,
		}
		if err = server.ListenAndServeTLS(certFile, keyFile); err != nil {
			return
		}
	default:
		fmt.Println("exited with code 127")
		os.Exit(127)
	}
}

type bufferedWriteCloser struct {
	*bufio.Writer
	io.Closer
}

func NewBufferedWriteCloser(writer *bufio.Writer, closer io.Closer) io.WriteCloser {
	return &bufferedWriteCloser{Writer: writer, Closer: closer}
}

func (b bufferedWriteCloser) Close() error {
	if err := b.Writer.Flush(); err != nil {
		return err
	}
	return b.Closer.Close()
}
