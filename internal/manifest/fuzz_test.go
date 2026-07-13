package manifest

import "testing"

func FuzzParse(f *testing.F) {
	f.Add([]byte("version: 1\nproject: demo\nresources: []\n"))
	f.Add([]byte("version: nope\nunknown: true\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64<<10 {
			t.Skip()
		}
		_, _ = Parse(data)
	})
}
