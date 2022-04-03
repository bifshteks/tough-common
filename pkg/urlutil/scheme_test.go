package urlutil

import "testing"

func TestGetScheme(t *testing.T) {
	var want string
	want = "ws://"
	if got := GetScheme("ws", false); want != got {
		t.Error("Not equal", want, got)
	}

	want = "wss://"
	if got := GetScheme("ws", true); want != got {
		t.Error("Not equal", want, got)
	}
	want = "http://"
	if got := GetScheme("http", false); want != got {
		t.Error("Not equal", want, got)
	}
	want = "https://"
	if got := GetScheme("http", true); want != got {
		t.Error("Not equal", want, got)
	}

	want = ""
	if got := GetScheme("___", false); want != got {
		t.Error("Not equal", want, got)
	}
	if got := GetScheme("qwerty", false); want != got {
		t.Error("Not equal", want, got)
	}
	if got := GetScheme("", false); want != got {
		t.Error("Not equal", want, got)
	}
}
