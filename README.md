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
- **PrIM Benchmarks Suite**: Reference implementation for code structure (see limitations below)

### Important Note on PrIM Benchmarks

**PrIM benchmarks** were originally proposed in early UPMEM architecture research and have become widely used as a standard test suite in the research community. **uPIMulator**, as an unofficial cycle-accurate simulator, includes PrIM benchmark code for validation and accuracy verification.

However, **our hybrid simulation method has a technical limitation**: it does not support host-to-DPU data transfer during the simulation process. Therefore:

- The PrIM benchmarks  in this repository serve as **code reference only**
- They demonstrate DPU program structure and can be used as templates for developing your own programs
- **They are not intended for direct execution** in the hybrid simulation workflow
- Users should develop their own DPU programs or adapt existing ones to work within the hybrid simulation constraints

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
   
   **Why this modification is needed**: uPIMulator allocates debug information sections before the `.mram` section, causing address misalignment. This reserved space ensures both simulators use identical memory layouts.
   
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
- Use the PrIM benchmarks as reference for code writing, compilation, and execution patterns

**Example DPU Program Structure**:
```c
// dpu/dpu_program.c
#include <stdint.h>
#include <stdio.h>
#include <defs.h>
#include <mram.h>
#include <alloc.h>
#include <perfcounter.h>
#include <barrier.h>

// Initialize data in MRAM (no host transfer)
__mram uint32_t input_data[DATA_SIZE];
__mram uint32_t output_data[DATA_SIZE];

int main() {
    // Initialize data directly in DPU
    // Perform computation
    // Write results to MRAM
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
   
   **Note**: For Makefile structure and build configuration, you can reference the Makefiles in the PrIM benchmarks (e.g., `prim-benchmarks/VA/Makefile`, `prim-benchmarks/GEMV/Makefile`) to understand the proper compilation flags, linking options, and build targets for DPU programs.

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
   # Set benchmark to match your choice (e.g., VA)
   sed -i 's/BENCHMARK=.*/BENCHMARK="VA"/' run_uPIMulator.sh
   ```

2. **Generate debug information**:
   ```bash
   ./run_uPIMulator.sh
   # This creates bin/mram.bin with debug sections
   ```

### Step 4: Memory State Migration

**Purpose**: Extract the actual used portions of `.mram` and `.mram_noinit` sections from UPMEM SDK simulator's MRAM image and replace the corresponding parts in uPIMulator's `mram.bin`.

**Key Points**:
- Only extract the **actually used portions** of `.mram` and `.mram_noinit` sections for performance optimization
- The specific content to extract depends on your program's memory usage and should be determined by the user
- This approach avoids copying the entire MRAM image, focusing only on the relevant data sections

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
