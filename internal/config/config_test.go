package config

import "testing"

func TestFromEnvironment(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		t.Setenv("EMULITH_ADDR", "")
		t.Setenv("EMULITH_DATA_DIR", "")
		got := FromEnvironment()
		if got.Addr != DefaultAddr || got.DataDir != DefaultDataDir {
			t.Fatalf("unexpected defaults: %+v", got)
		}
	})
	t.Run("environment", func(t *testing.T) {
		t.Setenv("EMULITH_ADDR", "127.0.0.1:9000")
		t.Setenv("EMULITH_DATA_DIR", "/tmp/emulith")
		got := FromEnvironment()
		if got.Addr != "127.0.0.1:9000" || got.DataDir != "/tmp/emulith" {
			t.Fatalf("unexpected environment config: %+v", got)
		}
	})
}
