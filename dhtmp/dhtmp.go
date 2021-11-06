package dhtmp

import (
	"fmt"
	"math"
)

const (
	h0d uint32 = 0x64756C6C
	h1d uint32 = 0x68617368
	h2d uint32 = 0x20697320
	h3d uint32 = 0x6120706F
	h4d uint32 = 0x6F722068
	h5d uint32 = 0x61736820
	h6d uint32 = 0x66756E63
	h7d uint32 = 0x74696F6E
)

func chunkify(data []byte) [][16]uint32 {
	data = append(data, 128)
	dataLen := len(data) * 8

	if len(data)%64 > 56 {
		for i := 0; i < len(data)%64; i++ {
			data = append(data, 0)
		}
	}
	for i := 0; i < (len(data)%64)%4; i++ {
		data = append(data, 0)
	}

	chunks := make([][16]uint32, (len(data)/64)+1)

	for i := 0; i < len(chunks)-1; i++ {
		for j := 0; j < 16; j++ {
			chunks[i][j] = uint32(data[(j*4)+(i*64)])<<24 |
				uint32(data[(j*4)+(i*64)+1])<<16 |
				uint32(data[(j*4)+(i*64)+2])<<8 |
				uint32(data[(j*4)+(i*64)+3])
		}
	}

	for i := 0; i < (len(data)%64)/4; i++ {
		chunks[len(chunks)-1][i] = uint32(data[((len(chunks)/64)*64)+(i*4)])<<24 |
			uint32(data[((len(chunks)/64)*64)+(i*4)+1])<<16 |
			uint32(data[((len(chunks)/64)*64)+(i*4)+2])<<8 |
			uint32(data[((len(chunks)/64)*64)+(i*4)+3])
	}

	chunks[len(chunks)-1][14] = uint32(dataLen >> 32)
	chunks[len(chunks)-1][15] = uint32(dataLen - (dataLen >> 32))

	return chunks
}

func addOverflow(x, y uint32) uint32 {
	if y > x {
		x, y = y, x
	}
	if y > math.MaxUint32-x {
		return y - (math.MaxUint32 - x)
	}
	return x + y
}

func Sum(data []byte) [32]byte {
	chunks := chunkify(data)
	//h0, h1, h2, h3, h4, h5, h6, h7 := h0d, h1d, h2d, h3d, h4d, h5d, h6d, h7d
	//for _, chunk := range chunks {
	//	a, b, c, d, e, f, g, h := h0, h1, h2, h3, h4, h5, h6, h7
	//
	//
	//
	//	h0, h1, h2, h3 = addOverflow(h0, a), addOverflow(h1, b), addOverflow(h2, c), addOverflow(h3, d)
	//	h4, h5, h6, h7 = addOverflow(h4, e), addOverflow(h5, f), addOverflow(h6, g), addOverflow(h7, h)
	//}

	for i, d := range data {
		fmt.Printf("%d: %08b ", i, d)
	}
	fmt.Printf("\n")
	for i, chunk := range chunks {
		fmt.Printf("chunk %02d:\n", i)
		for j, n := range chunk {
			fmt.Printf("%03d: %032b\n", j, n)
		}
	}

	return [32]byte{}
}
