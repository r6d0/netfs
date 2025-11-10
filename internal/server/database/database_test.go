package database_test

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
)

const size = 4294967294

// func TestRead(t *testing.T) {
// 	file, err := os.Open("./file")
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("OPEN_FILE")

// 	data := make([]byte, 4)
// 	_, err = file.ReadAt(data, 4*10)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(binary.BigEndian.Uint32(data))

// 	file.Close()

// 	SeqSearch(file, uint32(834148))
// 	BinSearch(file, uint32(834148), size)
// }

func TestCreateFile(t *testing.T) {
	file, err := os.OpenFile("./file", os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	data := make([]byte, 4)
	for i := range size {
		binary.BigEndian.PutUint32(data, uint32(i))

		_, err := file.Write(data)
		if err != nil {
			panic(err)
		}
	}
	file.Close()
}

func BenchmarkSeqSearch(b *testing.B) {
	file, err := os.Open("./file")
	if err != nil {
		panic(err)
	}
	fmt.Println("OPEN_FILE")

	value := uint32(834148)
	for b.Loop() {
		SeqSearch(file, value)
	}
}

func BenchmarkBinSearch(b *testing.B) {
	file, err := os.Open("./file")
	if err != nil {
		panic(err)
	}
	fmt.Println("OPEN_FILE")

	value := uint32(834148)
	for b.Loop() {
		BinSearch(file, value, size)
	}
}

func SeqSearch(file *os.File, seek uint32) {
	data := make([]byte, 4)
	end := false
	for !end {
		_, err := file.Read(data)
		if errors.Is(err, io.EOF) {
			panic("NOT_FOUND")
		} else {
			value := binary.BigEndian.Uint32(data)
			if value == seek {
				end = true
			}
		}
	}
}

func BinSearch(file *os.File, seek uint32, size uint32) {
	index := 0
	begin := 0
	end := int(size - 1)

	data := make([]byte, 4)
	for begin <= end {
		index = (begin + end) / 2

		_, err := file.ReadAt(data, int64(index*4))
		if errors.Is(err, io.EOF) {
			panic("NOT_FOUND")
		} else {
			value := binary.BigEndian.Uint32(data)
			if value == seek {
				return
			} else if value > seek {
				end = index - 1
			} else if value < seek {
				begin = index + 1
			}
		}
	}
}
