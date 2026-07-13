package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fake struct {
	provider, name, status string
	resetErr               error
}

func (f fake) Provider() string                    { return f.provider }
func (f fake) Name() string                        { return f.name }
func (f fake) Health(context.Context) HealthStatus { return HealthStatus{Status: f.status} }
func (f fake) Reset(context.Context) error         { return f.resetErr }
func TestRegistry(t *testing.T) {
	r := NewRegistry()
	for _, s := range []fake{{"aws", "sqs", "ok", nil}, {"aws", "s3", "ok", nil}} {
		if err := r.Register(s); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.Register(fake{"aws", "s3", "ok", nil}); err == nil {
		t.Fatal("duplicate accepted")
	}
	got := r.Services()
	names := []string{got[0].Name(), got[1].Name()}
	if !reflect.DeepEqual(names, []string{"s3", "sqs"}) {
		t.Fatalf("order=%v", names)
	}
	if _, ok := r.Find("aws", "s3"); !ok {
		t.Fatal("lookup failed")
	}
	copyView := r.Services()
	copyView[0] = nil
	if r.Services()[0] == nil {
		t.Fatal("mutable view")
	}
}
func TestRegistryHealthAndResetError(t *testing.T) {
	r := NewRegistry()
	r.Register(fake{"aws", "bad", "error", errors.New("failed")})
	if r.Health(context.Background())["aws.bad"].Status != "error" {
		t.Fatal("health hidden")
	}
	if err := r.Reset(context.Background()); err == nil {
		t.Fatal("reset error hidden")
	}
}
