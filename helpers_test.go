package app

import (
	"testing"
)

func TestCallerFunctionName(t *testing.T) {
	want := "github.com/MrEhbr/app.firstFunction"
	got := firstFunction()
	if want != got {
		t.Fatalf("want: %s, got: %s", want, got)
	}
}

func firstFunction() string {
	return secondFunction()
}

func secondFunction() string {
	return CallerFunctionName()
}

func TestCurrentFunctionName(t *testing.T) {
	want := "github.com/MrEhbr/app.currFunc"
	got := currFunc()
	if want != got {
		t.Fatalf("want: %s, got: %s", want, got)
	}
}

func currFunc() string {
	return CurrentFunctionName()
}

func Benchmark_getFrame(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = currFunc()
	}
}
