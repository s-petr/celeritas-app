package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHome(t *testing.T) {
	routes := getRoutes()

	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/")
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("returned wrong HTTP response code; expected status 200 but got %d", resp.StatusCode)
	}

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(bodyText), "awesome") {
		cel.TakeScreenShot(ts.URL+"/", "HomeTest", 1920, 1080)
		t.Error("did not find keyword \"awesome\"")
	}
}

func TestHome2(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)

	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	cel.Session.Put(ctx, "test", "hello world")

	h := http.HandlerFunc(testHandlers.Home)
	h.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Errorf("returned wrong HTTP response code; expected status 200 but got %d", rr.Code)
	}

	if cel.Session.Get(ctx, "test") != "hello world" {
		t.Errorf("could not fetch saved value from session; expected test=\"hello world\", got test=\"%s\"", cel.Session.Get(ctx, "test"))
	}
}
