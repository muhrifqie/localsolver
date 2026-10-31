package solver

import (
	"encoding/binary"
	"encoding/hex"
	"testing"
)

func TestBlake3AgainstHAR(t *testing.T) {
	// Known puzzle from HAR capture:
	// Config: 3fbdd269b1020102010001005d5700006cf3d5f1f72e9235
	// Solution for solution_number=0: solver=2, retry=0, nonce=122019
	// Threshold: 22365 (0x575d)
	configHex := "3fbdd269b1020102010001005d5700006cf3d5f1f72e9235"
	configBytes, _ := hex.DecodeString(configHex)

	threshold := binary.LittleEndian.Uint32(configBytes[12:16])
	t.Logf("Threshold: %d (0x%x)", threshold, threshold)

	// Load config into uint32 array (zero-padded to 64 bytes)
	var msg [16]uint32
	for i := 0; i+3 < len(configBytes); i += 4 {
		msg[i/4] = binary.LittleEndian.Uint32(configBytes[i : i+4])
	}

	// Set solution_number=0, solver_type=2 (wasm) at word 14
	msg[14] = 0 | (2 << 8) // solution_number=0, solver=2, retry=0

	// Set known nonce from HAR: 122019
	msg[15] = 122019

	var out [16]uint32
	compress(&out, &msg)

	t.Logf("blake3 output[0]: %d (0x%08x)", out[0], out[0])
	t.Logf("threshold:        %d (0x%08x)", threshold, threshold)

	if out[0] >= threshold {
		t.Errorf("Expected blake3 output[0] (%d) < threshold (%d)", out[0], threshold)
	} else {
		t.Logf("PASS: %d < %d", out[0], threshold)
	}

	// Also verify that nonce 122018 does NOT pass
	msg[15] = 122018
	compress(&out, &msg)
	t.Logf("nonce 122018 -> output[0]: %d (pass: %v)", out[0], out[0] < threshold)
}
