package main

import (
	"net/http"
	"fmt"
	"time"
	"log"
	"encoding/json"
	"os/exec"
	"os"
	"regexp"
)

type VM struct {
	Key string `json:"key"`
	Name string `json:"name"`
	Os string `json:"os"`
	Ram int `json:"ram"`
	Vcpus int `json:"cpu"`
	Storage int `json:"storage"`
}

// ------------CONFIG----------- //
const gw = "69.30.244.113"
const ipv6gw = "2604:4300:a:6a::1"
//--––––------------------------//

const sh = "/bin/sh"
const c = "-c"

func main() {
	port := ":8000"
	fmt.Println("Starting webserver on " + port)

	server := &http.Server{
		Addr:         port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	http.HandleFunc("/create", create)
	log.Fatal(server.ListenAndServe())
}

func create(w http.ResponseWriter, r *http.Request) {
	var vm VM
	err := json.NewDecoder(r.Body).Decode(&vm)
	if err != nil || len(vm.Key) < 10 || vm.Name == "" || vm.Os == "" || vm.Ram == 0 || vm.Vcpus == 0 || vm.Storage == 0 || vm.Ram / 512 < 1 || vm.Storage > 50 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if (vm.Os != "centos" && vm.Os != "debian") {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Incorrect OS parameters, choose centos or debian")
		return
	}

	re := regexp.MustCompile("[^A-Za-z]")
	vm.Name = re.ReplaceAllString(vm.Name, "")

	if len(vm.Name) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Choose a longer VM name consisting of letters")
		return
	}

	imageName := vm.Name + ".qcow2"
	if _, err := os.Stat("~/" + imageName); err == nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Host already exists")
		return
	}

	// TODO: Check if IP space and server resources exist -- assign IP

	// Passed all checks, creating VM
	fmt.Printf("Hostname: %v || OS: %v || %vmb RAM || %v cores || Storage: %vGBs\n", vm.Name, vm.Os, vm.Ram, vm.Vcpus, vm.Storage)

	copyCommand := ""
	if		  vm.Os == "centos" {
		copyCommand = fmt.Sprintf("cp /var/lib/libvirt/images/CentOS-7-Base.qcow2 ~/%v.qcow2")
	} else if vm.Os == "debian" {
		copyCommand = fmt.Sprintf("cp /var/lib/libvirt/images/Debian-8.7-Base.qcow2 ~%v.qcow2")
	}
	
	cp, err := exec.Command(sh, c, copyCommand).Output()
		if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	fmt.Println(string(cp))

	resizeCommand := fmt.Sprintf("qemu-img resize %s.qcow2 +%vG", vm.Name, vm.Storage)
	qemu, err := exec.Command(sh, c, resizeCommand).Output()
	if err != nil {
		fmt.Println("Error resizing image")
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(string(qemu))

	editCommand := fmt.Sprintf(`
virt-edit -a %s "/etc/sysconfig/network" -e 's/GW/%s/' && \
virt-edit -a %s "/etc/sysconfig/network" -e 's/V6GW/%s/' && \
virt-edit -a %s "/etc/sysconfig/network-scripts/ifcfg-eth0" -e 's/IP/%s/' && \
virt-edit -a %s "/etc/sysconfig/network-scripts/ifcfg-eth0" -e 's/IPV6/%s\/128/' && \
virt-edit -a %s "/etc/sysconfig/network" "/etc/hostname" "/etc/hosts" -e 's/HN/%s/'
	`, imageName, gw, imageName, ipv6gw, imageName, "69.30.244.115", imageName, "2604:4300:a:6a::3", imageName, imageName)
	edit, err := exec.Command(sh, c, editCommand).Output()
	if err != nil {
		fmt.Println("Error editing VM image")
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(string(edit))

	addKeyCommand := fmt.Sprintf(`guestfish -a %s -i write /root/.ssh/authorized_keys "%s"`, imageName, vm.Key)
	addKey, err := exec.Command(sh, c, addKeyCommand).Output()
	if err != nil {
		fmt.Println("Error adding SSH key")
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(string(addKey))

	variant := "debian8"
	if vm.Os == "centos" {
		variant = "rhel6"
	}
	provisionCommand := fmt.Sprintf(`
		sudo virt-install --import --name %s --ram %v --vcpus %v --disk ~/%s,format=qcow2,bus=virtio 
		--network bridge=br0,model=virtio --os-type=linux --os-variant=%s --noautoconsole`, vm.Name, vm.Ram, imageName, variant)
	provision, err := exec.Command(sh, c, provisionCommand).Output()
	if err != nil {
		fmt.Println("Error provisioning VM")
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(string(provision))
	w.WriteHeader(http.StatusOK)
}