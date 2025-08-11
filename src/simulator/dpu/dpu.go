package dpu

import (
	"errors"
	"fmt"
	"uPIMulator/src/global"
	"uPIMulator/src/misc"
	"uPIMulator/src/simulator/dpu/dram"
	"uPIMulator/src/simulator/dpu/logic"
	"uPIMulator/src/simulator/dpu/sram"
)

type Dpu struct {
	channel_id int
	rank_id    int
	dpu_id     int

	cycles int64

	threads           []*logic.Thread
	thread_scheduler  *logic.ThreadScheduler
	atomic            *sram.Atomic
	iram              *sram.Iram
	wram              *sram.Wram
	mram              *dram.Mram
	operand_collector *logic.OperandCollector
	memory_controller *dram.MemoryController
	dma               *logic.Dma
	logic             *logic.Logic

	stat_factory *misc.StatFactory
}

func (this *Dpu) Init(
	channel_id int,
	rank_id int,
	dpu_id int,
) {
	if channel_id < 0 {
		err := errors.New("channel ID < 0")
		panic(err)
	} else if rank_id < 0 {
		err := errors.New("rank ID < 0")
		panic(err)
	} else if dpu_id < 0 {
		err := errors.New("DPU ID < 0")
		panic(err)
	}

	this.channel_id = channel_id
	this.rank_id = rank_id
	this.dpu_id = dpu_id

	this.cycles = 0

	this.threads = make([]*logic.Thread, 0)
	for i := 0; i < global.NumTasklets; i++ {
		thread := new(logic.Thread)
		thread.Init(i)
		this.threads = append(this.threads, thread)
	}

	this.thread_scheduler = new(logic.ThreadScheduler)
	this.thread_scheduler.Init(channel_id, rank_id, dpu_id, this.threads)

	this.atomic = new(sram.Atomic)
	this.atomic.Init()

	this.iram = new(sram.Iram)
	this.iram.Init()

	this.wram = new(sram.Wram)
	this.wram.Init()

	this.mram = new(dram.Mram)
	this.mram.Init()

	// if global.LoadLocal == 1 {
	// 	mFilename := fmt.Sprintf("%s/mram_image_%d_%d_%d", global.ImageDirpath, this.channel_id, this.rank_id, this.dpu_id)
	// 	err := this.mram.Replace(mFilename)
	// 	if err != nil {
	// 		panic("Fail to load mram image")
	// 	}

	// 	wFilename := fmt.Sprintf("%s/wram_image_%d_%d_%d", global.ImageDirpath, this.channel_id, this.rank_id, this.dpu_id)
	// 	err = this.wram.Replace(wFilename)
	// 	if err != nil {
	// 		panic("Fail to load wram image")
	// 	}
	// }

	this.operand_collector = new(logic.OperandCollector)
	this.operand_collector.Init()
	this.operand_collector.ConnectWram(this.wram)

	this.memory_controller = new(dram.MemoryController)
	this.memory_controller.Init(channel_id, rank_id, dpu_id)
	this.memory_controller.ConnectMram(this.mram)

	this.dma = new(logic.Dma)
	this.dma.Init()
	this.dma.ConnectAtomic(this.atomic)
	this.dma.ConnectIram(this.iram)
	this.dma.ConnectOperandCollector(this.operand_collector)
	this.dma.ConnectMemoryController(this.memory_controller)

	this.logic = new(logic.Logic)
	this.logic.Init(channel_id, rank_id, dpu_id)
	this.logic.ConnectThreadScheduler(this.thread_scheduler)
	this.logic.ConnectAtomic(this.atomic)
	this.logic.ConnectIram(this.iram)
	this.logic.ConnectOperandCollector(this.operand_collector)
	this.logic.ConnectDma(this.dma)

	name := fmt.Sprintf("DPU%d-%d-%d", channel_id, rank_id, dpu_id)
	this.stat_factory = new(misc.StatFactory)
	this.stat_factory.Init(name)
}

func (this *Dpu) Fini() {
	for _, thread := range this.threads {
		thread.Fini()
	}

	this.atomic.Fini()
	this.iram.Fini()
	this.wram.Fini()
	this.mram.Fini()

	this.operand_collector.Fini()
	this.memory_controller.Fini()

	this.logic.Fini()
	this.dma.Fini()
}

func (this *Dpu) ChannelId() int {
	return this.channel_id
}

func (this *Dpu) RankId() int {
	return this.rank_id
}

func (this *Dpu) DpuId() int {
	return this.dpu_id
}

func (this *Dpu) ThreadScheduler() *logic.ThreadScheduler {
	return this.thread_scheduler
}

func (this *Dpu) Logic() *logic.Logic {
	return this.logic
}

func (this *Dpu) MemoryController() *dram.MemoryController {
	return this.memory_controller
}

func (this *Dpu) Dma() *logic.Dma {
	return this.dma
}

func (this *Dpu) Threads() []*logic.Thread {
	return this.threads
}

func (this *Dpu) StatFactory() *misc.StatFactory {
	return this.stat_factory
}

func (this *Dpu) Boot() {
	this.thread_scheduler.Boot(0)
}

func (this *Dpu) IsZombie() bool {
	for _, thread := range this.threads {
		if thread.ThreadState() != logic.ZOMBIE {
			return false
		}
	}
	return this.logic.IsEmpty() && this.memory_controller.IsEmpty()
}

func (this *Dpu) Cycle() {
	for _, thread := range this.threads {
		thread.IncrementIssueCycle()
	}

	this.thread_scheduler.Cycle()
	this.logic.Cycle()
	this.dma.Cycle()

	num_memory_cycles := int(global.FrequencyRatio*float64(this.cycles) - global.FrequencyRatio*float64(this.cycles-1))
	for i := 0; i < num_memory_cycles; i++ {
		this.memory_controller.Cycle()
	}

	this.cycles++
	//fmt.Printf("Channel id: %d, Rank id: %d, Dpu id: %d, cycle: %d\n", this.channel_id, this.rank_id, this.dpu_id, this.cycles)
}

func (this *Dpu) SaveImage() {
	var isFirstRun string
	if global.LoadLocal == 0 {
		isFirstRun = "one"
	} else {
		isFirstRun = "two"

	}

	mFilename := fmt.Sprintf("%s/mram_image%s_%d_%d_%d", global.ImageDirpath, isFirstRun, this.channel_id, this.rank_id, this.dpu_id)
	err := this.mram.SaveToJson(mFilename)
	if err != nil {
		panic(err)
	}

	wFilename := fmt.Sprintf("%s/wram_image_%d_%d_%d", global.ImageDirpath, this.channel_id, this.rank_id, this.dpu_id)
	err = this.wram.SaveToJson(wFilename)
	if err != nil {
		panic("Fail to save mram image")
	}
}

func (this *Dpu) Replace() {
	mFilename := fmt.Sprintf("%s/mram_image_%d_%d_%d", global.ImageDirpath, this.channel_id, this.rank_id, this.dpu_id)
	err := this.mram.Replace(mFilename)
	if err != nil {
		panic("Fail to load mram image")
	}

	wFilename := fmt.Sprintf("%s/wram_image_%d_%d_%d", global.ImageDirpath, this.channel_id, this.rank_id, this.dpu_id)
	err = this.wram.Replace(wFilename)
	if err != nil {
		panic("Fail to load wram image")
	}
}
