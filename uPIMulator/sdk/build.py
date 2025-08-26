import os
import shutil
import subprocess
import argparse

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--num_tasklets", type=int, default=1)
    args = parser.parse_args()

    sdk_dir_path = os.path.dirname(__file__)
    build_dir_path = os.path.join(sdk_dir_path, "build")

    if os.path.exists(build_dir_path):
        shutil.rmtree(build_dir_path)
    os.makedirs(build_dir_path)

    subprocess.run(
        [
            "cmake",
            "-D", f"NR_TASKLETS={args.num_tasklets}",
            "-D", "CMAKE_BUILD_TYPE=None",  # 不使用默认的构建类型
            "-D", "CMAKE_C_FLAGS=-O0 -g0 -fno-asynchronous-unwind-tables -fno-unwind-tables -fno-dwarf2-cfi-asm",
            "-D", "CMAKE_CXX_FLAGS=-O0 -g0 -fno-asynchronous-unwind-tables -fno-unwind-tables -fno-dwarf2-cfi-asm",
            "-D", "CMAKE_EXE_LINKER_FLAGS=-Wl,--strip-debug",
            "-S", sdk_dir_path,
            "-B", build_dir_path,
            "-G", "Ninja",
        ],
        check=True
    )

    subprocess.run(["ninja", "-C", build_dir_path], check=True)

