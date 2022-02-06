package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/knqyf263/go-deb-version"
	"github.com/stapelberg/godebiancontrol"
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

	resp, err := http.Get(fmt.Sprintf("%s/dists/%s/%s/binary-%s/Packages", url, suite, component, architecture))
	FatalErr(err, "Failed to download package list")
	defer DeferClose(resp.Body, "Failed to close http response body")

	paragraphs, err := godebiancontrol.Parse(resp.Body)
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
