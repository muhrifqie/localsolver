package solver

//bless claude opus 🙏

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// PuzzleInput layout offsets
const (
	offTimestampSec  = 0  // uint32 LE - solveContextTimestamp / 1000
	offTimestampMs   = 4  // uint16 LE - solveContextTimestamp % 1000
	offConfigVersion = 6  // uint8    - configurationVersion (1)
	offSolveMode     = 7  // uint8    - solveMode (1=noninteractive, 2=active)
	offPoWType       = 8  // uint16 LE - poWType (1)
	offPoWVersion    = 10 // uint16 LE - poWVersion (1)
	offThreshold     = 12 // uint32 LE - solutionThreshold
	offContextNonce  = 16 // uint32 LE - solveContextNonce
	offSubmitNonce   = 20 // uint32 LE - submissionNonce
	puzzleSize       = 24 // total bytes

	solutionsPerWorker = 36 // "noninteractive" mode (not "passive" which uses 16)
	minWorkPerSolution = 100000
)

type PuzzleParams struct {
	Timestamp   int64  // solveContextTimestamp in ms
	Nonce       uint32 // solveContextNonce
	SubmitNonce uint32 // submissionNonce (random)
	Cost        int    // cost from quote response (total hash work)
}

type SolveResult struct {
	Submission string  // "configBase64.solutionsBase64.metaBase64"
	HashRate   float64 // aggregate hashes/sec
	Duration   int     // wall time in ms
}

// ComputePuzzle calculates threshold and solution count from the API cost value.
// Mirrors the JS: r = max(ceil(cost/36), 100000); threshold = floor(0xFFFFFFFF / r)
func ComputePuzzle(cost int) (threshold uint32, numSolutions int) {
	workPerSolution := int(math.Max(math.Ceil(float64(cost)/solutionsPerWorker), minWorkPerSolution))
	numSolutions = int(math.Ceil(float64(cost) / float64(workPerSolution)))
	threshold = uint32(math.Floor(4294967295 / float64(workPerSolution)))
	return
}

// BuildPuzzleInput constructs the 24-byte puzzle buffer.
func BuildPuzzleInput(p PuzzleParams, threshold uint32) []byte {
	buf := make([]byte, puzzleSize)
	binary.LittleEndian.PutUint32(buf[offTimestampSec:], uint32(p.Timestamp/1000))
	binary.LittleEndian.PutUint16(buf[offTimestampMs:], uint16(p.Timestamp%1000))
	buf[offConfigVersion] = 1
	buf[offSolveMode] = 2 // active (d() is always called with mode:"active" for real solves)
	binary.LittleEndian.PutUint16(buf[offPoWType:], 1)
	binary.LittleEndian.PutUint16(buf[offPoWVersion:], 1)
	binary.LittleEndian.PutUint32(buf[offThreshold:], threshold)
	binary.LittleEndian.PutUint32(buf[offContextNonce:], p.Nonce)
	binary.LittleEndian.PutUint32(buf[offSubmitNonce:], p.SubmitNonce)
	return buf
}

// solveSingle finds a nonce where blake3_compress(puzzle)[0] < threshold.
// Returns the 8-byte solution (words 14 and 15 of the state).
func solveSingle(msg *[16]uint32, solutionNumber int, threshold uint32) ([8]byte, uint32, error) {
	// Set solution_number at byte 56 (word 14, byte 0)
	// Set solver type at byte 57 (word 14, byte 1) — 2 = wasm
	msg[14] = uint32(solutionNumber) | (2 << 8)
	msg[15] = 0

	var out [16]uint32
	var nonce uint32
	for nonce = 0; ; nonce++ {
		msg[15] = nonce
		compress(&out, msg)
		if out[0] < threshold {
			var sol [8]byte
			binary.LittleEndian.PutUint32(sol[0:4], msg[14])
			binary.LittleEndian.PutUint32(sol[4:8], msg[15])
			return sol, nonce + 1, nil
		}
		if nonce == 0xFFFFFFFF {
			return [8]byte{}, 0, fmt.Errorf("solution %d: exhausted nonce space", solutionNumber)
		}
	}
}

// SolveAll solves all puzzles for a given cost and returns a single submission string.
// The submission format is: configBase64.solutionsBase64.metaBase64
// where solutionsBase64 contains all 8-byte solutions concatenated.
func SolveAll(params PuzzleParams) (SolveResult, error) {
	threshold, numSolutions := ComputePuzzle(params.Cost)

	puzzleInput := BuildPuzzleInput(params, threshold)
	configB64 := base64.URLEncoding.EncodeToString(puzzleInput)

	// Load puzzle into 16 x uint32 (zero-padded to 64 bytes)
	var msg [16]uint32
	for i := 0; i+3 < len(puzzleInput); i += 4 {
		msg[i/4] = binary.LittleEndian.Uint32(puzzleInput[i : i+4])
	}

	start := time.Now()
	var totalHashes uint64
	allSolutions := make([]byte, 0, numSolutions*8)

	for i := 0; i < numSolutions; i++ {
		sol, hashes, err := solveSingle(&msg, i, threshold)
		if err != nil {
			return SolveResult{}, err
		}
		allSolutions = append(allSolutions, sol[:]...)
		totalHashes += uint64(hashes)
	}

	elapsed := time.Since(start)
	wallMs := int(elapsed.Milliseconds())
	hashRate := float64(totalHashes) / (elapsed.Seconds() + 0.0005)

	solutionsB64 := base64.URLEncoding.EncodeToString(allSolutions)

	meta := map[string]float64{
		"hash_rate_wall":        hashRate,
		"hash_rate_accumulated": hashRate,
	}
	metaJSON, _ := json.Marshal(meta)
	metaB64 := base64.URLEncoding.EncodeToString(metaJSON)

	return SolveResult{
		Submission: fmt.Sprintf("%s.%s.%s", configB64, solutionsB64, metaB64),
		HashRate:   hashRate,
		Duration:   wallMs,
	}, nil
}
