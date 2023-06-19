package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"google.golang.org/grpc/credentials"
)

func ServerTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed client's certificate
	pemClientCA, err := ioutil.ReadFile("./cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair("./cert/server-cert.pem", "./cert/server-key.pem")
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}

func ClientTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	// pemServerCA, err := ioutil.ReadFile("./cert/ca-cert.pem")
	// if err != nil {
	// 	return nil, err
	// }

	// certPool := x509.NewCertPool()
	// if !certPool.AppendCertsFromPEM(pemServerCA) {
	// 	return nil, fmt.Errorf("failed to add server CA's certificate")
	// }

	cert, _ := credentials.NewClientTLSFromFile("./cert/client-cert.pem", "")

	return cert, nil

	// // Load client's certificate and private key
	// clientCert, err := tls.LoadX509KeyPair("./cert/client-cert.pem", "./cert/client-key.pem")
	// if err != nil {
	// 	return nil, err
	// }

	// // Create the credentials and return it
	// config := &tls.Config{
	// 	Certificates: []tls.Certificate{clientCert},
	// 	RootCAs:      certPool,
	// }

	// return credentials.NewTLS(config), nil
}
