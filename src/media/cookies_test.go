package media

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testCookies = "# Netscape HTTP Cookie File\n.tiktok.com\tTRUE\t/\tTRUE\t0\tsessionid\ttest-secret\n"

func TestConfiguredCookieFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	for _, configuredDir := range []string{"", "relative cookies", filepath.Join(dir, "absolute cookies")} {
		t.Run(configuredDir, func(t *testing.T) {
			t.Setenv("MR_COOKIES_DIR", configuredDir)
			cookieDir := configuredDir
			if cookieDir == "" {
				cookieDir = "cookies"
			}
			if err := os.MkdirAll(cookieDir, 0700); err != nil {
				t.Fatal(err)
			}
			if path, err := configuredCookieFile(nil); path != "" || err != nil {
				t.Fatalf("absent cookie file should be optional: %q %v", path, err)
			}
			filename, err := filepath.Abs(filepath.Join(cookieDir, "cookies.txt"))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filename, []byte(testCookies), 0600); err != nil {
				t.Fatal(err)
			}
			if path, err := configuredCookieFile(nil); path != filename || err != nil {
				t.Fatalf("expected configured file %q, got %q: %v", filename, path, err)
			}
			for _, option := range []string{"--cookies", "--no-cookies", "--cookies-from-browser"} {
				if path, err := configuredCookieFile(map[string]string{option: ""}); path != "" || err != nil {
					t.Fatalf("explicit %s must override the configured file: %q %v", option, path, err)
				}
			}
		})
	}
}

func TestCookieDirectoryRejected(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MR_COOKIES_DIR", dir)
	if err := os.Mkdir(filepath.Join(dir, "cookies.txt"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := configuredCookieFile(nil); err == nil {
		t.Fatal("a directory named cookies.txt should be rejected")
	}
}

func TestDownloadUsesConfiguredCookies(t *testing.T) {
	dir := t.TempDir()
	cookieDir := filepath.Join(dir, "mounted cookies")
	if err := os.Mkdir(cookieDir, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MR_COOKIES_DIR", cookieDir)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	capture := filepath.Join(dir, "arguments")
	t.Setenv("TEST_ARGUMENTS", capture)
	// Stand in for yt-dlp without network access or real credentials.
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$TEST_ARGUMENTS\"\n" +
		"echo test-secret >&2\nexit \"$TEST_EXIT_CODE\"\n"
	if err := os.WriteFile(filepath.Join(dir, "yt-dlp"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(cookieDir, "cookies.txt")
	if err := os.WriteFile(filename, []byte(testCookies), 0600); err != nil {
		t.Fatal(err)
	}
	for _, exitCode := range []string{"0", "1"} {
		t.Run(exitCode, func(t *testing.T) {
			t.Setenv("TEST_EXIT_CODE", exitCode)
			_, message, err := downloadMedia("https://tiktok.com/@user/video/123", nil)
			if (err != nil) != (exitCode == "1") {
				t.Fatalf("unexpected download result: %v", err)
			}
			if exitCode == "1" && (strings.Contains(message, "test-secret") || !strings.Contains(message, "fresh cookie export")) {
				t.Fatalf("cookie download error not handled safely: %s", message)
			}
			data, err := os.ReadFile(capture)
			if err != nil {
				t.Fatal(err)
			}
			args := strings.Split(string(data), "\n")
			if len(args) < 3 || args[0] != "--cookies" || args[1] != filename {
				t.Fatalf("yt-dlp must receive the mounted cookie file directly: %q", args)
			}
			if _, err := os.Stat(filename); err != nil {
				t.Fatalf("download removed the configured cookie file: %v", err)
			}
		})
	}
}

func TestGetUrlIncludesPreset(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/download?preset=ios&url=https://example.com/video", nil)
	url, args := getUrl(r)
	if url != "https://example.com/video" {
		t.Fatalf("unexpected url: %q", url)
	}
	if args["preset"] != "ios" {
		t.Fatalf("expected preset=ios, got %q", args["preset"])
	}
}

func TestIndexHasNoCookieUpload(t *testing.T) {
	t.Setenv("MR_MEDIA_LIST_ENABLED", "false")
	w := httptest.NewRecorder()
	Index(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusOK || strings.Contains(w.Body.String(), "type=\"file\"") || strings.Contains(w.Body.String(), "/cookies/") {
		t.Fatal("home page must render without cookie upload controls")
	}
}
