package kernel

import (
	"errors"
	"uPIMulator/src/abi/encoding"
)

type SectionName int

const (
	ATOMIC SectionName = iota
	BSS
	DATA
	DEBUG_ABBREV
	DEBUG_FRAME
	DEBUG_INFO
	DEBUG_LINE
	DEBUG_LOC
	DEBUG_RANGES
	DEBUG_STR
	DPU_HOST
	MRAM
	RODATA
	STACK_SIZES
	TEXT
)

type SectionFlag int

const (
	ALLOC SectionFlag = iota
	WRITE
	EXECINSTR
	LINK_ORDER
	MERGE
	STRINGS
)

type SectionType int

const (
	PROGBITS SectionType = iota
	NOBITS
)

// mramSectionTypes holds the set of section names that are allocated in MRAM
// and require 8-byte alignment between their internal labels.
var mramSectionTypes = map[SectionName]bool{
	DEBUG_ABBREV: true,
	DEBUG_FRAME:  true,
	DEBUG_INFO:   true,
	DEBUG_LINE:   true,
	DEBUG_LOC:    true,
	DEBUG_RANGES: true,
	DEBUG_STR:    true,
	STACK_SIZES:  true,
	MRAM:         true,
}

type Section struct {
	section_name  SectionName
	name          string
	section_flags map[SectionFlag]bool
	section_type  SectionType

	labels    []*Label
	cur_label *Label
}

func (this *Section) Init(
	section_name SectionName,
	name string,
	section_flags map[SectionFlag]bool,
	section_type SectionType,
) {
	this.section_name = section_name
	this.name = name
	this.section_flags = section_flags
	this.section_type = section_type

	default_label := new(Label)
	default_label.Init(this.HiddenLabelName())

	this.labels = make([]*Label, 0)
	this.labels = append(this.labels, default_label)

	this.cur_label = default_label
}

func (this *Section) SectionName() SectionName {
	return this.section_name
}

func (this *Section) Name() string {
	return this.name
}

func (this *Section) SectionFlags() map[SectionFlag]bool {
	return this.section_flags
}

func (this *Section) SectionType() SectionType {
	return this.section_type
}

func (this *Section) Address() int64 {
	return this.labels[0].Address()
}

func (this *Section) SetAddress(address int64) {
	cur_address := address

	// The logic here must be consistent with the Size() function.
	if _, isMramSection := mramSectionTypes[this.section_name]; isMramSection {
		const alignment int64 = 8
		for _, label := range this.labels {
			label.SetAddress(cur_address)
			cur_address += label.Size()

			// Align the address for the next label to an 8-byte boundary.
			cur_address = (cur_address + alignment - 1) &^ (alignment - 1)
		}
	} else {
		// For non-MRAM sections, labels are placed contiguously.
		for _, label := range this.labels {
			label.SetAddress(cur_address)
			cur_address += label.Size()
		}
	}
}

func (this *Section) Size() int64 {
	// The logic here must match SetAddress to ensure the linker calculates
	// the correct section sizes before placing them in memory.
	if _, isMramSection := mramSectionTypes[this.section_name]; isMramSection {
		offset := int64(0)
		const alignment int64 = 8
		for _, label := range this.labels {
			offset += label.Size()
			// Align the offset for the next label to an 8-byte boundary.
			offset = (offset + alignment - 1) &^ (alignment - 1)
		}
		return offset
	}

	// For non-MRAM sections, the size is a simple sum of label sizes.
	size := int64(0)
	for _, label := range this.labels {
		size += label.Size()
	}
	return size
}

func (this *Section) Label(label_name string) *Label {
	for _, label := range this.labels {
		if label.Name() == label_name {
			return label
		}
	}

	return nil
}

func (this *Section) Labels() []*Label {
	return this.labels
}

func (this *Section) AppendLabel(label_name string) {
	label := new(Label)
	label.Init(label_name)

	this.labels = append(this.labels, label)
}

func (this *Section) CheckoutLabel(label_name string) {
	if this.Label(label_name) == nil {
		err := errors.New("label is not found")
		panic(err)
	}

	this.cur_label = this.Label(label_name)
}

func (this *Section) CurLabel() *Label {
	return this.cur_label
}

func (this *Section) ToByteStream() *encoding.ByteStream {
	byte_stream := new(encoding.ByteStream)
	byte_stream.Init()

	if _, isMramSection := mramSectionTypes[this.section_name]; isMramSection {
		const alignment int64 = 8
		offset := int64(0)
		
		for _, label := range this.labels {
			// Add padding to align to 8-byte boundary if needed
			padding_needed := (alignment - (offset % alignment)) % alignment
			for i := int64(0); i < padding_needed; i++ {
				byte_stream.Append(0) // Add zero padding
			}
			offset += padding_needed
			
			// Merge the label's byte stream
			byte_stream.Merge(label.ToByteStream())
			offset += label.Size()
		}
	} else {
		// For non-MRAM sections, labels are placed contiguously.
		for _, label := range this.labels {
			byte_stream.Merge(label.ToByteStream())
		}
	}

	return byte_stream
}

func (this *Section) HiddenLabelName() string {
	if this.section_name == ATOMIC {
		return "atomic." + this.name
	} else if this.section_name == BSS {
		return "bss." + this.name
	} else if this.section_name == DATA {
		return "data." + this.name
	} else if this.section_name == DEBUG_ABBREV {
		return "debug_abbrev." + this.name
	} else if this.section_name == DEBUG_FRAME {
		return "debug_frame." + this.name
	} else if this.section_name == DEBUG_INFO {
		return "debug_info." + this.name
	} else if this.section_name == DEBUG_LINE {
		return "debug_line." + this.name
	} else if this.section_name == DEBUG_LOC {
		return "debug_loc." + this.name
	} else if this.section_name == DEBUG_RANGES {
		return "debug_ranges." + this.name
	} else if this.section_name == DEBUG_STR {
		return "debug_str." + this.name
	} else if this.section_name == DPU_HOST {
		return "dpu_host." + this.name
	} else if this.section_name == MRAM {
		return "mram." + this.name
	} else if this.section_name == RODATA {
		return "rodata." + this.name
	} else if this.section_name == STACK_SIZES {
		return "stack_sizes." + this.name
	} else if this.section_name == TEXT {
		return "text." + this.name
	} else {
		err := errors.New("section name is not valid")
		panic(err)
	}
}
