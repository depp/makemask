package main

import (
	"flag"
	"fmt"
	"os"
)

func makeMask(filename string, bootdata []byte) error {
	fmt.Printf("Masking %s.\n", filename)
	f, err := os.OpenFile(filename, os.O_RDWR, 0755)
	if err != nil {
		return fmt.Errorf("could not open ROM: %v", err)
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return err
	}
	const minSize = checksumStart + checksumLength
	if st.Size() < minSize {
		fmt.Fprintf(os.Stderr, "Padding to %d bytes\n", minSize)
		if err := f.Truncate(minSize); err != nil {
			return err
		}
	}
	_, err = f.WriteAt(bootdata, 0x40)
	if err != nil {
		return fmt.Errorf("could not write: %v", err)
	}
	crcBytes, err := calculateN64Crc(f, 6102)
	if err != nil {
		return fmt.Errorf("could not calculate CRC: %v", err)
	}
	fmt.Printf("Generate checksum: 0x%X\n", crcBytes)
	_, err = f.WriteAt(crcBytes[:], 0x10)
	if err != nil {
		return fmt.Errorf("could not write: %v", err)
	}
	return nil
}

func mainE() error {
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: makemask [rom file, rom file, ...]")
		os.Exit(64)
	}
	bootdata, err := Asset("data/boot.6102")
	if err != nil {
		return fmt.Errorf("could not load boot data: %v", err)
	}
	for _, arg := range flag.Args() {
		if err := makeMask(arg, bootdata); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := mainE(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
