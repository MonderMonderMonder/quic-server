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
	"net/http"
	"os"
	"strings"
	"sync"
)

func main() {
	SSLKEYLOGFILE := os.Getenv("SSLKEYLOGFILE")
	QLOGDIR := os.Getenv("QLOGDIR")
	LOGS := os.Getenv("LOGS")
	TESTCASE := os.Getenv("TESTCASE")
	DOWNLOADS := os.Getenv("DOWNLOADS")
	REQUESTS := strings.Split(os.Getenv("REQUESTS"), " ")

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
		Versions: []quic.VersionNumber{quic.Version1},
	}

	keyLogWriter, err := os.Create(SSLKEYLOGFILE)
	if err != nil {
		log.Fatal(err)
	}
	defer keyLogWriter.Close()

	//pool, err := x509.SystemCertPool()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//AddRootCA(pool)

	tlsConf := &tls.Config{
		//RootCAs:            pool,
		InsecureSkipVerify: true,
		KeyLogWriter:       keyLogWriter,
	}

	switch TESTCASE {
	case "handshake", "retry", "transfer":
		roundTripper := &http3.RoundTripper{QuicConfig: quicConf, TLSClientConfig: tlsConf}
		defer roundTripper.Close()
		client := &http.Client{Transport: roundTripper}
		var wg sync.WaitGroup
		wg.Add(len(REQUESTS))
		for _, request := range REQUESTS {
			go func(request string) {
				resp, err := client.Get(request)
				if err != nil {
					log.Fatal(err)
				}
				defer resp.Body.Close()
				decomp := strings.Split(request, "/")
				file, err := os.Create(DOWNLOADS + "/" + decomp[len(decomp)-1])
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()
				io.Copy(file, resp.Body)
				wg.Done()
			}(request)
		}
		wg.Wait()
	case "multihandshake":
		var wg sync.WaitGroup
		wg.Add(len(REQUESTS))
		for _, request := range REQUESTS {
			go func(request string) {
				roundTripper := &http3.RoundTripper{QuicConfig: quicConf, TLSClientConfig: tlsConf}
				client := &http.Client{Transport: roundTripper}
				resp, err := client.Get(request)
				if err != nil {
					log.Fatal(err)
				}
				defer resp.Body.Close()
				decomp := strings.Split(request, "/")
				file, err := os.Create(DOWNLOADS + "/" + decomp[len(decomp)-1])
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()
				io.Copy(file, resp.Body)
				roundTripper.Close()
				wg.Done()
			}(request)
		}
	default:
		fmt.Println("exited with code 127")
		os.Exit(127)
	}
}

//func AddRootCA(certPool *x509.CertPool) {
//	_, callerFile, _, ok := runtime.Caller(0)
//	if !ok {
//		log.Fatal("Failed to get current frame in AddRootCA")
//	}
//	certPath := path.Dir(callerFile)
//	caCertPath := path.Join(certPath, "ca.pem")
//	caCertRaw, err := os.ReadFile(caCertPath)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if ok := certPool.AppendCertsFromPEM(caCertRaw); !ok {
//		log.Fatal("Could not add root ceritificate to pool.")
//	}
//}

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
