package simulator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"uPIMulator/src/core"
	"uPIMulator/src/global"
	"uPIMulator/src/misc"
	"uPIMulator/src/simulator/channel"
	"uPIMulator/src/simulator/host"
)

type Simulator struct {
	host     *host.Host
	channels []*channel.Channel

	execution int
}

func (this *Simulator) Init() {
	this.host = new(host.Host)
	this.host.Init()

	this.channels = make([]*channel.Channel, 0)
	for i := 0; i < global.NumChannels; i++ {
		channel_ := new(channel.Channel)
		channel_.Init(i)

		this.channels = append(this.channels, channel_)
	}

	this.host.ConnectChannels(this.channels)

	this.execution = 0

	this.host.Load()
	this.host.Schedule(this.execution)
	this.host.Launch()
}

func (this *Simulator) Fini() {
	this.host.Fini()

	for _, channel_ := range this.channels {
		channel_.Fini()
	}
}

func (this *Simulator) IsFinished() bool {
	return this.execution == this.host.NumExecutions()
}

func (this *Simulator) Cycle() {
	this.host.Cycle()

	thread_pool := new(core.ThreadPool)
	thread_pool.Init(global.NumSimulationtThreads)

	dpus := this.host.Dpus()
	for _, dpu_ := range dpus {
		cycle_job := new(CycleJob)
		cycle_job.Init(dpu_)

		thread_pool.Enque(cycle_job)
	}

	thread_pool.Start()

	if this.host.IsZombie() {
		fmt.Printf("execution (%d) is finished...\n", this.execution)

		this.host.Check(this.execution)
		this.execution++

		if !this.IsFinished() {
			this.host.Schedule(this.execution)
			this.host.Launch()
		}
	}

	// if global.Verbose >= 1 {
	// 	fmt.Println("system is cycling...")
	// }
}

func (this *Simulator) Dump() {
	file_dumper := new(misc.FileDumper)
	file_dumper.Init(filepath.Join(global.BinDirpath, "log.txt"))

	lines := make([]string, 0)

	dpus := this.host.Dpus()
	for _, dpu_ := range dpus {
		lines = append(lines, dpu_.StatFactory().ToLines()...)
		lines = append(lines, dpu_.ThreadScheduler().StatFactory().ToLines()...)
		lines = append(lines, dpu_.Logic().StatFactory().ToLines()...)
		lines = append(lines, dpu_.Logic().CycleRule().StatFactory().ToLines()...)
		lines = append(lines, dpu_.MemoryController().StatFactory().ToLines()...)
		lines = append(lines, dpu_.MemoryController().MemoryScheduler().StatFactory().ToLines()...)
		lines = append(lines, dpu_.MemoryController().RowBuffer().StatFactory().ToLines()...)
		dpu_.SaveImage()
	}

	file_dumper.WriteLines(lines)

	CopyWramBin()

}

func CopyWramBin() {
	src := filepath.Join(global.BinDirpath, "wram.bin")

	// Destination file path
	dst := filepath.Join(global.ImageDirpath, "wram.bin")

	// Make sure the destination directory exists
	err := os.MkdirAll(global.ImageDirpath, os.ModePerm)
	if err != nil {
		fmt.Printf("Failed to create destination directory: %v\n", err)
		return
	}

	sourceFile, err := os.Open(src)
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)

}
