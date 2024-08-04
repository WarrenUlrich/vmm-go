{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.git
    pkgs.makeWrapper
    pkgs.gcc8
    pkgs.pkg-config
    pkgs.libusb1
    pkgs.fuse
    pkgs.lz4
    pkgs.go
    pkgs.nixd
    pkgs.clang-tools
  ];

  shellHook = ''
    initial_dir=$(pwd)

    mkdir -p build
    cd build

    echo "Cloning LeechCore..."
    git clone https://github.com/ufrisk/LeechCore

    echo "Cloning LeechCore-plugins..."
    git clone https://github.com/ufrisk/LeechCore-plugins

    echo "Cloning MemProcFS..."
    git clone https://github.com/ufrisk/MemProcFS

    echo "Building LeechCore..."
    cd LeechCore/leechcore
    make

    echo "Building LeechCore-plugins..."
    cd ../../LeechCore-plugins/leechcore_device_qemu
    make

    echo "Building MemProcFS..."
    cd ../../MemProcFS/vmm
    make

    cd ../memprocfs
    make

    # Create directories for the binaries and headers
    mkdir -p $initial_dir/pkg/vmm/lib
    mkdir -p $initial_dir/pkg/vmm/include

    # Move the shared library files
    cp $initial_dir/build/MemProcFS/files/vmm.so $initial_dir/pkg/vmm/lib/vmm.so
    cp $initial_dir/build/MemProcFS/files/leechcore.so $initial_dir/pkg/vmm/lib/leechcore.so
    cp $initial_dir/build/LeechCore-plugins/files/leechcore_device_qemu.so $initial_dir/pkg/vmm/lib/leechcore_device_qemu.so

    # Move the header files
    cp $initial_dir/build/MemProcFS/includes/*.h $initial_dir/pkg/vmm/include/

    # Cleaning up build directory
    echo "Cleaning up build directory..."
    rm -rf $initial_dir/build

    cd $initial_dir
  '';
}
