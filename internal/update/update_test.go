package update

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func serveLatest(t *testing.T, status int, body string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	previous := latestURL
	latestURL = server.URL
	t.Cleanup(func() { latestURL = previous })
}

func TestCheckLatestReturnsTag(t *testing.T) {
	serveLatest(t, http.StatusOK, `{"tag_name":"v1.2.3"}`)

	got, err := CheckLatest()
	if err != nil || got != "v1.2.3" {
		t.Errorf("CheckLatest = %q, %v; want v1.2.3, nil", got, err)
	}
}

func TestCheckLatestNoReleases(t *testing.T) {
	serveLatest(t, http.StatusNotFound, `{}`)

	if _, err := CheckLatest(); !errors.Is(err, ErrNoReleases) {
		t.Errorf("err = %v, want ErrNoReleases", err)
	}
}

func TestCheckLatestEmptyTagMeansNoReleases(t *testing.T) {
	serveLatest(t, http.StatusOK, `{"tag_name":""}`)

	if _, err := CheckLatest(); !errors.Is(err, ErrNoReleases) {
		t.Errorf("err = %v, want ErrNoReleases", err)
	}
}

func TestCheckLatestServerError(t *testing.T) {
	serveLatest(t, http.StatusInternalServerError, ``)

	if _, err := CheckLatest(); err == nil || errors.Is(err, ErrNoReleases) {
		t.Errorf("err = %v, want a non-ErrNoReleases error", err)
	}
}

func TestCheckLatestBadJSON(t *testing.T) {
	serveLatest(t, http.StatusOK, `{not json`)

	if _, err := CheckLatest(); err == nil {
		t.Error("CheckLatest with bad JSON succeeded, want error")
	}
}

func TestIsNewer(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"0.1.0", "v0.2.0", true},
		{"v0.1.0", "v0.1.1", true},
		{"0.1.0", "1.0.0", true},
		{"1.0.0", "1.0.0", false},
		{"1.2.0", "1.1.9", false},
		{"2.0.0", "v1.9.9", false},
		{"garbage", "1.0.0", false},
		{"1.0.0", "garbage", false},
	}
	for _, c := range cases {
		if got := IsNewer(c.current, c.latest); got != c.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", c.current, c.latest, got, c.want)
		}
	}
}
