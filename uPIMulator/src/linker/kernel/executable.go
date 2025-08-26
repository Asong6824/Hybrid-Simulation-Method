package kernel

import (
	"errors"
	"fmt"
	"sort"
	"uPIMulator/src/abi/encoding"
	"uPIMulator/src/linker/lexer"
	"uPIMulator/src/linker/parser"
	"uPIMulator/src/misc"

	"github.com/elliotchance/orderedmap"
)

type Executable struct {
	name string
	path string

	benchmark_relocatable *Relocatable
	sdk_relocatables      *orderedmap.OrderedMap

	token_stream *lexer.TokenStream
	ast          *parser.Ast
	liveness     *Liveness

	sections    *orderedmap.OrderedMap
	cur_section *Section
}

func (this *Executable) Init(name string) {
	this.name = name

	this.sdk_relocatables = orderedmap.NewOrderedMap()

	this.liveness = new(Liveness)
	this.liveness.Init()

	this.sections = orderedmap.NewOrderedMap()
}

func (this *Executable) Name() string {
	return this.name
}

func (this *Executable) Path() string {
	return this.path
}

func (this *Executable) SetPath(path string) {
	this.path = path
}

func (this *Executable) SetBenchmarkRelocatable(relocatable *Relocatable) {
	this.benchmark_relocatable = relocatable

	this.UpdateUnresolvedSymbols(relocatable)
}

func (this *Executable) AddSdkRelocatable(relocatable *Relocatable) {
	this.sdk_relocatables.Set(relocatable, true)

	this.UpdateLocalSymbols(relocatable)
	this.UpdateUnresolvedSymbols(relocatable)
}

func (this *Executable) TokenStream() *lexer.TokenStream {
	return this.token_stream
}

func (this *Executable) SetTokenStream(token_stream *lexer.TokenStream) {
	this.token_stream = token_stream
}

func (this *Executable) Ast() *parser.Ast {
	return this.ast
}

func (this *Executable) SetAst(ast *parser.Ast) {
	this.ast = ast
}

func (this *Executable) Liveness() *Liveness {
	return this.liveness
}

func (this *Executable) DumpAssembly() {
	lines := this.benchmark_relocatable.Lines()

	for el := this.sdk_relocatables.Front(); el != nil; el = el.Next() {
		sdk_relocatable := el.Key.(*Relocatable)
		lines = append(lines, sdk_relocatable.Lines()...)
	}

	file_dumper := new(misc.FileDumper)
	file_dumper.Init(this.path)
	file_dumper.WriteLines(lines)
}

func (this *Executable) DumpAddresses(path string) {
	lines := make([]string, 0)

	addresses := this.Addresses()
	for el := addresses.Front(); el != nil; el = el.Next() {
		label_name := el.Key.(string)
		label_address := el.Value.(int64)
		line := fmt.Sprintf("%s: %d", label_name, label_address)
		lines = append(lines, line)
	}

	file_dumper := new(misc.FileDumper)
	file_dumper.Init(path)
	file_dumper.WriteLines(lines)
}

func (this *Executable) DumpAtomic(path string) {
	atomic_byte_stream := this.AtomicByteStream()

	lines := make([]string, 0)
	for i := int64(0); i < atomic_byte_stream.Size(); i++ {
		line := fmt.Sprintf("%d", atomic_byte_stream.Get(int(i)))

		lines = append(lines, line)
	}

	file_dumper := new(misc.FileDumper)
	file_dumper.Init(path)
	file_dumper.WriteLines(lines)
}

func (this *Executable) DumpIram(path string) {
	iram_byte_stream := this.IramByteStream()

	lines := make([]string, 0)
	for i := int64(0); i < iram_byte_stream.Size(); i++ {
		line := fmt.Sprintf("%d", iram_byte_stream.Get(int(i)))

		lines = append(lines, line)
	}

	file_dumper := new(misc.FileDumper)
	file_dumper.Init(path)
	file_dumper.WriteLines(lines)
}

func (this *Executable) DumpWram(path string) {
	wram_byte_stream := this.WramByteStream()

	lines := make([]string, 0)
	for i := int64(0); i < wram_byte_stream.Size(); i++ {
		line := fmt.Sprintf("%d", wram_byte_stream.Get(int(i)))

		lines = append(lines, line)
	}

	file_dumper := new(misc.FileDumper)
	file_dumper.Init(path)
	file_dumper.WriteLines(lines)
}

func (this *Executable) DumpMram(path string) {
	mram_byte_stream := this.MramByteStream()

	lines := make([]string, 0)
	for i := int64(0); i < mram_byte_stream.Size(); i++ {
		line := fmt.Sprintf("%d", mram_byte_stream.Get(int(i)))

		lines = append(lines, line)
	}

	file_dumper := new(misc.FileDumper)
	file_dumper.Init(path)
	file_dumper.WriteLines(lines)
}

func (this *Executable) Section(section_name SectionName, name string) *Section {
	for el := this.sections.Front(); el != nil; el = el.Next() {
		section := el.Key.(*Section)
		if section.SectionName() == section_name && section.Name() == name {
			return section
		}
	}
	return nil
}

func (this *Executable) Sections(section_name SectionName) *orderedmap.OrderedMap {
	sections := orderedmap.NewOrderedMap()

	for el := this.sections.Front(); el != nil; el = el.Next() {
		section := el.Key.(*Section)
		if section.SectionName() == section_name {
			sections.Set(section, true)
		}
	}
	return sections
}

func (this *Executable) AddSection(
	section_name SectionName,
	name string,
	section_flags map[SectionFlag]bool,
	section_type SectionType,
) {
	if this.Section(section_name, name) == nil {
		section := new(Section)
		section.Init(section_name, name, section_flags, section_type)
		this.sections.Set(section, true)
	}
}

func (this *Executable) CurSection() *Section {
	return this.cur_section
}

func (this *Executable) CheckoutSection(section_name SectionName, name string) {
	if section := this.Section(section_name, name); section != nil {
		this.cur_section = section
	} else {
		err := errors.New("section is not found")
		panic(err)
	}
}

func (this *Executable) Label(label_name string) *Label {
	var label *Label = nil
	for el := this.sections.Front(); el != nil; el = el.Next() {
		section := el.Key.(*Section)
		section_label := section.Label(label_name)

		if section_label != nil {
			if label != nil {
				err := errors.New("labels are duplicated")
				panic(err)
			}

			label = section_label
		}
	}
	return label
}

func (this *Executable) Addresses() *orderedmap.OrderedMap {
	addresses := orderedmap.NewOrderedMap()
	for el := this.sections.Front(); el != nil; el = el.Next() {
		section := el.Key.(*Section)
		for _, label := range section.Labels() {
			addresses.Set(label.Name(), label.Address())
		}
	}
	return addresses
}

func (this *Executable) AtomicByteStream() *encoding.ByteStream {
	config_loader := new(misc.ConfigLoader)
	config_loader.Init()

	atomic_sections := this.Sort(
		config_loader.AtomicOffset(),
		config_loader.AtomicOffset()+config_loader.AtomicSize(),
	)

	byte_stream := new(encoding.ByteStream)
	byte_stream.Init()

	for _, atomic_section := range atomic_sections {
		byte_stream.Merge(atomic_section.ToByteStream())
	}

	return byte_stream
}

func (this *Executable) IramByteStream() *encoding.ByteStream {
	config_loader := new(misc.ConfigLoader)
	config_loader.Init()

	iram_sections := this.Sort(
		config_loader.IramOffset(),
		config_loader.IramOffset()+config_loader.IramSize(),
	)

	byte_stream := new(encoding.ByteStream)
	byte_stream.Init()

	for _, iram_section := range iram_sections {
		byte_stream.Merge(iram_section.ToByteStream())
	}

	return byte_stream
}

func (this *Executable) WramByteStream() *encoding.ByteStream {
	config_loader := new(misc.ConfigLoader)
	config_loader.Init()

	wram_sections := this.Sort(
		config_loader.WramOffset(),
		config_loader.WramOffset()+config_loader.WramSize(),
	)

	byte_stream := new(encoding.ByteStream)
	byte_stream.Init()

	for _, wram_section := range wram_sections {
		byte_stream.Merge(wram_section.ToByteStream())
	}

	return byte_stream
}

func (this *Executable) MramByteStream() *encoding.ByteStream {
	config_loader := new(misc.ConfigLoader)
	config_loader.Init()

	mram_sections := this.Sort(
		config_loader.MramOffset(),
		config_loader.MramOffset()+config_loader.MramSize(),
	)

	byte_stream := new(encoding.ByteStream)
	byte_stream.Init()

	for _, mram_section := range mram_sections {
		byte_stream.MergeMemoryBlocks(mram_section.ToByteStream(), mram_section.Address())
	}

	return byte_stream
}

func (this *Executable) UpdateLocalSymbols(relocatable *Relocatable) {
	for el := relocatable.Liveness().LocalSymbols().Front(); el != nil; el = el.Next() {
		old_name := el.Key.(string)
		new_name := relocatable.Name() + "." + old_name

		relocatable.RenameLocalSymbol(old_name, new_name)
	}
}
func (this *Executable) UpdateUnresolvedSymbols(relocatable *Relocatable) {
	// 遍历 Defs
	for el := relocatable.Liveness().Defs().Front(); el != nil; el = el.Next() {
		def := el.Key.(string)
		this.liveness.AddDef(def)
	}

	// 遍历 Uses
	for el := relocatable.Liveness().Uses().Front(); el != nil; el = el.Next() {
		use := el.Key.(string)
		this.liveness.AddUse(use)
	}

	// 遍历 GlobalSymbols
	for el := relocatable.Liveness().GlobalSymbols().Front(); el != nil; el = el.Next() {
		global_symbol := el.Key.(string)
		this.liveness.AddGlobalSymbol(global_symbol)
	}
}

func (this *Executable) Sort(begin_address int64, end_address int64) []*Section {
	sections := make([]*Section, 0)

	for el := this.sections.Front(); el != nil; el = el.Next() {
		section := el.Key.(*Section)
		address := section.Address()

		if begin_address <= address && address < end_address {
			sections = append(sections, section)
		}
	}

	sort_fn := func(i int, j int) bool {
		return sections[i].Address() < sections[j].Address()
	}

	sort.Slice(sections, sort_fn)

	return sections
}
