package dram

import (
	"encoding/json"
	"errors"
	"os"
	"uPIMulator/src/abi/encoding"
	"uPIMulator/src/global"
	"uPIMulator/src/misc"
)

type Mram struct {
	Address_   int64       `json:"address"`
	Size_      int64       `json:"size"`
	Wordlines_ []*Wordline `json:"wordlines"`
}

func (this *Mram) Init() {
	config_loader := new(misc.ConfigLoader)
	config_loader.Init()

	this.Address_ = config_loader.MramOffset()
	this.Size_ = config_loader.MramSize()

	if global.WordlineSize <= 0 {
		err := errors.New("wordline size <= 0")
		panic(err)
	} else if this.Address_%global.WordlineSize != 0 {
		err := errors.New("address is not aligned with wordline size")
		panic(err)
	} else if this.Size_%global.WordlineSize != 0 {
		err := errors.New("size is not aligned with wordline size")
		panic(err)
	}

	this.Wordlines_ = make([]*Wordline, 0)
	num_wordlines := int(this.Size_ / global.WordlineSize)
	for i := 0; i < num_wordlines; i++ {
		wordline := new(Wordline)
		wordline.Init(this.Address_+int64(i)*global.WordlineSize, global.WordlineSize)
		this.Wordlines_ = append(this.Wordlines_, wordline)
	}
}

func (this *Mram) Fini() {
	for _, wordline := range this.Wordlines_ {
		wordline.Fini()
	}
}

func (this *Mram) Address() int64 {
	return this.Address_
}

func (this *Mram) Size() int64 {
	return this.Size_
}

func (this *Mram) Read(address int64) *encoding.ByteStream {
	return this.Wordlines_[this.Index(address)].Read()
}

func (this *Mram) Write(address int64, byte_stream *encoding.ByteStream) {
	this.Wordlines_[this.Index(address)].Write(byte_stream)
}

func (this *Mram) Index(address int64) int {
	if address < this.Address_ {
		err := errors.New("address < MRAM offset")
		panic(err)
	} else if address+global.WordlineSize > this.Address_+this.Size_ {
		err := errors.New("address + wordline size > MRAM offset + MRAM size")
		panic(err)
	} else if address%global.WordlineSize != 0 {
		err := errors.New("address is not aligned with wordline size")
		panic(err)
	}

	return int((address - this.Address_) / global.WordlineSize)
}

func (m *Mram) SaveToJson(filename string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func (m *Mram) Replace(filename string) error {
	newMram, err := LoadFromJson(filename)
	if err != nil {
		return err
	}
	*m = *newMram
	return nil
}

func LoadFromJson(filename string) (*Mram, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var mram Mram
	err = json.Unmarshal(data, &mram)
	if err != nil {
		return nil, err
	}
	return &mram, nil
}
