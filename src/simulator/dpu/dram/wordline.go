package dram

import (
	"errors"
	"uPIMulator/src/abi/encoding"
	"uPIMulator/src/misc"
)

type Wordline struct {
	Address     int64                `json:"address"`
	Size        int64                `json:"size"`
	Byte_stream *encoding.ByteStream `json:"byte_stream"`
}

func (this *Wordline) Init(address int64, size int64) {
	config_loader := new(misc.ConfigLoader)
	config_loader.Init()

	mram_data_size := int64(config_loader.MramDataWidth() / 8)

	if address < 0 {
		err := errors.New("address < 0")
		panic(err)
	} else if size <= 0 {
		err := errors.New("size <= 0")
		panic(err)
	} else if size%mram_data_size != 0 {
		err := errors.New("size is not aligned with MRAM data size")
		panic(err)
	}

	this.Address = address
	this.Size = size

	this.Byte_stream = new(encoding.ByteStream)
	this.Byte_stream.Init()
	for i := int64(0); i < size; i++ {
		this.Byte_stream.Append(0)
	}
}

func (this *Wordline) Fini() {
}

func (this *Wordline) Read() *encoding.ByteStream {
	byte_stream := new(encoding.ByteStream)
	byte_stream.Init()

	for i := int64(0); i < this.Byte_stream.Size(); i++ {
		byte_stream.Append(this.Byte_stream.Get(int(i)))
	}

	return byte_stream
}

// attention
func (this *Wordline) Write(byte_stream *encoding.ByteStream) {
	// if this.Size != byte_stream.Size() {
	// 	err := errors.New("wordline's size != byte stream's size")
	// 	panic(err)
	// }

	for i := int64(0); i < byte_stream.Size(); i++ {
		this.Byte_stream.Set(int(i), byte_stream.Get(int(i)))
	}
}
