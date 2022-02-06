package main

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/knqyf263/go-deb-version"
	"github.com/stapelberg/godebiancontrol"
	"github.com/ulikunitz/xz"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	packageNames string
	url          string
	suite        string
	component    string
	architecture string
)

func FatalErr(err error, str string) {
	if err != nil {
		log.Fatalf("%s: %s", str, err.Error())
	}
}

func CheckErr(err error, str string) {
	if err != nil {
		log.Printf("%s: %s", str, err.Error())
	}
}

func DeferClose(f io.Closer, str string) {
	CheckErr(f.Close(), str)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

type packageVersion struct {
	version.Version
}

func (p *packageVersion) MarshalJSON() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p packageVersion) MarshalText() (text []byte, err error) {
	return []byte(p.String()), nil
}

func DownloadPackageList() (resp *http.Response, err error) {
	for _, filename := range []string{"Packages", "Packages.gz", "Packages.xz"} {
		resp, err = http.Get(fmt.Sprintf("%s/dists/%s/%s/binary-%s/%s", url, suite, component, architecture, filename))
		if err != nil {
			return resp, err
		}
		if resp.StatusCode == 200 {
			return resp, err
		}
	}
	return resp, errors.New(resp.Status)
}

func main() {
	flag.StringVar(&packageNames, "name", "", "comma separated list of package name to check")
	flag.StringVar(&url, "url", "", "the base of the Debian distribution")
	flag.StringVar(&suite, "suite", "", "the distribution is generally a suite name")
	flag.StringVar(&component, "component", "main", "the component name")
	flag.StringVar(&architecture, "architecture", "amd64", "package architecture")
	flag.Parse()
	log.SetOutput(os.Stderr)

	if len(packageNames) == 0 || len(url) == 0 || len(suite) == 0 || len(component) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	resp, err := DownloadPackageList()
	FatalErr(err, "Failed to download package list")
	defer DeferClose(resp.Body, "Failed to close http response body")

	// Check that the server actually sent compressed data
	var reader io.Reader
	switch resp.Header.Get("Content-Type") {
	case "application/x-xz":
		xzr, err := xz.NewReader(resp.Body)
		FatalErr(err, "Failed to unpack package list")
		//defer DeferClose(reader, "Failed to close gzip stream")
		reader = io.Reader(xzr)
	case "application/x-gzip":
		gzr, err := gzip.NewReader(resp.Body)
		FatalErr(err, "Failed to unpack package list")
		defer DeferClose(gzr, "Failed to close gzip stream")
		reader = io.Reader(gzr)
	default:
		reader = resp.Body
	}
	paragraphs, err := godebiancontrol.Parse(reader)
	FatalErr(err, "Failed to parse package list")

	packages := strings.Split(packageNames, ",")
	versions := make(map[string]packageVersion)
	for _, pkg := range paragraphs {
		packageName := pkg["Package"]
		if contains(packages, packageName) {
			current, err := version.NewVersion(pkg["Version"])
			FatalErr(err, "Failed to parse package version")

			latest, ok := versions[packageName]
			if ok {
				if current.GreaterThan(latest.Version) {
					versions[packageName] = packageVersion{Version: current}
				}
			} else {
				versions[packageName] = packageVersion{Version: current}
			}
		}
	}
	d, err := json.Marshal(versions)
	FatalErr(err, "Failed to marshal data")
	fmt.Print(string(d))
}
