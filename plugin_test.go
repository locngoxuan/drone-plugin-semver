package main

import "testing"

func TestValidVersion(t *testing.T) {
	p := &Plugin{
		Config: Config{
			Src: "test/VALID_VERSION",
		},
	}
	err := p.Exec()
	if err != nil {
		t.Error(err)
	}
}

func TestInValidVersion(t *testing.T) {
	p := &Plugin{
		Config: Config{
			Src: "test/INVALID_VERSION",
		},
	}
	err := p.Exec()
	if err == nil {
		t.Errorf(`version should be invalid`)
	}
}
