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
  ];

  shellHook = ''
    initial_dir=$(pwd)

    mkdir -p $initial_dir/build
    cd $initial_dir/build

    echo "Cloning LeechCore..."
    git clone https://github.com/ufrisk/LeechCore
    git clone https://github.com/ufrisk/MemProcFS
    git clone https://github.com/ufrisk/LeechCore-plugins

    echo "Building LeechCore..."
    cd LeechCore/leechcore
    make

    cd ../../MemProcFS/vmm
    make
    
    cd ../memprocfs
    make

    # mkdir -p $initial_dir/build/includes
    # mkdir -p $initial_dir/build/files

    mkdir -p $initial_dir/pkg/vmm/lib
    mkdir -p $initial_dir/pkg/vmm/include
    
    cp ../files/*.so $initial_dir/pkg/vmm/lib/
    cp ../includes/*.h $initial_dir/pkg/vmm/include/
    
    cd ../../LeechCore-plugins/leechcore_device_qemu
    make CFLAGS="-I../includes -D LINUX -Wno-error=unused-result -Wno-error=maybe-uninitialized"

    cp ../files/*.so $initial_dir/pkg/vmm/lib/

    cd $initial_dir

    rm -rf $initial_dir/build
  '';
}
