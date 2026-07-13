package aws_test

import (
	"github.com/alekpopovic/emulith/test/compatibility/aws/compat"
	"os"
	"testing"
)

func TestMain(m *testing.M) { code := m.Run(); compat.Write(); os.Exit(code) }
