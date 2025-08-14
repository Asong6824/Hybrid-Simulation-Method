# Installation

## Install UPMEM SDK

- Download and install the UPMEM SDK from the [official website](http://sdk-releases.upmem.com/2023.2.0/ubuntu_18.04/upmem-2023.2.0-Linux-x86_64.tar.gz).

- Modify the UPMEM SDK linker script `share/upmem/include/link/dpu.lds`

    Before the existing `.mram.noinit` and `.mram` sections, add the following section:

    ```
    .mram.reserve (NOLOAD) : {
        . = ALIGN(8);
        . += 0x80000; 
    } > mram
    ```

    The modified MRAM-related sections should look like this:
    ```
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

---

## Install uPIMulator

### Prerequisites
- **Go 1.21.5+**
- **Python 3.x**
- **Docker** (user added to docker group)


---

### Build

1. **Navigate to the build script directory**:
    ```bash
    cd /path/to/Hybrid-Simulation-Method/script
    ```

2. **Build the project**:
    ```bash
    python3 build.py
    ```

# Usage
1. Develop and Build the DPU Program

    - Implement the DPU-side program.

    - Compile and link it using the UPMEM SDK toolchain.

2. Run in the UPMEM SDK Simulator

    - Use the SDK’s LLDB tool to load and execute the compiled DPU binary.
        ```bash
        dpu-lldb ./bin/dpu_code
        ```

    - Dump MRAM contents via the memory read command in LLDB.
        ```bash
        (lldb) memory read --force --outfile mram.bin --format u --size 1 0x08000000 0x0c000000
        ```

3. Run in uPIMulator to generated debug information
    - Open `run_uPIMulator.sh` and modify the command-line parameter `LOAD_LOCAL=0`

    - Run the uPIMulator to generate an output file containing debug information sections
        ```bash
        ./run_uPIMulator.sh
        ```
        The generated debug information is stored in `bin/mram.bin`.


4. Merge MRAM Dump and Debug Info

    - From the generated `bin/mram.bin`, extract the first `0x80000` bytes (in hexadecimal) and replace the first `0x80000` bytes of the `mram.bin` file previously obtained from LLDB.

5. Run uPIMulator Simulation
    - Replace bin/mram.bin with the merged mram.bin.
    - Open `run_uPIMulator.sh` and modify the command-line parameter `LOAD_LOCAL=1`
    - Run the uPIMulator


