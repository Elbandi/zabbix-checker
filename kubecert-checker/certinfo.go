package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	webcertificate "golang.zabbix.com/agent2/plugins/web/certificate"
	"os"
	"path/filepath"
	"strings"
)

const (
	dateFormat = "Jan 02 15:04:05 2006 GMT"
)

type KubernetesCert struct {
	webcertificate.Output
	Name string `json:"name"`
}

var certInfoCommand = cli.Command{
	Name:   "certinfo",
	Usage:  "certinfo /var/lib/rancher/k3s",
	Action: cmdCertInfo,
}

func getValidationResult(leaf *x509.Certificate, opts x509.VerifyOptions, subject, issuer string) webcertificate.ValidationResult {
	var out webcertificate.ValidationResult

	if _, err := leaf.Verify(opts); err != nil {
		if errors.As(err, &x509.UnknownAuthorityError{}) && subject == issuer {
			out = webcertificate.ValidationResult{
				Value:   "valid-but-self-signed",
				Message: "certificate verified successfully, but determined to be self signed",
			}
		} else {
			out = webcertificate.ValidationResult{Value: "invalid", Message: fmt.Sprintf("failed to verify certificate: %s", err.Error())}
		}
	} else {
		out = webcertificate.ValidationResult{Value: "valid", Message: "certificate verified successfully"}
	}

	return out
}

func cmdCertInfo(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		return errors.New("missing rancher path")
	}
	rancherPath := ctx.Args().First()
	ltsPath := filepath.Join(rancherPath, "server", "tls")
	if stat, err := os.Stat(ltsPath); (err != nil) || !stat.IsDir() {
		return errors.New("invalid rancher path")
	}
	output := make([]KubernetesCert, 0)
	err := filepath.Walk(ltsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if ltsPath == path {
			return nil
		}
		if !strings.HasSuffix(path, ".crt") {
			return nil
		}
		fullPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		certPEM, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read certificate '%s' file: %w", fullPath, err)
		}
		block, rest := pem.Decode(certPEM)
		if block == nil {
			return fmt.Errorf("failed to decode PEM block from '%s' file", fullPath)
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse certificate: %w", err)
		}
		rootPool := x509.NewCertPool()
		if len(rest) > 0 {
			ok := rootPool.AppendCertsFromPEM(rest)
			if !ok {
				return fmt.Errorf("couldn't add pems from fullchain")
			}
		}
		var o KubernetesCert
		o.Name = info.Name()
		o.X509 = webcertificate.Cert{
			Version:            cert.Version,
			Serial:             fmt.Sprintf("%x", cert.SerialNumber.Bytes()),
			SignatureAlgorithm: cert.SignatureAlgorithm.String(),
			Issuer:             cert.Issuer.ToRDNSequence().String(),
			NotBefore:          webcertificate.CertTime{Value: cert.NotBefore.UTC().Format(dateFormat), Timestamp: cert.NotBefore.Unix()},
			NotAfter:           webcertificate.CertTime{Value: cert.NotAfter.UTC().Format(dateFormat), Timestamp: cert.NotAfter.Unix()},
			Subject:            cert.Subject.ToRDNSequence().String(),
			PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
			AlternativeNames:   cert.DNSNames,
		}
		o.Result = getValidationResult(
			cert, x509.VerifyOptions{Roots: rootPool, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny}},
			cert.Subject.ToRDNSequence().String(), cert.Issuer.ToRDNSequence().String(),
		)

		o.Sha1Fingerprint = fmt.Sprintf("%x", sha1.Sum(cert.Raw))
		o.Sha256Fingerprint = fmt.Sprintf("%x", sha256.Sum256(cert.Raw))
		output = append(output, o)
		return nil
	})
	if err != nil {
		return err
	}

	//	b, err := json.MarshalIndent(output, "", "  ")
	b, err := json.Marshal(output)
	if err != nil {
		return err
	}
	fmt.Println(string(b))

	return nil
}
