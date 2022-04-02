package ssh

import "testing"

func TestGenerateSSHKeys(t *testing.T) {
	private, public, err := GenerateSSHKeys()
	if err != nil {
		t.Fatal("error generating ssh keys")
	}

	if private == "" || public == "" {
		t.Fatal("some of the keys are empty")
	}
}