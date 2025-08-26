package host

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"uPIMulator/src/abi/encoding"
	"uPIMulator/src/core"
	"uPIMulator/src/global"
	"uPIMulator/src/misc"
	"uPIMulator/src/simulator/channel"
	"uPIMulator/src/simulator/dpu"
)

type Host struct {
	addresses map[string]int64
	values    map[string]int64

	atomic *encoding.ByteStream
	iram   *encoding.ByteStream
	wram   *encoding.ByteStream
	mram   *encoding.ByteStream

	num_executions int

	input_dpu_host  []*Chunk
	output_dpu_host []*Chunk

	input_dpu_mram_heap_pointer_name  []*Chunk
	output_dpu_mram_heap_pointer_name []*Chunk

	channels []*channel.Channel
}

func (this *Host) Init() {

	this.channels = make([]*channel.Channel, 0)

	this.InitAddresses()
	this.InitValues()
	this.InitAtomic()
	this.InitIram()
	this.InitWram()
	this.InitMram()
	this.InitNumExecutions()
	this.InitChunks()
}

func (this *Host) InitAddresses() {

	var path string
	path = filepath.Join(global.BinDirpath, "addresses.txt")

	file_scanner := new(misc.FileScanner)
	file_scanner.Init(path)

	lines := file_scanner.ReadLines()

	this.addresses = make(map[string]int64, 0)

	for _, line := range lines {
		words := strings.Split(line, ":")

		name := words[0]
		address, err := strconv.ParseInt(words[1][1:], 10, 64)

		if err != nil {
			panic(err)
		}

		this.addresses[name] = address
	}
}

func (this *Host) InitValues() {
	path := filepath.Join(global.BinDirpath, "values.txt")

	file_scanner := new(misc.FileScanner)
	file_scanner.Init(path)

	lines := file_scanner.ReadLines()

	this.values = make(map[string]int64, 0)

	for _, line := range lines {
		words := strings.Split(line, ":")

		name := words[0]
		address, err := strconv.ParseInt(words[1][1:], 10, 64)

		if err != nil {
			panic(err)
		}

		this.values[name] = address
	}
}

func (this *Host) InitAtomic() {
	path := filepath.Join(global.BinDirpath, "atomic.bin")

	this.atomic = this.InitByteStream(path)
}

func (this *Host) InitIram() {
	path := filepath.Join(global.BinDirpath, "iram.bin")

	this.iram = this.InitByteStream(path)
}

func (this *Host) InitWram() {
	path := filepath.Join(global.BinDirpath, "wram.bin")

	this.wram = this.InitByteStream(path)
}

func (this *Host) InitMram() {
	var path string
	if global.LoadLocal == 0 {
		path = filepath.Join(global.BinDirpath, "mram.bin")
	} else {
		path = filepath.Join(global.ImageDirpath, "mram.bin")
	}

	this.mram = this.InitByteStream(path)
}

func (this *Host) InitNumExecutions() {
	path := filepath.Join(global.BinDirpath, "num_executions.txt")

	file_scanner := new(misc.FileScanner)
	file_scanner.Init(path)

	lines := file_scanner.ReadLines()

	if len(lines) != 1 {
		err := errors.New("lines' length != 1")
		panic(err)
	}

	var err error
	this.num_executions, err = strconv.Atoi(lines[0])

	if err != nil {
		panic(err)
	}
}

func (this *Host) InitChunks() {
	entries, bin_dir_read_err := os.ReadDir(global.BinDirpath)

	if bin_dir_read_err != nil {
		panic(bin_dir_read_err)
	}

	this.input_dpu_host = make([]*Chunk, 0)
	this.output_dpu_host = make([]*Chunk, 0)
	this.input_dpu_mram_heap_pointer_name = make([]*Chunk, 0)
	this.output_dpu_mram_heap_pointer_name = make([]*Chunk, 0)

	for _, entry := range entries {
		filename := entry.Name()

		words := strings.Split(strings.Split(filename, ".")[0], "_")

		if words[0] == "input" || words[0] == "output" {
			byte_stream := this.InitByteStream(filepath.Join(global.BinDirpath, filename))

			chunk := new(Chunk)
			chunk.Init(filename, byte_stream)

			if chunk.ChunkType() == INPUT_DPU_HOST {
				this.input_dpu_host = append(this.input_dpu_host, chunk)
			} else if chunk.ChunkType() == OUTPUT_DPU_HOST {
				this.output_dpu_host = append(this.output_dpu_host, chunk)
			} else if chunk.ChunkType() == INPUT_DPU_MRAM_HEAP_POINTER_NAME {
				this.input_dpu_mram_heap_pointer_name = append(this.input_dpu_mram_heap_pointer_name, chunk)
			} else if chunk.ChunkType() == OUTPUT_DPU_MRAM_HEAP_POINTER_NAME {
				this.output_dpu_mram_heap_pointer_name = append(this.output_dpu_mram_heap_pointer_name, chunk)
			} else {
				chunk_type_err := errors.New("chunk type is not valid")
				panic(chunk_type_err)
			}
		}
	}
}

func (this *Host) InitByteStream(path string) *encoding.ByteStream {
	file_scanner := new(misc.FileScanner)
	file_scanner.Init(path)
	lines := file_scanner.ReadLines()

	byte_stream := new(encoding.ByteStream)
	byte_stream.Init()

	for _, line := range lines {
		value, err := strconv.Atoi(line)

		if err != nil {
			panic(err)
		}

		byte_stream.Append(uint8(value))
	}

	return byte_stream
}

func (this *Host) Fini() {
}

func (this *Host) ConnectChannels(channels []*channel.Channel) {
	this.channels = channels
}

func (this *Host) NumExecutions() int {
	return this.num_executions
}

func (this *Host) Dpus() []*dpu.Dpu {
	dpus := make([]*dpu.Dpu, 0)

	for _, channel_ := range this.channels {
		dpus = append(dpus, channel_.Dpus()...)
	}

	return dpus
}

func (this *Host) IsZombie() bool {
	dpus := this.Dpus()

	for _, dpu_ := range dpus {
		if !dpu_.IsZombie() {
			return false
		}
	}

	return true
}

func (this *Host) Load() {
	this.DmaTransferToAtomic()
	this.DmaTransferToIram()
	this.DmaTransferToWram()
	this.DmaTransferToMram()

}

func (this *Host) Schedule(execution int) {
	// TODO(bongjoon.hyun@gmail.com): fix this
	if global.Benchmark == "TRNS" {
		this.Load()
	}

	this.ChannelTransferInputDpuHost(execution)
	this.ChannelTransferInputDpuMramHeapPointerName(execution)
}

func (this *Host) Check(execution int) {
	this.ChannelTransferOutputDpuHost(execution)
	this.ChannelTransferOutputDpuMramHeapPointerName(execution)
}

func (this *Host) Launch() {
	config_loader := new(misc.ConfigLoader)
	config_loader.Init()

	dpus := this.Dpus()

	for _, dpu_ := range dpus {
		threads := dpu_.Threads()

		for _, thread := range threads {
			bootstrap := config_loader.IramOffset()
			thread.RegFile().WritePcReg(bootstrap)
		}

		dpu_.Boot()
		// if global.LoadLocal == 1 {
		// 	dpu_.Replace()
		// }

	}
}

func (this *Host) DmaTransferToAtomic() {
	dpus := this.Dpus()

	thread_pool := new(core.ThreadPool)
	thread_pool.Init(global.NumSimulationtThreads)

	for _, dpu_ := range dpus {
		dma_transfer_to_atomic_job := new(DmaTransferToAtomicJob)
		dma_transfer_to_atomic_job.Init(this.atomic, dpu_)

		thread_pool.Enque(dma_transfer_to_atomic_job)
	}

	thread_pool.Start()
}

func (this *Host) DmaTransferToIram() {
	dpus := this.Dpus()

	thread_pool := new(core.ThreadPool)
	thread_pool.Init(global.NumSimulationtThreads)

	for _, dpu_ := range dpus {
		dma_transfer_to_iram_job := new(DmaTransferToIramJob)
		dma_transfer_to_iram_job.Init(this.iram, dpu_)

		thread_pool.Enque(dma_transfer_to_iram_job)
	}

	thread_pool.Start()
}

func (this *Host) DmaTransferToWram() {
	dpus := this.Dpus()

	thread_pool := new(core.ThreadPool)
	thread_pool.Init(global.NumSimulationtThreads)

	for _, dpu_ := range dpus {
		dma_transfer_to_wram_job := new(DmaTransferToWramJob)
		dma_transfer_to_wram_job.Init(this.wram, dpu_)

		thread_pool.Enque(dma_transfer_to_wram_job)
	}

	thread_pool.Start()
}

func (this *Host) DmaTransferToMram() {
	dpus := this.Dpus()

	thread_pool := new(core.ThreadPool)
	thread_pool.Init(global.NumSimulationtThreads)

	for _, dpu_ := range dpus {
		dma_transfer_to_mram_job := new(DmaTransferToMramJob)
		dma_transfer_to_mram_job.Init(this.mram, dpu_)

		thread_pool.Enque(dma_transfer_to_mram_job)
	}

	thread_pool.Start()
}

func (this *Host) ChannelTransferInputDpuHost(execution int) {
	thread_pool := new(core.ThreadPool)
	thread_pool.Init(global.NumSimulationtThreads)

	pointers := this.FindInputDpuHostPointers(execution)

	for pointer, _ := range pointers {
		if _, found := this.addresses[pointer]; !found {
			err := errors.New("pointer is not found")
			panic(err)
		}

		address := this.addresses[pointer]

		for _, channel_ := range this.channels {
			channel_id := channel_.ChannelId()
			ranks := channel_.Ranks()

			for _, rank_ := range ranks {
				rank_id := rank_.RankId()
				dpus := rank_.Dpus()

				for i := 0; i < 8; i++ {
					dpu_ids := make([]int, 0)
					byte_streams := make([]*encoding.ByteStream, 0)

					for _, dpu_ := range dpus {
						dpu_id := dpu_.DpuId()
						unique_dpu_id := channel_id*global.NumRanksPerChannel*global.NumDpusPerRank + rank_id*global.NumDpusPerRank + dpu_id
						if dpu_id%8 == i {
							chunk := this.FindInputDpuHostChunk(pointer, execution, unique_dpu_id)

							dpu_ids = append(dpu_ids, dpu_id)
							byte_streams = append(byte_streams, chunk.ByteStream())
						}
					}

					if len(byte_streams) != 0 && byte_streams[0].Size() != 0 {
						channel_message := new(channel.ChannelMessage)
						channel_message.InitWrite(
							channel_.ChannelId(),
							rank_.RankId(),
							dpu_ids,
							address,
							byte_streams[0].Size(),
							byte_streams,
						)

						channel_transfer_write_job := new(ChannelTransferWriteJob)
						channel_transfer_write_job.Init(channel_message, channel_)

						thread_pool.Enque(channel_transfer_write_job)
					}
				}
			}
		}
	}

	thread_pool.Start()
}

func (this *Host) ChannelTransferOutputDpuHost(execution int) {
	thread_pool := new(core.ThreadPool)
	thread_pool.Init(global.NumSimulationtThreads)

	pointers := this.FindOutputDpuHostPointers(execution)

	for pointer, _ := range pointers {
		if _, found := this.addresses[pointer]; !found {
			err := errors.New("pointer is not found")
			panic(err)
		}

		address := this.addresses[pointer]

		for _, channel_ := range this.channels {
			channel_id := channel_.ChannelId()
			ranks := channel_.Ranks()

			for _, rank_ := range ranks {
				rank_id := rank_.RankId()
				dpus := rank_.Dpus()

				for i := 0; i < 8; i++ {
					dpu_ids := make([]int, 0)
					byte_streams := make([]*encoding.ByteStream, 0)

					for _, dpu_ := range dpus {
						dpu_id := dpu_.DpuId()
						unique_dpu_id := channel_id*global.NumRanksPerChannel*global.NumDpusPerRank + rank_id*global.NumDpusPerRank + dpu_id

						if dpu_id%8 == i {
							chunk := this.FindOutputDpuHostChunk(pointer, execution, unique_dpu_id)

							dpu_ids = append(dpu_ids, dpu_id)
							byte_streams = append(byte_streams, chunk.ByteStream())
						}
					}

					if len(byte_streams) != 0 && byte_streams[0].Size() != 0 {
						channel_message := new(channel.ChannelMessage)
						channel_message.InitRead(
							channel_.ChannelId(),
							rank_.RankId(),
							dpu_ids,
							address,
							byte_streams[0].Size(),
						)

						channel_transfer_read_job := new(ChannelTransferReadJob)
						channel_transfer_read_job.Init(channel_message, byte_streams, channel_)

						thread_pool.Enque(channel_transfer_read_job)
					}
				}
			}
		}
	}

	thread_pool.Start()
}

func (this *Host) ChannelTransferInputDpuMramHeapPointerName(execution int) {
	thread_pool := new(core.ThreadPool)
	thread_pool.Init(global.NumSimulationtThreads)

	if _, found := this.values["__sys_used_mram_end"]; !found {
		err := errors.New("__sys_used_mram_end is not found")
		panic(err)
	}

	sys_used_mram_end := this.values["__sys_used_mram_end"]

	offsets := this.FindInputDpuMramHeapPointerNameOffsets(execution)

	for offset, _ := range offsets {
		address := sys_used_mram_end + offset

		for _, channel_ := range this.channels {
			channel_id := channel_.ChannelId()
			ranks := channel_.Ranks()

			for _, rank_ := range ranks {
				rank_id := rank_.RankId()
				dpus := rank_.Dpus()

				for i := 0; i < 8; i++ {
					dpu_ids := make([]int, 0)
					byte_streams := make([]*encoding.ByteStream, 0)

					for _, dpu_ := range dpus {
						dpu_id := dpu_.DpuId()
						unique_dpu_id := channel_id*global.NumRanksPerChannel*global.NumDpusPerRank + rank_id*global.NumDpusPerRank + dpu_id

						if dpu_id%8 == i {
							chunk := this.FindInputDpuMramHeapPointerNameChunk(
								offset,
								execution,
								unique_dpu_id,
							)

							dpu_ids = append(dpu_ids, dpu_id)
							byte_streams = append(byte_streams, chunk.ByteStream())
						}
					}

					if len(byte_streams) != 0 && byte_streams[0].Size() != 0 {
						channel_message := new(channel.ChannelMessage)
						channel_message.InitWrite(
							channel_.ChannelId(),
							rank_.RankId(),
							dpu_ids,
							address,
							byte_streams[0].Size(),
							byte_streams,
						)

						channel_transfer_write_job := new(ChannelTransferWriteJob)
						channel_transfer_write_job.Init(channel_message, channel_)

						thread_pool.Enque(channel_transfer_write_job)
					}
				}
			}
		}
	}

	thread_pool.Start()
}

func (this *Host) ChannelTransferOutputDpuMramHeapPointerName(execution int) {
	thread_pool := new(core.ThreadPool)
	thread_pool.Init(global.NumSimulationtThreads)

	if _, found := this.values["__sys_used_mram_end"]; !found {
		err := errors.New("__sys_used_mram_end is not found")
		panic(err)
	}

	sys_used_mram_end := this.values["__sys_used_mram_end"]

	offsets := this.FindOutputDpuMramHeapPointerNameOffsets(execution)

	for offset, _ := range offsets {
		address := sys_used_mram_end + offset

		for _, channel_ := range this.channels {
			channel_id := channel_.ChannelId()
			ranks := channel_.Ranks()

			for _, rank_ := range ranks {
				rank_id := rank_.RankId()
				dpus := rank_.Dpus()

				for i := 0; i < 8; i++ {
					dpu_ids := make([]int, 0)
					byte_streams := make([]*encoding.ByteStream, 0)

					for _, dpu_ := range dpus {
						dpu_id := dpu_.DpuId()
						unique_dpu_id := channel_id*global.NumRanksPerChannel*global.NumDpusPerRank + rank_id*global.NumDpusPerRank + dpu_id

						if dpu_id%8 == i {
							chunk := this.FindOutputDpuMramHeapPointerNameChunk(
								offset,
								execution,
								unique_dpu_id,
							)

							dpu_ids = append(dpu_ids, dpu_id)
							byte_streams = append(byte_streams, chunk.ByteStream())
						}
					}

					if len(byte_streams) != 0 && byte_streams[0].Size() != 0 {
						channel_message := new(channel.ChannelMessage)
						channel_message.InitRead(
							channel_.ChannelId(),
							rank_.RankId(),
							dpu_ids,
							address,
							byte_streams[0].Size(),
						)

						channel_transfer_read_job := new(ChannelTransferReadJob)
						channel_transfer_read_job.Init(channel_message, byte_streams, channel_)

						thread_pool.Enque(channel_transfer_read_job)
					}
				}
			}
		}

		thread_pool.Start()
	}
}

func (this *Host) FindInputDpuHostPointers(execution int) map[string]bool {
	pointers := make(map[string]bool, 0)

	for _, chunk := range this.input_dpu_host {
		if chunk.Execution() == execution {
			pointers[chunk.Name()] = true
		}
	}

	return pointers
}

func (this *Host) FindOutputDpuHostPointers(execution int) map[string]bool {
	pointers := make(map[string]bool, 0)

	for _, chunk := range this.output_dpu_host {
		if chunk.Execution() == execution {
			pointers[chunk.Name()] = true
		}
	}

	return pointers
}

func (this *Host) FindInputDpuMramHeapPointerNameOffsets(execution int) map[int64]bool {
	offsets := make(map[int64]bool, 0)

	for _, chunk := range this.input_dpu_mram_heap_pointer_name {
		if chunk.Execution() == execution {
			offsets[chunk.Offset()] = true
		}
	}

	return offsets
}

func (this *Host) FindOutputDpuMramHeapPointerNameOffsets(execution int) map[int64]bool {
	offsets := make(map[int64]bool, 0)

	for _, chunk := range this.output_dpu_mram_heap_pointer_name {
		if chunk.Execution() == execution {
			offsets[chunk.Offset()] = true
		}
	}

	return offsets
}

func (this *Host) FindInputDpuHostChunk(pointer string, execution int, dpu_id int) *Chunk {
	for _, chunk := range this.input_dpu_host {
		if chunk.Name() == pointer && chunk.Execution() == execution && chunk.DpuId() == dpu_id {
			return chunk
		}
	}

	err := errors.New("chunk is not found")
	panic(err)
}

func (this *Host) FindOutputDpuHostChunk(pointer string, execution int, dpu_id int) *Chunk {
	for _, chunk := range this.output_dpu_host {
		if chunk.Name() == pointer && chunk.Execution() == execution && chunk.DpuId() == dpu_id {
			return chunk
		}
	}

	err := errors.New("chunk is not found")
	panic(err)
}

func (this *Host) FindInputDpuMramHeapPointerNameChunk(
	offset int64,
	execution int,
	dpu_id int,
) *Chunk {
	for _, chunk := range this.input_dpu_mram_heap_pointer_name {
		if chunk.Offset() == offset && chunk.Execution() == execution && chunk.DpuId() == dpu_id {
			return chunk
		}
	}

	err := errors.New("chunk is not found")
	panic(err)
}

func (this *Host) FindOutputDpuMramHeapPointerNameChunk(
	offset int64,
	execution int,
	dpu_id int,
) *Chunk {
	for _, chunk := range this.output_dpu_mram_heap_pointer_name {
		if chunk.Offset() == offset && chunk.Execution() == execution && chunk.DpuId() == dpu_id {
			return chunk
		}
	}

	err := errors.New("chunk is not found")
	panic(err)
}

func (this *Host) Cycle() {
	if _, found := this.addresses["__sys_end"]; !found {
		err := errors.New("__sys_end is not found")
		panic(err)
	}

	sys_end := this.addresses["__sys_end"]

	thread_pool := new(core.ThreadPool)
	thread_pool.Init(global.NumSimulationtThreads)

	dpus := this.Dpus()

	for _, dpu_ := range dpus {
		cycle_job := new(CycleJob)
		cycle_job.Init(sys_end, dpu_)

		thread_pool.Enque(cycle_job)
	}

	thread_pool.Start()
}
