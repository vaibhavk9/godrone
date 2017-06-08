package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestHelloWorld(t *testing.T) {
	if HelloWorld() != "To hell with this world !" {
		t.Errorf("got %s expected %s", HelloWorld(), "To hell with this world !")
	}
}
