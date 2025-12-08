package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"slices"
	"sync"
	"syscall"
	"sync/atomic"

	"github.com/aytechnet/decimal"
)

type (
	measurement struct {
		hash    atomic.Uint64
		minT    atomic.Int64
		maxT    atomic.Int64
		sumT    atomic.Int64
		countT  atomic.Int64
		nameLen int
		nameBuf [208]byte
	}

	measurements struct {
		total       atomic.Int64
		numParsers  int
		results     [capacity]measurement
	}

	job struct {
		maxOffset atomic.Int64
		bufLen    int
		buf       [bufSize]byte
	}
)

const (
	delta = 439
	capacity = 1 << 16 // must be a power of 2 for modulo calculation

	// buffer size
	bufSize = 512 * 1024 // 1Mb

	// use FNV-1a hash
	fnv1aOffset64 = 14695981039346656037
	fnv1aPrime64  = 1099511628211
)

func main() {
	var mode, filename, cpuprofile string
	var res measurements

	flag.StringVar(&mode, "mode", "default", "Which mode to use among 'mmap', 'seq' and 'default'")
	flag.StringVar(&filename, "file", "", "Measurements file to use")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "Write cpuprofile to file")
	flag.IntVar(&res.numParsers, "parsers", runtime.NumCPU(), "Number of thread to use for parsing")
 
	flag.Parse()

	if filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	if cpuprofile != "" {
		if f, err := os.Create(cpuprofile); err != nil {
			log.Fatal(err)
		} else {
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
	}

	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Open: %v", err)
	}
	defer f.Close()

	switch mode {
	case "mmap":
		res.processMmap(f)
	case "seq":
		res.processSeq(f)
	default:
		res.process(f)
	}
}

func (res *measurements) processMmap(f *os.File) {
	jobs := make([]job, res.numParsers)

	fi, err := f.Stat()
	if err != nil {
		log.Fatalf("Stat: %v", err)
	}

	size := fi.Size()
	chunkSize := size/int64(len(jobs))
	if chunkSize <= 100 {
		log.Fatalf("Invalid file size: %d", size)
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		log.Fatalf("Mmap: %v", err)
	}

	defer func() {
		if err := syscall.Munmap(data); err != nil {
			log.Fatalf("Munmap: %v", err)
		}
	}()

	offset := chunkSize
	for i := range jobs {
		j := &jobs[i]

		if i == len(jobs)-1 {
			j.maxOffset.Store(size)
		} else {
			j.maxOffset.Store(-offset)
		}
		offset += chunkSize
	}

	var wg sync.WaitGroup

	wg.Add(len(jobs))

	offset = 0
	for i := range jobs {
		go func(i int, offset int64){
			defer wg.Done()

			j := &jobs[i]

			maxOffset := j.maxOffset.Load()
			if maxOffset < 0 {
				maxOffset = -maxOffset

				if nlPos := bytes.IndexByte(data[maxOffset:], '\n'); nlPos >= 0 {
					maxOffset += int64(nlPos+1)
				}
			}

			if offset > 0 {
				if nlPos := bytes.IndexByte(data[offset:maxOffset], '\n'); nlPos >= 0 {
					offset += int64(nlPos+1)
				}
			}

			res.processChunk(data[offset:maxOffset])
		}(i, offset)

		offset += chunkSize
	}

	wg.Wait()

	res.printResults()
}

func (res *measurements) processSeq(f *os.File) {
	jobs := make([]job, res.numParsers)
	free_jobs := make(chan *job, len(jobs))
	ready_jobs := make(chan *job, len(jobs))

	var wg sync.WaitGroup

	wg.Add(len(jobs))

	for i := range jobs {
		j := &jobs[i]

		free_jobs <- j
		go func(){
			for j := range ready_jobs {
				res.processChunk(j.buf[:j.bufLen])
				free_jobs <- j
			}
			wg.Done()
		}()
	}

	var prev_j *job

	pos := 0
	nlPos := 0
	for j := range free_jobs {
		if count, _ := f.Read(j.buf[pos:]); count > 0 {
			// finalize previous line from previous job
			if pos > 0 /* && prev_j != nil */ {
				copy(j.buf[:pos], prev_j.buf[bufSize-pos:])
			}

			// prepare next buffer
			if nlPos = bytes.LastIndexByte(j.buf[:pos+count], '\n'); nlPos < 0 {
				log.Fatalf("buffer too small for a complete line");
			} else {
				pos = pos + count - nlPos - 1
				prev_j = j
				j.bufLen = nlPos + 1
			}

			// spawn a new job on this buffer
			ready_jobs <- j
		} else {
			break
		}
	}

	close(ready_jobs)

	wg.Wait()

	close(free_jobs)

	res.printResults()
}

func (res *measurements) process(f *os.File) {
	jobs := make([]job, res.numParsers)

	fi, err := f.Stat()
	if err != nil {
		log.Fatalf("Stat: %v", err)
	}

	size := fi.Size()
	chunkSize := size/int64(len(jobs))
	if chunkSize <= 100 {
		log.Fatalf("Invalid file size: %d", size)
	}

	offset := chunkSize
	for i := range jobs {
		j := &jobs[i]

		if i == len(jobs)-1 {
			j.maxOffset.Store(size)
		} else {
			j.maxOffset.Store(-offset)
		}
		offset += chunkSize
	}

	var wg sync.WaitGroup

	wg.Add(len(jobs))

	offset = 0
	for i := range jobs {
		go func(i int, offset int64){
			defer wg.Done()

			j := &jobs[i]

			nlSkipFirst := offset > 0

			for {
				maxLen := bufSize
				maxOffset := j.maxOffset.Load()
				if maxOffset < 0 {
					maxOffset = -maxOffset
				}
				if offset + int64(maxLen) > maxOffset {
					maxLen = int(maxOffset - offset)
				}

				if count, _ := f.ReadAt(j.buf[:maxLen], offset); count > 0 {
					pos := 0
					if nlSkipFirst {
						if nlPos := bytes.IndexByte(j.buf[:maxLen], '\n'); nlPos >= 0 {
							pos = nlPos + 1
							jobs[i-1].maxOffset.Store(offset + int64(pos))
							nlSkipFirst = false
						} else {
							log.Fatalf("Unable to seek to next line at job n°%d", i)
						}
					}

					if nlPos := bytes.LastIndexByte(j.buf[pos:maxLen], '\n'); nlPos >= 0 {
						j.bufLen = pos + nlPos + 1
						offset += int64(j.bufLen)

						res.processChunk(j.buf[pos:j.bufLen])
					} else {
						log.Fatalf("Buffer too small at job n°%d", i)
					}
				} else {
					return
				}

				maxOffset = j.maxOffset.Load()
				for maxOffset < 0 {
					maxOffset = j.maxOffset.Load()
				}
				if offset >= maxOffset {
					return
				}
			}
		}(i, offset)

		offset += chunkSize
	}

	wg.Wait()

	res.printResults()
}

func (res *measurements) printResults() {
	log.Printf("Read %d entries", res.total.Load())

	ids := make([]int, 0, capacity)

	for i := range res.results {
		m := &res.results[i]

		if m.nameLen > 0 {
			ids = append(ids, i)
		}
	}

	slices.SortFunc(ids, func(a,b int) int {
		return bytes.Compare(res.results[a].nameBuf[0:res.results[a].nameLen], res.results[b].nameBuf[0:res.results[b].nameLen])
	})

	count := 0
	fmt.Print("{")
	for _, i := range ids {
		m := &res.results[i]

		var buf [128]byte // name is 100 chars max, each 3 decimals is 5 bytes max on output
		b := buf[0:0]

		if count > 0 {
			b = append(b, ',', ' ')
		}
		
		b = append(b, m.nameBuf[0:m.nameLen]...)
		b = append(b, '=')
		b = decimal.New(m.minT.Load(), -1).BytesToFixed(b, 1)
		b = append(b, '/')
		b = decimal.New(m.sumT.Load(), -1).Div(decimal.NewFromInt(m.countT.Load()-1)).BytesToFixed(b, 1)
		b = append(b, '/')
		b = decimal.New(m.maxT.Load(), -1).BytesToFixed(b, 1)
		count++

		fmt.Print(string(b))
		// fmt.Printf("%s=%s/%s/%s", m.nameBuf[0:m.nameLen], decimal.New(m.minT.Load(), -1).StringFixed(1), decimal.New(m.sumT.Load(), -1).Div(decimal.NewFromInt(m.countT.Load()-1)).StringFixed(1), decimal.New(m.maxT.Load(), -1).StringFixed(1))
		count++
	}
	fmt.Println("}")
}

func (res *measurements) processChunk(data []byte) {
	var total int64

	// assume valid input
	for len(data) > 0 {
		i := 0

		// compute FNV-1a hash
		idHash := uint64(fnv1aOffset64)
		for j, b := range data {
			if b == ';' {
				i = j
				break
			}

			// calculate FNV-1a hash
			idHash ^= uint64(b)
			idHash *= fnv1aPrime64
		}
		if idHash == 0 {
			idHash = uint64(len(data))
		}

		idData := data[:i]

		i++ // now i points to temperature

		var temp int64
		// parseNumber
		{
			negative := data[i] == '-'
			if negative {
				i++
			}

			temp = int64(data[i]-'0')
			i++

			if data[i] != '.' {
				temp = temp*10 + int64(data[i]-'0')
				i++
			}
			i++ // data[i] is '.'
			temp = temp*10 + int64(data[i]-'0')
			if negative {
				temp = -temp
			}

			data = data[i+2:]
		}

		// get measurement
		{
			i := idHash & uint64(capacity-1)
			entry := &res.results[i]

			for {
				if entry.hash.CompareAndSwap(0, idHash) {
					// make sure no race occurs as entry may be updated meanwhile as hash has been updated
					if entry.countT.Add(1) == 1 {
						entry.nameLen = len(idData)
						copy(entry.nameBuf[:], idData)
						entry.minT.Store(temp)
						entry.maxT.Store(temp)
						entry.sumT.Store(temp)
						entry.countT.Add(1) // unlock for update below
					} else {
						// wait for countT to be at least 2 for entry init to be complete
						for entry.countT.Load() < 2 {}

						// update existing entry
						minT := entry.minT.Load()
						for minT > temp {
							if entry.minT.CompareAndSwap(minT, temp) {
								break
							} else {
								minT = entry.minT.Load()
							}
						}
						maxT := entry.maxT.Load()
						for maxT < temp {
							if entry.maxT.CompareAndSwap(maxT, temp) {
								break
							} else {
								maxT = entry.maxT.Load()
							}
						}
						entry.sumT.Add(temp)
						entry.countT.Add(1)
					}
					break
				} else if entry.hash.Load() == idHash {
					// the entry is found and may be being updated by another thread
					// wait for countT to be at least 2 for entry init to be complete
					for entry.countT.Load() < 2 {}

					// now that name has been updated, check it is matching
					if len(idData) == entry.nameLen /* bytes.Compare(idData, entry.nameBuf[0:entry.nameLen]) == 0 */ {
						// update existing entry
						minT := entry.minT.Load()
						for minT > temp {
							if entry.minT.CompareAndSwap(minT, temp) {
								break
							} else {
								minT = entry.minT.Load()
							}
						}
						maxT := entry.maxT.Load()
						for maxT < temp {
							if entry.maxT.CompareAndSwap(maxT, temp) {
								break
							} else {
								maxT = entry.maxT.Load()
							}
						}
						entry.sumT.Add(temp)
						entry.countT.Add(1)

						break
					} else {
						// name does not match idData so jump to next entry
						i = (i + delta) & uint64(capacity-1)
						entry = &res.results[i]
					}
				} else {
					i = (i + delta) & uint64(capacity-1)
					entry = &res.results[i]
				}
			}
		}

		total++
	}

	res.total.Add(total)
}

func (entry *measurement) update(temp int64) {
	minT := entry.minT.Load()
	for minT > temp {
		if entry.minT.CompareAndSwap(minT, temp) {
			break
		} else {
			minT = entry.minT.Load()
		}
	}
	maxT := entry.maxT.Load()
	for maxT < temp {
		if entry.maxT.CompareAndSwap(maxT, temp) {
			break
		} else {
			maxT = entry.maxT.Load()
		}
	}
	entry.sumT.Add(temp)
	entry.countT.Add(1)
}

