package codes_test

import (
	"testing"

	"github.com/lmorg/mxtty/codes"
	"github.com/lmorg/mxtty/types"
)

func TestGetAnsiEscSeqWithModifer(t *testing.T) {
	b := codes.GetAnsiEscSeq(types.KeysNormal, codes.AnsiF5, codes.MOD_SHIFT)
	if string(b) != string(codes.Csi)+"15;2~" {
		t.Errorf("Incorrect string '%s'", string(b))
	}
}
