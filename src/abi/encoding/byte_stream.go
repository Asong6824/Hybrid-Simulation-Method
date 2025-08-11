package encoding

import (
	"encoding/json"
	"strconv"
	"strings"
	"uPIMulator/src/misc"
)

type ByteStream struct {
	Bytes []uint8 `json:"bytes"`
}

func (this *ByteStream) Init() {
	this.Bytes = make([]uint8, 0)
}

func (this *ByteStream) InitWithSize(size int64) {
	this.Bytes = make([]uint8, size)
}

func (this *ByteStream) Size() int64 {
	return int64(len(this.Bytes))
}

func (this *ByteStream) Get(pos int) uint8 {
	return this.Bytes[pos]
}

func (this *ByteStream) Set(pos int, value uint8) {
	this.Bytes[pos] = value
}

func (this *ByteStream) Append(value uint8) {
	this.Bytes = append(this.Bytes, value)
}

func (this *ByteStream) Merge(byte_stream *ByteStream) {
	for i := int64(0); i < byte_stream.Size(); i++ {
		value := byte_stream.Get(int(i))
		this.Append(value)
	}
}

func (this *ByteStream) MergeMemoryBlocks(other *ByteStream, otherStartAddress int64) {
	config_loader := new(misc.ConfigLoader)
	config_loader.Init()

	thisStartAddress := config_loader.MramOffset()

	var mergedSize int64
	otherEndAddress := otherStartAddress + other.Size()

	if this.Size() > 0 {
		thisEndAddress := thisStartAddress + this.Size()
		mergedSize = maxInt64(thisEndAddress, otherEndAddress) - thisStartAddress
	} else {
		mergedSize = otherEndAddress - thisStartAddress
	}

	mergedStream := new(ByteStream)
	mergedStream.InitWithSize(mergedSize)

	for i := int64(0); i < this.Size(); i++ {
		mergedStream.Set(int(i), this.Get(int(i)))
	}

	for i := int64(0); i < other.Size(); i++ {
		mergedStream.Set(int(otherStartAddress-thisStartAddress+i), other.Get(int(i)))
	}

	this.Bytes = mergedStream.Bytes
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func (b *ByteStream) MarshalJSON() ([]byte, error) {
	if b == nil {
		return []byte("null"), nil
	}
	var result strings.Builder
	result.WriteString("[")
	for i, byteVal := range b.Bytes {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(strconv.Itoa(int(byteVal)))
	}
	result.WriteString("]")
	return []byte(result.String()), nil
}

func (b *ByteStream) UnmarshalJSON(data []byte) error {
	var intSlice []int
	if err := json.Unmarshal(data, &intSlice); err != nil {
		return err
	}
	b.Bytes = make([]uint8, len(intSlice))
	for i, num := range intSlice {
		b.Bytes[i] = uint8(num)
	}
	return nil
}
