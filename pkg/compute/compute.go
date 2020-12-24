package compute

import (
	"errors"
	"fmt"
	"github.com/mousybusiness/go-web/web"
	errs "github.com/pkg/errors"
	"os"
	"regexp"
	"time"
)

const serverMetaProjectID = "http://metadata.google.internal/computeMetadata/v1/project/project-id"
const serverMetaExternalIP = "http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip"
const serverMetaInternalIP = "http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/ip"
const serverMetaSubnetMask = "http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/subnetmask"
const serverMetaHostname = "http://metadata.google.internal/computeMetadata/v1/instance/hostname"
const serverMetaZone = "http://metadata.google.internal/computeMetadata/v1/instance/zone"
const serverMetaProjectNumber = "http://metadata.google.internal/computeMetadata/v1/project/numeric-project-id"

var (
	projectIDCache  string
	externalIPCache string
	internalIPCache string
	subnetMaskCache string
	hostCache       string
	zoneCache       string
)

func ClearCache() {
	projectIDCache = ""
	externalIPCache = ""
	internalIPCache = ""
	subnetMaskCache = ""
	hostCache = ""
	zoneCache = ""
}

// ensures GOOGLE_CLOUD_PROJECT is set with project id
func EnsureProjectID() {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT") // environment variable provided by app engine
	if projectID == "" {
		projectID, err := GetProjectID()
		if err != nil {
			panic(err)
		}
		os.Setenv("GOOGLE_CLOUD_PROJECT", projectID)
	}
}

// gets servers project on GCP using metadata endpoint
func GetProjectID() (string, error) {
	if os.Getenv("LOCAL_ENV") == "true" {
		return os.Getenv("GOOGLE_CLOUD_PROJECT"), nil
	}

	if projectIDCache != "" {
		return projectIDCache, nil
	}

	metadata, err := getMetadata(serverMetaProjectID)
	if err != nil {
		return "", err
	}

	match, err := regexp.MatchString(`^[a-zA-Z\d-]{8,}$`, metadata)
	if err != nil {
		return "", errs.Wrap(err, "failed to compile")
	}
	if !match {
		return "", errors.New(fmt.Sprintf("metadata response didn't match regex: %s", metadata))
	}

	projectIDCache = metadata
	return metadata, nil
}

// internal DNS works within VPC
func GetInternalDNS() (string, error) {
	if os.Getenv("LOCAL_ENV") == "true" {
		return "localhost", nil
	}

	if hostCache != "" {
		return hostCache, nil
	}

	metadata, err := getMetadata(serverMetaHostname)
	if err != nil {
		return "", err
	}

	match, err := regexp.MatchString(`^.+?\.\w+?-\w+?-\w\.c\..+?\.internal$`, metadata)
	if err != nil {
		return "", err
	}

	if !match {
		return "", errors.New(fmt.Sprintf("metadata response didn't match regex: %s", metadata))
	}

	hostCache = metadata
	return metadata, nil
}

func GetSubnetMask() (string, error) {
	if os.Getenv("LOCAL_ENV") == "true" {
		return "255.255.0.0", nil
	}

	if subnetMaskCache != "" {
		return subnetMaskCache, nil
	}

	metadata, err := getMetadata(serverMetaSubnetMask)
	if err != nil {
		return "", err
	}

	match, err := regexp.MatchString(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`, metadata)
	if err != nil {
		return "", errs.Wrap(err, "failed to compile")
	}
	if !match {
		return "", errors.New(fmt.Sprintf("metadata response didn't match regex: %s", metadata))
	}

	subnetMaskCache = metadata
	return metadata, err
}

func GetZone() (string, error) {
	if os.Getenv("LOCAL_ENV") == "true" {
		return "europe-west2-c", nil
	}

	if zoneCache != "" {
		return zoneCache, nil
	}

	metadata, err := getMetadata(serverMetaZone)
	if err != nil {
		return "", err
	}

	// returned format: projects/1016716848681/zones/europe-west2-c
	exp := regexp.MustCompile(`^projects\/\d+?\/zones\/(\w+?-\w+?-\w)$`)
	match := exp.MatchString(metadata)
	if !match {
		return "", errors.New("failed to match zone regex")
	}

	// extract only the zone
	zone := exp.ReplaceAllString(metadata, `$1`)
	zoneCache = zone
	return zone, nil
}

// gets servers external IP on GCP using metadata endpoint
func GetExternalIP() (string, error) {
	if os.Getenv("LOCAL_ENV") == "true" {
		return "127.0.0.1", nil
	}

	if externalIPCache != "" {
		return externalIPCache, nil
	}

	metadata, err := getMetadata(serverMetaExternalIP)
	if err != nil {
		return "", err
	}

	match, err := regexp.MatchString(`^\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}$`, metadata)
	if err != nil {
		return "", errs.Wrap(err, "failed to compile")
	}
	if !match {
		return "", errors.New(fmt.Sprintf("metadata response didn't match regex: %s", metadata))
	}

	externalIPCache = metadata
	return metadata, err
}

func GetInternalIP() (string, error) {
	if os.Getenv("LOCAL_ENV") == "true" {
		return "127.0.0.1", nil
	}

	if internalIPCache != "" {
		return internalIPCache, nil
	}

	metadata, err := getMetadata(serverMetaInternalIP)
	if err != nil {
		return "", err
	}

	match, err := regexp.MatchString(`^\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}$`, metadata)
	if err != nil {
		return "", errs.Wrap(err, "failed to compile")
	}
	if !match {
		return "", errors.New(fmt.Sprintf("metadata response didn't match regex: %s", metadata))
	}

	internalIPCache = metadata
	return metadata, err
}

// Google Compute servers have metadata accessable via a local HTTP endpoint
func getMetadata(url string) (string, error) {
	_, b, err := web.Get(url, time.Second*2, web.KV{"Metadata-Flavor", "Google"})
	return string(b), err
}
