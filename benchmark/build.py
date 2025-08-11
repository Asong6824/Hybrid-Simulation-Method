import os
import shutil
import subprocess
import argparse


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--num_dpus", type=int, default=1)
    parser.add_argument("--num_tasklets", type=int, default=1)
    args = parser.parse_args()

    benchmark_dir_path = os.path.dirname(__file__)

    build_dir_path = os.path.join(benchmark_dir_path, "build")

    if os.path.exists(build_dir_path):
        shutil.rmtree(build_dir_path)
    os.makedirs(build_dir_path)

    subprocess.run(
    [
        "cmake",
        "-D", f"NR_DPUS={args.num_dpus}",
        "-D", f"NR_TASKLETS={args.num_tasklets}",
        "-D", "CMAKE_C_FLAGS=-O3 -fno-tree-dce -fno-toplevel-reorder -g1",
        "-D", "CMAKE_CXX_FLAGS=-O3 -fno-tree-dce -fno-toplevel-reorder -g1",
        "-D", "CMAKE_BUILD_TYPE=Release",
        "-S", benchmark_dir_path,
        "-B", build_dir_path,
        "-G", "Ninja",
    ],
    check=True
)

    subprocess.run(["ninja", "-C", build_dir_path])
