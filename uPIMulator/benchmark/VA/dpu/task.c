/*
* Vector addition with multiple tasklets
*
*/
#include <stdint.h>
// #include <stdio.h>
#include <defs.h>
#include <mram.h>
#include <alloc.h>
#include <perfcounter.h>
#include <barrier.h>
#include <limits.h>

// #include "../support/common.h"
// __host dpu_arguments_t DPU_INPUT_ARGUMENTS;

// // vector_addition: Computes the vector addition of a cached block 
// static void vector_addition(T *bufferB, T *bufferA, unsigned int l_size) {
//     // for (unsigned int i = 0; i < l_size; i++){
//     //     bufferB[i] += bufferA[i];
//     // }
//     int a = 0xffffffff;
//     // printf("%d", a);
// }

// // Barrier
// BARRIER_INIT(my_barrier, NR_TASKLETS);

// extern int main_kernel1(void);

// int (*kernels[nr_kernels])(void) = {main_kernel1};



/* Buffer in MRAM. */

#define GENERATE_DATA 1

typedef struct {
    int employee_id;
    int salary;
    int age;
} employee_record_t;

#define TABLE_SIZE 10
__mram employee_record_t employee_table[TABLE_SIZE];

#if GENERATE_DATA
void generate_database() {
    for (int i = 0; i < TABLE_SIZE; i += NR_TASKLETS) {
        employee_record_t record;
        record.employee_id = i;
        record.salary = 60000 + i*10000;
        record.age = 22 + i;
        employee_table[i] = record;
    }
}
#else
void count_high_earners() {
    #define SALARY_THRESHOLD 100000
    uint32_t tasklet_id = me();
    uint32_t local_count = 0;

    for (uint32_t i = tasklet_id; i < TABLE_SIZE; i += NR_TASKLETS) {
        if (employee_table[i].salary > SALARY_THRESHOLD) {
            local_count++;
        }
    }
}
#endif

int main() {
#if GENERATE_DATA
    generate_database();
#else
    count_high_earners();
#endif
    return 0;
}



// main_kernel1
// int main_kernel1() {
//     unsigned int tasklet_id = me();
// #if PRINT
//     printf("tasklet_id = %u\n", tasklet_id);
// #endif
//     if (tasklet_id == 0){ // Initialize once the cycle counter
//         mem_reset(); // Reset the heap
//     }
//     // Barrier
//     barrier_wait(&my_barrier);

//     uint32_t input_size_dpu_bytes = DPU_INPUT_ARGUMENTS.size; // Input size per DPU in bytes
//     uint32_t input_size_dpu_bytes_transfer = DPU_INPUT_ARGUMENTS.transfer_size; // Transfer input size per DPU in bytes

//     // Address of the current processing block in MRAM
//     uint32_t base_tasklet = tasklet_id << BLOCK_SIZE_LOG2;
//     uint32_t mram_base_addr_A = (uint32_t)DPU_MRAM_HEAP_POINTER;
//     uint32_t mram_base_addr_B = (uint32_t)(DPU_MRAM_HEAP_POINTER + input_size_dpu_bytes_transfer);

//     // Initialize a local cache to store the MRAM block
//     T *cache_A = (T *) mem_alloc(BLOCK_SIZE);
//     T *cache_B = (T *) mem_alloc(BLOCK_SIZE);

//     for(unsigned int byte_index = base_tasklet; byte_index < input_size_dpu_bytes; byte_index += BLOCK_SIZE * NR_TASKLETS){

//         // Bound checking
//         uint32_t l_size_bytes = (byte_index + BLOCK_SIZE >= input_size_dpu_bytes) ? (input_size_dpu_bytes - byte_index) : BLOCK_SIZE;

//         // Load cache with current MRAM block
//         mram_read((__mram_ptr void const*)(mram_base_addr_A + byte_index), cache_A, l_size_bytes);
//         mram_read((__mram_ptr void const*)(mram_base_addr_B + byte_index), cache_B, l_size_bytes);

//         // Computer vector addition
//         vector_addition(cache_B, cache_A, l_size_bytes >> DIV);

//         // Write cache to current MRAM block
//         mram_write(cache_B, (__mram_ptr void*)(mram_base_addr_B + byte_index), l_size_bytes);

//     }

//     return 0;
// }