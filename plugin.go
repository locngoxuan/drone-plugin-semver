package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	actionRelease = "release"
	actionPatch   = "patch"
)

type Version struct {
	Major         int
	Minor         int
	Patch         int
	PreRelease    string
	BuildMetadata string
	PreBuildMeta  string
}

func (v Version) currentRelease() string {
	return fmt.Sprintf("%d.%d.0", v.Major, v.Minor)
}

func (v Version) nextRelease() string {
	return fmt.Sprintf("%d.%d.0", v.Major, v.Minor+1)
}

func (v Version) currentPatch() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v Version) nextPatch() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch+1)
}

func (v Version) devVersion() string {
	return fmt.Sprintf("%d.%d.%d-%s%s%s", v.Major, v.Minor, v.Patch, v.PreRelease, v.PreBuildMeta, v.BuildMetadata)
}

type Config struct {
	Src              string
	Output           []string
	Action           string
	PreBuildMetadata string
	DroneBuildNumber string
	RequireAction    bool
}

type Plugin struct {
	Config Config
}

const (
	extractVersion = `^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)?$`
	semverPattern  = `^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
)

var extractVerReg = regexp.MustCompile(extractVersion)
var semverReg = regexp.MustCompile(semverPattern)

func readVersionFile(file string) (map[string]string, error) {
	f, e := os.Open(file)
	if e != nil {
		return nil, e
	}
	defer func() {
		_ = f.Close()
	}()
	result := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Replace(line, "\n", "", -1)
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}
		result[parts[0]] = strings.TrimSpace(strings.Join(parts[1:], " "))
	}
	return result, nil
}

func toVersion(numbers map[string]string, prerelease, buildmetadata, buildNumber, premeta string) (v Version, err error) {
	v.PreBuildMeta = premeta
	v.Major, err = strconv.Atoi(numbers["major"])
	if err != nil {
		err = fmt.Errorf("failed to read major %v", err)
		return
	}
	v.Minor, err = strconv.Atoi(numbers["minor"])
	if err != nil {
		err = fmt.Errorf("failed to read minor %v", err)
		return
	}
	v.Patch, err = strconv.Atoi(numbers["patch"])
	if err != nil {
		err = fmt.Errorf("failed to read patch %v", err)
		return
	}
	v.PreRelease = prerelease
	if strings.TrimSpace(v.PreRelease) == "" {
		v.PreRelease = "devel"
	}

	if strings.TrimSpace(buildNumber) == "" {
		buildNumber = time.Now().Format("20060102150406")
	}

	if strings.TrimSpace(buildmetadata) == "" {
		v.BuildMetadata = buildNumber
	} else {
		v.BuildMetadata = fmt.Sprintf("%s%s%s", v.BuildMetadata, v.PreBuildMeta, buildNumber)
	}
	return
}

func (p Plugin) Exec() error {
	m, err := readVersionFile(p.Config.Src)
	if err != nil {
		return err
	}
	version := m["version"]
	prerelease := m["prerelease"]
	buildmetadata := m["buildmetadata"]
	if !semverReg.MatchString(fmt.Sprintf("%s-%s-%s", version, prerelease, buildmetadata)) {
		fmt.Println("=== VERSION ===")
		fmt.Println("src:", p.Config.Src)
		fmt.Println("version:", version)
		fmt.Println("prerelease:", prerelease)
		fmt.Println("buildmetadata:", buildmetadata)
		return fmt.Errorf(`version %s is wrong. please see https://semver.org/`, fmt.Sprintf("%s-%s-%s", version, prerelease, buildmetadata))
	}

	verNumbers := make(map[string]string)
	match := extractVerReg.FindStringSubmatch(version)
	for i, name := range extractVerReg.SubexpNames() {
		if i > 0 && name != "" {
			if i > len(match) {
				verNumbers[name] = ""
			} else {
				verNumbers[name] = match[i]
			}
		}
	}

	v, err := toVersion(verNumbers, prerelease, buildmetadata, p.Config.DroneBuildNumber, p.Config.PreBuildMetadata)
	if err != nil {
		return err
	}
	switch p.Config.Action {
	case actionRelease:
		err = writeOutput(v.currentRelease(), v.nextRelease(), p.Config.Output)
		if err != nil {
			return err
		}
	case actionPatch:
		err = writeOutput(v.currentPatch(), v.nextPatch(), p.Config.Output)
		if err != nil {
			return err
		}
	default:
		err = writeOutput(v.devVersion(), v.devVersion(), p.Config.Output)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeOutput(cur, next string, files []string) error {
	if !semverReg.Match([]byte(cur)) {
		return fmt.Errorf(`version %s is wrong. please see https://semver.org/`, cur)
	}
	for _, file := range files {
		f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		defer func() {
			_ = f.Close()
		}()
		_, err = f.WriteString(cur)
		if err != nil {
			return err
		}
		_, err = f.WriteString("\n")
		if err != nil {
			return err
		}
		_, err = f.WriteString(next)
		if err != nil {
			return err
		}
	}
	return nil
}
