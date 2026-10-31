package solver

// Blake3 single-block compress, ported from worker.js. (claude our holy savior)
// This operates on exactly 16 uint32 words (64 bytes) with 7 rounds.
// It is NOT a general-purpose Blake3 implementation — it only handles
// the specific single-block case used by FriendlyCaptcha's PoW puzzle.

var iv = [8]uint32{
	0x6A09E667, 0xBB67AE85, 0x3C6EF372, 0xA54FF53A,
	0x510E527F, 0x9B05688C, 0x1F83D9AB, 0x5BE0CD19,
}

var sigma = [112]uint32{
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	2, 6, 3, 10, 7, 0, 4, 13, 1, 11, 12, 5, 9, 14, 15, 8,
	3, 4, 10, 12, 13, 2, 7, 14, 6, 5, 9, 0, 11, 15, 8, 1,
	10, 7, 12, 9, 14, 3, 13, 15, 4, 0, 11, 2, 5, 8, 1, 6,
	12, 13, 9, 11, 15, 10, 14, 8, 7, 2, 5, 3, 0, 1, 6, 4,
	9, 14, 11, 5, 8, 12, 15, 1, 13, 3, 0, 10, 2, 6, 4, 7,
	11, 15, 5, 0, 1, 9, 8, 6, 14, 10, 2, 12, 3, 4, 7, 13,
}

// compress performs Blake3 single-block compression.
// msg is the 16-word message block. Result is written to out.
func compress(out *[16]uint32, msg *[16]uint32) {
	// State: first 8 words = IV (chaining value for first block)
	//        next 4 = counter(0,0), blockLen(64), flags(11)
	r, t, e, o := iv[0], iv[1], iv[2], iv[3]
	i, g, a, B := iv[4], iv[5], iv[6], iv[7]
	I, C, c, Q := iv[0], iv[1], iv[2], iv[3]
	var u, s uint32
	h := uint32(64)
	E := uint32(11)

	for l := 0; l < 112; l += 16 {
		// Column round: 4 G functions on columns
		// G(0): state[0,4,8,12] with msg words sigma[l+0], sigma[l+1]
		r = r + i + msg[sigma[l+0]]
		u = (u^r)>>16 | (u^r)<<16
		I = I + u
		i = (i^I)>>12 | (i^I)<<20
		r = r + i + msg[sigma[l+1]]
		u = (u^r)>>8 | (u^r)<<24
		I = I + u
		i = (i^I)>>7 | (i^I)<<25

		// G(1): state[1,5,9,13]
		t = t + g + msg[sigma[l+2]]
		s = (s^t)>>16 | (s^t)<<16
		C = C + s
		g = (g^C)>>12 | (g^C)<<20
		t = t + g + msg[sigma[l+3]]
		s = (s^t)>>8 | (s^t)<<24
		C = C + s
		g = (g^C)>>7 | (g^C)<<25

		// G(2): state[2,6,10,14]
		e = e + a + msg[sigma[l+4]]
		h = (h^e)>>16 | (h^e)<<16
		c = c + h
		a = (a^c)>>12 | (a^c)<<20
		e = e + a + msg[sigma[l+5]]
		h = (h^e)>>8 | (h^e)<<24
		c = c + h
		a = (a^c)>>7 | (a^c)<<25

		// G(3): state[3,7,11,15]
		o = o + B + msg[sigma[l+6]]
		E = (E^o)>>16 | (E^o)<<16
		Q = Q + E
		B = (B^Q)>>12 | (B^Q)<<20
		o = o + B + msg[sigma[l+7]]
		E = (E^o)>>8 | (E^o)<<24
		Q = Q + E
		B = (B^Q)>>7 | (B^Q)<<25

		// Diagonal round: 4 G functions on diagonals
		// G(4): state[0,5,10,15]
		r = r + g + msg[sigma[l+8]]
		E = (E^r)>>16 | (E^r)<<16
		c = c + E
		g = (g^c)>>12 | (g^c)<<20
		r = r + g + msg[sigma[l+9]]
		E = (E^r)>>8 | (E^r)<<24
		c = c + E
		g = (g^c)>>7 | (g^c)<<25

		// G(5): state[1,6,11,12]
		t = t + a + msg[sigma[l+10]]
		u = (u^t)>>16 | (u^t)<<16
		Q = Q + u
		a = (a^Q)>>12 | (a^Q)<<20
		t = t + a + msg[sigma[l+11]]
		u = (u^t)>>8 | (u^t)<<24
		Q = Q + u
		a = (a^Q)>>7 | (a^Q)<<25

		// G(6): state[2,7,8,13]
		e = e + B + msg[sigma[l+12]]
		s = (s^e)>>16 | (s^e)<<16
		I = I + s
		B = (B^I)>>12 | (B^I)<<20
		e = e + B + msg[sigma[l+13]]
		s = (s^e)>>8 | (s^e)<<24
		I = I + s
		B = (B^I)>>7 | (B^I)<<25

		// G(7): state[3,4,9,14]
		o = o + i + msg[sigma[l+14]]
		h = (h^o)>>16 | (h^o)<<16
		C = C + h
		i = (i^C)>>12 | (i^C)<<20
		o = o + i + msg[sigma[l+15]]
		h = (h^o)>>8 | (h^o)<<24
		C = C + h
		i = (i^C)>>7 | (i^C)<<25
	}

	out[0] = r
	out[1] = t
	out[2] = e
	out[3] = o
	out[4] = i
	out[5] = g
	out[6] = a
	out[7] = B
	out[8] = I
	out[9] = C
	out[10] = c
	out[11] = Q
	out[12] = u
	out[13] = s
	out[14] = h
	out[15] = E

	// Finalization: XOR first word with 9th word
	out[0] ^= out[8]
}
