# 1brc
Revisit the One Billion Row Challenge in Golang

I discovered the [One Billion Row Challenge](https://github.com/gunnarmorling/1brc) recently :
> The One Billion Row Challenge (1BRC) is a fun exploration of how far modern Java can be pushed for aggregating one billion rows from a text file.
> Grab all your (virtual) threads, reach out to SIMD, optimize your GC, or pull any other trick, and create the fastest implementation for solving this task!

[Here](https://github.com/aytechnet/1brc/src/main.go) is a new proposed solution in Golang for this challenge.

## Results on a Ryzen 5 8540U with 12Gb of available RAM and a NVMe SSD disk

These are the results from running the challenge winner in Java with alternate proposed Golang implementation on a six core on a budget laptop running an AMD Ryzen 5 8540U with 16Gb of RAM and a NVMe SSD disk. Please note there is only 12Gb of available RAM. Tests are run 5 times.

| # | Result (m:s.ms) | Implementation     | Language | Submitter     | Notes     |
|---|-----------------|--------------------|----------|---------------|-----------|
| 1 | 00:05.032 | [link](https://github.com/aytechnet/1brc/src/main.go)| Golang 1.25.1| [François Pons](https://github.com/aytechnet)| Golang 1.25.1, default mode |
| 2 | 00:05.841 | [link](https://github.com/aytechnet/1brc/src/main.go)| Golang 1.25.1| [François Pons](https://github.com/aytechnet)| Golang 1.25.1, mmap mode |
| 3 | 00:07.334 | [link](https://github.com/gunnarmorling/1brc/blob/main/src/main/java/dev/morling/onebrc/CalculateAverage_thomaswue.java)| Java 25.0.1| [Thomas Wuerthinger](https://github.com/thomaswue), [Quan Anh Mai](https://github.com/merykitty), [Alfonso² Peterssen](https://github.com/mukel) | GraalVM 25.0.1 native binary with -O3, uses Unsafe |
| 4 | 00:07.452 | [link](https://github.com/gunnarmorling/1brc/blob/main/src/main/java/dev/morling/onebrc/CalculateAverage_thomaswue.java)| Java 25.0.1| [Thomas Wuerthinger](https://github.com/thomaswue), [Quan Anh Mai](https://github.com/merykitty), [Alfonso² Peterssen](https://github.com/mukel) | OpenJDK 25.0.1, uses Unsafe |
| 5 | 00:08.247 | [link](https://github.com/aytechnet/1brc/src/main.go)| Golang 1.25.1| [François Pons](https://github.com/aytechnet)| Golang 1.25.1, seq mode |
| 6 | 00:08.602 | [link](https://github.com/gunnarmorling/1brc/blob/main/src/main/go/AlexanderYastrebov/calc.go)| Golang 1.25.1| [Alexander Yastrebov](https://github.com/AlexanderYastrebov)| Golang 1.25.1 |
| 7 | 00:08.624 | [link](https://github.com/gunnarmorling/1brc/blob/main/src/main/go/elh/main.go)| Golang 1.25.1| [elh](https://github.com/elh)| Golang 1.25.1 |

## Results on a Ryzen 9 6900HX with 60Gb of available RAM and a NVMe SSD disk

These are the results from running the challenge winner in Java with alternate proposed Golang implementation on a eight core of a mini-PC running an AMD Ryzen 9 6900HX with 64Gb of RAM and a NVMe SSD disk. The `measurements.txt` file is fully cached. Tests are run 5 times.

| # | Result (m:s.ms) | Implementation     | Language | Submitter     | Notes     |
|---|-----------------|--------------------|----------|---------------|-----------|
| 1 | 00:02.007 | [link](https://github.com/gunnarmorling/1brc/blob/main/src/main/java/dev/morling/onebrc/CalculateAverage_thomaswue.java)| Java 25.0.1| [Thomas Wuerthinger](https://github.com/thomaswue), [Quan Anh Mai](https://github.com/merykitty), [Alfonso² Peterssen](https://github.com/mukel) | GraalVM 25.0.1 native binary with -O3, uses Unsafe |
| 2 | 00:02.706 | [link](https://github.com/aytechnet/1brc/src/main.go)| Golang 1.25.1| [François Pons](https://github.com/aytechnet)| Golang 1.25.1, default mode |
| 3 | 00:02.708 | [link](https://github.com/aytechnet/1brc/src/main.go)| Golang 1.25.1| [François Pons](https://github.com/aytechnet)| Golang 1.25.1, seq mode |
| 4 | 00:02.821 | [link](https://github.com/aytechnet/1brc/src/main.go)| Golang 1.25.1| [François Pons](https://github.com/aytechnet)| Golang 1.25.1, mmap mode |
| 5 | 00:03.072 | [link](https://github.com/gunnarmorling/1brc/blob/main/src/main/java/dev/morling/onebrc/CalculateAverage_thomaswue.java)| Java 25.0.1| [Thomas Wuerthinger](https://github.com/thomaswue), [Quan Anh Mai](https://github.com/merykitty), [Alfonso² Peterssen](https://github.com/mukel) | OpenJDK 25.0.1, uses Unsafe |
| 6 | 00:04.845 | [link](https://github.com/gunnarmorling/1brc/blob/main/src/main/go/elh/main.go)| Golang 1.25.1| [elh](https://github.com/elh)| Golang 1.25.1 |
| 7 | 00:06.259 | [link](https://github.com/gunnarmorling/1brc/blob/main/src/main/go/AlexanderYastrebov/calc.go)| Golang 1.25.1| [Alexander Yastrebov](https://github.com/AlexanderYastrebov)| Golang 1.25.1 |

