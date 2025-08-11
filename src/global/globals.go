package global

import "uPIMulator/src/misc"

var (
	Benchmark                   string
	Verbose                     int
	NumChannels                 int
	NumRanksPerChannel          int
	NumDpusPerRank              int
	NumTasklets                 int
	NumSimulationtThreads       int
	NumPipelineStages           int
	BinDirpath                  string
	ImageDirpath                string
	ReadBandwidth               int64
	WriteBandwidth              int64
	LogicFrequency              int64
	MemoryFrequency             int64
	FrequencyRatio              float64
	WordlineSize                int64
	MinAccessGranularity        int64
	TRas                        int64
	TRcd                        int64
	TCl                         int64
	TBl                         int64
	TRp                         int64
	NumRevolverSchedulingCycles int64
	LoadLocal                   int
)

func Init(command_line_parser *misc.CommandLineParser) {
	Benchmark = command_line_parser.StringParameter("benchmark")
	Verbose = int(command_line_parser.IntParameter("verbose"))
	NumSimulationtThreads = int(command_line_parser.IntParameter("num_simulation_threads"))
	NumChannels = int(command_line_parser.IntParameter("num_channels"))
	NumRanksPerChannel = int(command_line_parser.IntParameter("num_ranks_per_channel"))
	NumDpusPerRank = int(command_line_parser.IntParameter("num_dpus_per_rank"))
	NumTasklets = int(command_line_parser.IntParameter("num_tasklets"))
	NumPipelineStages = int(command_line_parser.IntParameter("num_pipeline_stages"))
	BinDirpath = command_line_parser.StringParameter("bin_dirpath")
	ImageDirpath = command_line_parser.StringParameter("image_dirpath")
	ReadBandwidth = command_line_parser.IntParameter("read_bandwidth")
	WriteBandwidth = command_line_parser.IntParameter("write_bandwidth")
	LogicFrequency = command_line_parser.IntParameter("logic_frequency")
	MemoryFrequency = command_line_parser.IntParameter("memory_frequency")
	FrequencyRatio = float64(MemoryFrequency) / float64(LogicFrequency)
	WordlineSize = command_line_parser.IntParameter("wordline_size")
	MinAccessGranularity = command_line_parser.IntParameter("min_access_granularity")
	TRas = command_line_parser.IntParameter("t_ras")
	TRcd = command_line_parser.IntParameter("t_rcd")
	TCl = command_line_parser.IntParameter("t_cl")
	TBl = command_line_parser.IntParameter("t_bl")
	TRp = command_line_parser.IntParameter("t_rp")
	NumRevolverSchedulingCycles = command_line_parser.IntParameter(
		"num_revolver_scheduling_cycles",
	)
	LoadLocal = int(command_line_parser.IntParameter("load_local"))

}
