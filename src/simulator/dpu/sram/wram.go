package sram

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"os"
	"uPIMulator/src/abi/encoding"
	"uPIMulator/src/misc"
)

type Wram struct {
	Address_ int64 `json:"address"`
	Size_    int64 `json:"size"`

	ByteStream_ *encoding.ByteStream `json:"byte_stream"`
}

func (this *Wram) Init() {
	config_loader := new(misc.ConfigLoader)
	config_loader.Init()

	this.Address_ = config_loader.WramOffset()
	this.Size_ = config_loader.WramSize()

	this.ByteStream_ = new(encoding.ByteStream)
	this.ByteStream_.Init()
	for i := int64(0); i < this.Size_; i++ {
		this.ByteStream_.Append(0)
	}
}

func (this *Wram) Fini() {
}

func (this *Wram) Address() int64 {
	return this.Address_
}

func (this *Wram) Size() int64 {
	return this.Size_
}

func (this *Wram) Read(address int64, size int64) *encoding.ByteStream {
	byte_stream := new(encoding.ByteStream)
	byte_stream.Init()

	for i := int64(0); i < size; i++ {
		index := this.Index(address) + int(i)

		byte_stream.Append(this.ByteStream_.Get(index))
	}

	return byte_stream
}

func (this *Wram) Write(address int64, size int64, byte_stream *encoding.ByteStream) {
	if size != byte_stream.Size() {
		err := errors.New("size != byte stream's size")
		panic(err)
	}

	for i := int64(0); i < size; i++ {
		index := this.Index(address) + int(i)

		this.ByteStream_.Set(index, byte_stream.Get(int(i)))
	}
}

func (this *Wram) Index(address int64) int {
	if address < this.Address_ {
		err := errors.New("address < WRAM offset")
		panic(err)
	} else if address >= this.Address_+this.Size_ {
		err := errors.New("address >= WRAM offset + WRAM size")
		panic(err)
	}

	return int(address - this.Address_)
}

func (w *Wram) SaveToJson(filename string) error {
	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func (w *Wram) Replace(filename string) error {
	newWram, err := LoadFromJson(filename)
	if err != nil {
		return err
	}
	*w = *newWram
	return nil
}

func LoadFromJson(filename string) (*Wram, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var wram Wram
	err = json.Unmarshal(data, &wram)
	if err != nil {
		return nil, err
	}
	return &wram, nil
}

func (w *Wram) SaveToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(w); err != nil {
		return err
	}
	return nil
}

func LoadFromFile(filename string, w *Wram) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(w); err != nil {
		return err
	}
	return nil
}
