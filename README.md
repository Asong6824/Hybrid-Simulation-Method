# Hybrid Simulation Method for PIM Programs

## Overview

This repository implements a hybrid simulation method for UPMEM programs. This simulation method can only be used to simulate programs running on the UPMEM architecture. The method addresses the limitations of existing PIM simulation tools by combining the strengths of two complementary simulators:

- **UPMEM SDK Simulator**: Provides full-application simulation capability but lacks cycle-accurate performance analysis
- **uPIMulator**: Offers cycle-accurate simulation but cannot simulate entire programs, especially host-side code

### Key Innovation

Our hybrid approach enables **accurate full-application analysis** by:
1. Using UPMEM SDK simulator for functional simulation and reaching target program states
2. Extracting memory state (MRAM contents) from the SDK simulator
3. Loading the extracted state into uPIMulator for cycle-accurate performance analysis

### Repository Components

- **Standard UPMEM SDK**: Regular UPMEM SDK package (users need to install and manually modify linker scripts)
- **Modified uPIMulator**: Updated with compatible memory layout and symbol alignment

## Prerequisites

- **Linux x86_64 system** (tested on Ubuntu 18.04+)
- **Python 3.x** for build scripts
- **Go 1.21.5+** for uPIMulator
- **Docker** 

## Installation

### Step 1: Install and Configure UPMEM SDK

1. **Extract the UPMEM SDK**:
   ```bash
   cd UPMEM-SDK/
   tar -xzf upmem-2023.2.0-Linux-x86_64.tar.gz
   export UPMEM_HOME=$(pwd)/upmem-2023.2.0-Linux-x86_64
   export PATH=$UPMEM_HOME/bin:$PATH
   ```

2. **Modify the linker script for memory layout compatibility**
   
   The hybrid simulation method requires compatible memory layouts between UPMEM SDK and uPIMulator. Edit the linker script:
   
   ```bash
   vim $UPMEM_HOME/share/upmem/include/link/dpu.lds
   ```
   
   **Before** the existing `.mram.noinit` and `.mram` sections, add this reserved section:
   
   ```ld
   .mram.reserve (NOLOAD) : {
       . = ALIGN(8);
       . += 0x80000;  /* Reserve 512KB for uPIMulator debug info */
   } > mram
   ```
   
   **Why this modification is needed**: uPIMulator allocates debug information sections before the `.mram` section, causing address misalignment between the UPMEM SDK simulator and uPIMulator. This reserved space ensures both simulators use identical memory layouts.
   
   The complete MRAM section should look like:
   ```ld
   .mram.reserve (NOLOAD) : {
       . = ALIGN(8);
       . += 0x80000; 
   } > mram

   .mram.noinit (NOLOAD) : {
       *(.mram.noinit .mram.noinit.*)
       KEEP(*(.mram.noinit.keep .mram.noinit.keep.*))
   } > mram

   .mram : {
       *(.mram .mram.*)
       KEEP(*(.mram.keep .mram.keep.*))
       . = ALIGN(8);
       __sys_used_mram_end = .;
   } > mram
   ```

### Step 2: Build Modified uPIMulator

Our modified uPIMulator includes compatibility fixes for the hybrid simulation method:
- **8-byte symbol alignment** (matching UPMEM SDK behavior)
- **Compatible memory layout** with explicit `.mram` section addressing

1. **Navigate to the uPIMulator directory**:
   ```bash
   cd uPIMulator/
   ```

2. **Build the simulator**:
   ```bash
   cd script/
   python3 build.py
   ```
   
   The build process will:
   - Compile the Go-based simulator core
   - Build benchmark programs with compatible linking
   - Create the `uPIMulator` binary in `../build/`

3. **Verify the build**:
   ```bash
   ls ../build/uPIMulator  # Should exist
   ```

## Hybrid Simulation Workflow

The hybrid simulation method follows a 4-step process that combines functional simulation with cycle-accurate analysis:

### Step 1: Develop Your DPU Program

**Important**: Due to technical limitations, the hybrid simulation method does not support host-to-DPU data transfer. You need to develop DPU programs that:

- Initialize data directly in DPU memory (WRAM/MRAM)
- Do not rely on runtime host data transfers

**Complete Example DPU Program**:

1. **Create the DPU program** (`dpu/dpu_program.c`):
```c
#include <stdint.h>
#include <stdio.h>
#include <defs.h>
#include <mram.h>
#include <alloc.h>
#include <perfcounter.h>
#include <barrier.h>

#define DATA_SIZE 1024

// Initialize data in MRAM (no host transfer)
__mram uint32_t input_data[DATA_SIZE] = {1, 2, 3, 4, 5}; // Sample initialization
__mram uint32_t output_data[DATA_SIZE];

int main() {
    // Simple vector addition example
    for (int i = 0; i < DATA_SIZE; i++) {
        output_data[i] = input_data[i] + 1;
    }
    return 0;
}
```

2. **Create a minimal Makefile**:
```makefile
DPU_DIR := dpu
HOST_DIR := host
BUILDDIR := bin

NR_TASKLETS := 16
NR_DPUS := 1

# DPU binary
DPU_TARGET := ${BUILDDIR}/dpu_code
DPU_SOURCES := $(wildcard ${DPU_DIR}/*.c)

# Host binary  
HOST_TARGET := ${BUILDDIR}/host_code
HOST_SOURCES := $(wildcard ${HOST_DIR}/*.c)

.PHONY: all clean

all: ${DPU_TARGET} ${HOST_TARGET}

${BUILDDIR}:
	mkdir -p ${BUILDDIR}

${DPU_TARGET}: ${DPU_SOURCES} | ${BUILDDIR}
	dpu-upmem-dpurte-clang ${DPU_SOURCES} -o $@ -DNR_TASKLETS=${NR_TASKLETS}

${HOST_TARGET}: ${HOST_SOURCES} | ${BUILDDIR}
	gcc ${HOST_SOURCES} -o $@ `dpu-pkg-config --cflags --libs dpu` -DNR_DPUS=${NR_DPUS} -DNR_TASKLETS=${NR_TASKLETS}

clean:
	rm -rf ${BUILDDIR}
```

3. **Create a simple host program** (`host/host_program.c`):
```c
#include <stdio.h>
#include <stdlib.h>
#include <dpu.h>

#ifndef NR_DPUS
#define NR_DPUS 1
#endif

int main() {
    struct dpu_set_t dpu_set, dpu;
    
    // Allocate DPUs
    DPU_ASSERT(dpu_alloc(NR_DPUS, NULL, &dpu_set));
    
    // Load DPU program
    DPU_ASSERT(dpu_load(dpu_set, "bin/dpu_code", NULL));
    
    // Launch DPU program
    DPU_ASSERT(dpu_launch(dpu_set, DPU_SYNCHRONOUS));
    
    printf("DPU program executed successfully\n");
    
    // Free DPUs
    DPU_ASSERT(dpu_free(dpu_set));
    
    return 0;
}
```

### Step 2: Functional Simulation with UPMEM SDK

**Purpose**: Execute the complete application to reach the desired program state and verify functional correctness.

1. **Build your DPU program**:
   ```bash
   cd your-dpu-program/
   make
   ```


2. **Run functional simulation using UPMEM SDK**:
   ```bash
   # Load the DPU program in LLDB simulator
   dpu-lldb ./bin/dpu_code
   
   # In LLDB, execute the program
   (lldb) run
   # ... program executes and reaches target state ...
   ```

3. **Extract MRAM state after execution**:
   ```bash
   # Dump MRAM contents (0x08000000 to 0x0c000000 = 64MB)
   (lldb) memory read --force --outfile mram_sdk.bin --format u --size 1 0x08000000 0x0c000000
   (lldb) quit
   ```

### Step 3: Generate uPIMulator Debug Information

**Purpose**: Create compatible debug information sections for memory state migration.

1. **Configure uPIMulator for debug info generation**:
   ```bash
   cd uPIMulator/
   # Edit run_uPIMulator.sh: set LOAD_LOCAL=0
   sed -i 's/LOAD_LOCAL=.*/LOAD_LOCAL=0/' run_uPIMulator.sh
   # Set benchmark name to match your DPU program directory name
   # Replace "your_benchmark" with the actual name of your benchmark directory
   sed -i 's/BENCHMARK=.*/BENCHMARK="your_benchmark"/' run_uPIMulator.sh
   ```

2. **Generate debug information**:
   ```bash
   ./run_uPIMulator.sh
   # This creates bin/mram.bin with debug sections
   ```

### Step 4: Memory State Migration

**Purpose**: Extract the actual used portions of `.mram` and `.mram_noinit` sections from UPMEM SDK simulator's MRAM image and replace the corresponding parts in uPIMulator's `mram.bin`.

**Key Points**:
- Only extract the **actually used portions** of `.mram` and `.mram_noinit` sections
- The specific content to extract depends on your program's memory usage and should be determined by the user
- Reference examples of memory state fragments are provided in the `mram-image-example/` directory, including sample MRAM images from both SDK simulator (`mram_sdk.bin`) and uPIMulator (`mram_upimulator.bin`)

```bash
# Example: Extract specific .mram and .mram_noinit sections from SDK MRAM dump
# Note: Adjust the offset and size based on your program's actual memory layout

# Extract .mram section (example: offset 0x80000, size determined by actual usage)
dd if=mram_sdk.bin of=mram_section.bin bs=1 skip=524288 count=<actual_mram_size>

# Extract .mram_noinit section (example: offset and size based on linker map)
dd if=mram_sdk.bin of=mram_noinit_section.bin bs=1 skip=<noinit_offset> count=<actual_noinit_size>

# Replace corresponding sections in uPIMulator's mram.bin
dd if=mram_section.bin of=uPIMulator/bin/mram.bin bs=1 seek=524288 conv=notrunc
dd if=mram_noinit_section.bin of=uPIMulator/bin/mram.bin bs=1 seek=<noinit_offset> conv=notrunc
```

**Important**: The exact offsets and sizes depend on your program's memory layout. Use tools like `objdump` or `readelf` to analyze the compiled binary and determine the precise locations of `.mram` and `.mram_noinit` sections.

### Step 5: Cycle-Accurate Analysis with uPIMulator

**Purpose**: Perform detailed performance analysis starting from the migrated state.

1. **Configure uPIMulator for state loading**:
   ```bash
   cd uPIMulator/
   # Set LOAD_LOCAL=1 to load from saved state
   sed -i 's/LOAD_LOCAL=.*/LOAD_LOCAL=1/' run_uPIMulator.sh
   ```

2. **Run cycle-accurate simulation**:
   ```bash
   ./run_uPIMulator.sh
   ```

3. **Analyze results**:
   The simulator will output detailed performance metrics including:
   - Cycle counts for different operations
   - Memory access patterns
   - Instruction-level performance data
   - Energy consumption estimates
