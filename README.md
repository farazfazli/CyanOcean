# cyanocean
Open Source KVM Virtualization Panel written in Go

Create Debian and CentOS VMs in under 30 seconds. Currently in progress.

# Requirements
- libguestfs for editing VM images
- qemu-kvm for qemu/kvm
- libvirt-bin for interfacing with virtualization
- virtinst for easily provisioning virutal machines
- bridge-utils for bridging a connection to VMs

# YUM Pre-req install
```
yum -y install qemu-kvm libvirt virt-install bridge-utils
```

# Ubuntu/Debian
```
sudo apt-get -y install qemu-kvm libvirt-bin virtinst bridge-utils
```

# Setting up the Host->VM bridge

## Ubuntu/Debian config
```
// TODO
```

## CentOS/RHEL config
```
// TODO
```
