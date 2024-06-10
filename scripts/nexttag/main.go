package main

import (
	"bufio"
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	out, err := exec.Command("git", "tag").CombinedOutput()
	if err != nil {
		panic(err)
	}

	r := bytes.NewReader(bytes.TrimSpace(out))
	sc := bufio.NewScanner(r)

	highest := semver{}
	for sc.Scan() {
		s, err := parseSemver(sc.Text())
		if err != nil {
			fmt.Println("malformed semver: ", err)
			continue
		}

		if s.cmp(highest) > 0 {
			highest = s
		}
	}

	highest.patch++
	fmt.Println(highest)
}

type semver struct {
	major int
	minor int
	patch int
}

func (s semver) String() string {
	return fmt.Sprintf("v%d.%d.%d", s.major, s.minor, s.patch)
}

func (s semver) cmp(other semver) int {
	if s.major != other.major {
		return cmp.Compare(s.major, other.major)
	}

	if s.minor != other.minor {
		return cmp.Compare(s.minor, other.minor)
	}

	return cmp.Compare(s.patch, other.patch)
}

func parseSemver(s string) (semver, error) {
	if !strings.HasPrefix(s, "v") {
		return semver{}, errors.New("semver should start with 'v'")
	}

	parts := strings.Split(strings.TrimPrefix(s, "v"), ".")
	if len(parts) != 3 {
		return semver{}, errors.New("semver should have three parts")
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return semver{}, err
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return semver{}, err
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return semver{}, err
	}

	return semver{
		major: major,
		minor: minor,
		patch: patch,
	}, nil
}

func compareSemver(a, b string) {
}
