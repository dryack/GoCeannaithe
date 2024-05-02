package bloom

import (
	"fmt"
	"github.com/dryack/GoCeannaithe/pkg/common"
	"os"
	"testing"
)

func TestFilePersistence_getPath(t *testing.T) {
	type testCase[T common.Hashable] struct {
		name string
		fp   FilePersistence[T]
		want string
	}
	tests := []testCase[int]{
		{
			name: "base test",
			fp: FilePersistence[int]{
				directory: ".",
				filename:  "test.dat",
			},
			want: fmt.Sprintf(".%ctest.dat", os.PathSeparator),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fp.getFullPath(); got != tt.want {
				t.Errorf("getFullPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
