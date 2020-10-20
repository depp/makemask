package main

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

const (
	checksumStart  = 0x00001000
	checksumLength = 0x00100000
)

// This file is mostly copied wholesale from http://n64dev.org/n64crc.html.

func rol(num uint32, shift uint32) uint32 {
	return (num << shift) | (num >> (32 - shift))
}

func calculateN64Crc(f *os.File, bootcode int) (crc [8]byte, err error) {
	const seed uint32 = 0xF8CA4DDC
	t1 := seed
	t2 := seed
	t3 := seed
	t4 := seed
	t5 := seed
	t6 := seed

	crcInts := [2]uint32{}
	if bootcode != 6102 {
		return crc, errors.New("non-6102 bootcodes not yet supported")
	}
	data := make([]byte, checksumLength)
	if _, err := f.ReadAt(data, checksumStart); err != nil {
		if err == io.EOF {
			return crc, errors.New("ROM too short")
		}
		return crc, err
	}

	for i := 0; i < checksumLength; i += 4 {
		d := binary.BigEndian.Uint32(data[i:])
		if (t6 + d) < t6 {
			t4++
		}
		t6 += d
		t3 ^= d
		r := rol(d, (d & 0x1F))
		t5 += r
		if t2 > d {
			t2 ^= r
		} else {
			t2 ^= t6 ^ d
		}

		if bootcode == 6105 {
			// TODO(tylerrhodes): figure this out
			//t1 += BYTES2LONG(&data[N64_HEADER_SIZE+0x0710+(i&0xFF)]) ^ d
		} else {
			t1 += t5 ^ d
		}
	}
	if bootcode == 6103 {
		crcInts[0] = (t6 ^ t4) + t3
		crcInts[1] = (t5 ^ t2) + t1
	} else if bootcode == 6106 {
		crcInts[0] = (t6 * t4) + t3
		crcInts[1] = (t5 * t2) + t1
	} else {
		crcInts[0] = t6 ^ t4 ^ t3
		crcInts[1] = t5 ^ t2 ^ t1
	}
	binary.BigEndian.PutUint32(crc[0:], crcInts[0])
	binary.BigEndian.PutUint32(crc[4:], crcInts[1])
	return crc, nil
}
