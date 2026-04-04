// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.
//
// Registers a MySQL TLS config named "env" from MYSQL_SSL_CA, MYSQL_SSL_CERT,
// and MYSQL_SSL_KEY environment variables. Use tls=env in the DSN to activate.
// go-sql-driver/mysql does not support file-path-based TLS via DSN parameters,
// so environment variables are used instead.

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/go-sql-driver/mysql"
)

func init() {
	registerMySQLTLSFromEnv()
}

// registerMySQLTLSFromEnv registers an "env" MySQL TLS config if
// MYSQL_SSL_CA, MYSQL_SSL_CERT, and MYSQL_SSL_KEY env vars are all set.
func registerMySQLTLSFromEnv() {
	caPath := os.Getenv("MYSQL_SSL_CA")
	certPath := os.Getenv("MYSQL_SSL_CERT")
	keyPath := os.Getenv("MYSQL_SSL_KEY")
	if caPath == "" || certPath == "" || keyPath == "" {
		return
	}
	caPEM, err := os.ReadFile(caPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "atlas: read CA cert %s: %v\n", caPath, err)
		os.Exit(1)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(caPEM) {
		fmt.Fprintf(os.Stderr, "atlas: no valid CA certificates in %s\n", caPath)
		os.Exit(1)
	}
	if err := mysql.RegisterTLSConfig("env", &tls.Config{
		RootCAs: roots,
		GetClientCertificate: func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert, err := tls.LoadX509KeyPair(certPath, keyPath)
			if err != nil {
				return nil, fmt.Errorf("load client cert: %w", err)
			}
			return &cert, nil
		},
	}); err != nil {
		fmt.Fprintf(os.Stderr, "atlas: register TLS config: %v\n", err)
		os.Exit(1)
	}
}
