#!/bin/bash

# set path
ROOT_DIRPATH="/home/asong/桌面/uPIMulator/golang/uPIMulator"
BIN_DIRPATH="/home/asong/桌面/uPIMulator/golang/uPIMulator/bin"
IMAGE_DIRPATH="/home/asong/桌面/uPIMulator/golang/uPIMulator/image"

# set benchmark type and parameter
VERBOSE=1
BENCHMARK="VA"
NUM_CHANNELS=1
NUM_RANKS_PER_CHANNEL=1
NUM_DPUS_PER_RANK=1
NUM_TASKLETS=1
DATA_PREP_PARAMS=16
LOAD_LOCAL=0

mkdir "${BIN_DIRPATH}"
mkdir "${IMAGE_DIRPATH}"

# execute uPIMulator
./build/uPIMulator --verbose $VERBOSE \
                   --root_dirpath $ROOT_DIRPATH \
                   --bin_dirpath $BIN_DIRPATH \
                   --image_dirpath $IMAGE_DIRPATH \
                   --benchmark $BENCHMARK \
                   --num_channels $NUM_CHANNELS \
                   --num_ranks_per_channel $NUM_RANKS_PER_CHANNEL \
                   --num_dpus_per_rank $NUM_DPUS_PER_RANK \
                   --num_tasklets $NUM_TASKLETS \
                   --load_local $LOAD_LOCAL \
                   --data_prep_params $DATA_PREP_PARAMS
