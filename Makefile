all: default seq mmap

default: src/aytechnet
	time src/aytechnet -file ./measurements.txt >measurements_aytechnet-default.out
seq: src/aytechnet
	time src/aytechnet -mode seq -file ./measurements.txt >measurements_aytechnet-seq.out
mmap: src/aytechnet
	time src/aytechnet -mode mmap -file ./measurements.txt >measurements_aytechnet-mmap.out

src/aytechnet: src/main.go
	cd src; go get -u -a; go build -o aytechnet main.go

