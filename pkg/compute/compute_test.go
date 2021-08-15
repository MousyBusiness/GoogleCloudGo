package instances

import (
	"github.com/mousybusiness/go-web/web"
	"github.com/mousybusiness/go-web/web/webtest"
	"net/http"
	"os"
	"testing"
)

func init() {
	web.Client = webtest.MockClient{}
}

func TestClearCache(t *testing.T) {
	s := []*string{&externalIPCache, &internalIPCache, &subnetMaskCache, &hostCache, &zoneCache}
	for _, v := range s {
		(*v) = "stub"
	}

	ClearCache()

	for _, v := range s {
		if (*v) == "stub" {
			t.Fatalf("clearing cache failed")
		}
	}
}

func TestEnsureProjectID(t *testing.T) {
	resp := "test-project"
	webtest.DoFunc = func(req *http.Request) (*http.Response, error) {
		return webtest.MockResponse(200, resp, nil)
	}

	EnsureProjectID()
	if os.Getenv("GOOGLE_CLOUD_PROJECT") == "" {
		t.Fatalf("setting project id to GOOGLE_CLOUD_PROJECT env variable failed")
	}
}

func TestAllGets(t *testing.T) {
	os.Setenv("GOOGLE_CLOUD_PROJECT", "local-test-project")
	tt := []struct {
		name           string
		f              func() (string, error)
		serverResponse string
		expected       string
		expectedLocal  string
	}{
		{"internal dns", GetInternalDNS, "test.europe-west2-c.c.test-project.internal", "test.europe-west2-c.c.test-project.internal", "localhost"},
		{"subnet mask", GetSubnetMask, "255.255.255.16", "255.255.255.16", "255.255.0.0"},
		{"zone", GetZone, "projects/1016716848681/zones/us-west1-b", "us-west1-b", "europe-west2-c"},
		{"external ip", GetExternalIP, "33.44.55.66", "33.44.55.66", "127.0.0.1"},
		{"internal ip", GetInternalIP, "192.168.0.3", "192.168.0.3", "127.0.0.1"},
		{"project id", GetProjectID, "test-project", "test-project", "local-test-project"},
	}

	for _, tc := range tt {
		// reset
		ClearCache()
		os.Setenv("LOCAL_ENV", "")

		// happy path
		resp := tc.serverResponse
		webtest.DoFunc = func(req *http.Request) (*http.Response, error) {
			return webtest.MockResponse(200, resp, nil)
		}

		val, err := tc.f()
		checkErr(t, err)
		if val != tc.expected {
			t.Errorf("happy path; wanted: %v, got: %v", resp, val)
		}

		// if cache isnt cleared it will use the previous response
		ClearCache()

		// valid error but with 404 payload
		resp = `
  <!DOCTYPE html>
  <html lang=en>
  <meta charset=utf-8>
  <meta name=viewport content="initial-scale=1, minimum-scale=1, width=device-width">
  <title>Error 404 (Not Found)!!1</title>
  <style>
  </style>
  <a href=//www.google.com/><span id=logo aria-label=Google></span></a>
  <p><b>404.</b> <ins>That’s an error.</ins>
  <p>The requested URL <code>stub</code> was not found on this server.  <ins>That’s all we know.</ins>`
		webtest.DoFunc = func(req *http.Request) (*http.Response, error) {
			return webtest.MockResponse(200, resp, nil)
		}

		if _, err := tc.f(); err == nil {
			t.Errorf("%v; expected error on 404 html response", tc.name)
		}

		ClearCache()

		// query error
		resp = `stub`
		webtest.DoFunc = func(req *http.Request) (*http.Response, error) {
			return webtest.MockResponse(500, resp, nil)
		}
		if _, err := tc.f(); err == nil {
			t.Errorf("%v; expected error on 500 response", tc.name)
		}

		// local run
		os.Setenv("LOCAL_ENV", "true")
		val, err = tc.f()
		checkErr(t, err)
		if val != tc.expectedLocal {
			t.Errorf("local run; wanted: %v, got: %v", tc.expectedLocal, val)
		}
	}
}

func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Helper()
		t.Fatal(err)
	}
}
