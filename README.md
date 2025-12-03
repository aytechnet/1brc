# 1brc
Revisit the One Billion Row Challenge in Golang

I discovered the [One Billion Row Challenge](https://github.com/gunnarmorling/1brc) recently :
> The One Billion Row Challenge (1BRC) is a fun exploration of how far modern Java can be pushed for aggregating one billion rows from a text file.
> Grab all your (virtual) threads, reach out to SIMD, optimize your GC, or pull any other trick, and create the fastest implementation for solving this task!

[Here](https://github.com/aytechnet/1brc/blob/main/src/main.go) is a new proposed solution in Golang for this challenge. It was inspired from [Alexander Yastrebov](https://github.com/AlexanderYastrebov) for using FNV-1a hashing and a custom map along with the following optimizations :
 * intensive use of [sync/atomic](https://pkg.go.dev/sync/atomic) in order to avoid the `reduce` computation from all thread
 * almost no allocation (no GC stress) and memory consumption is very low
 * [FNV-1a hash](https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function) is slightly modified to produce a non-zero hash which is used to find an empty slot in hash table (remenber no lock are used)
 * hash collision are only distinguished with length of name (FNV-1a collisions are very rare in general and there are only around 50000 distinct weather station name used)
 * buffer alignement to EOL is done in created thread to speed up the process a little
 * entry.update() has been manualy inlined

## Results on a Ryzen 5 8540U with 12Gb of available RAM and a NVMe SSD disk

These are the results from running the challenge winner in Java with alternate proposed Golang implementation on a six core (12 logical CPUs) on a budget laptop running an AMD Ryzen 5 8540U with 16Gb of RAM and a NVMe SSD disk. Please note there is only 12Gb of available RAM. Tests are run 5 times.

| # | Result (m:s.ms) | Implementation     | Language | Submitter     | Notes     |
|---|-----------------|--------------------|----------|---------------|-----------|
| 1 | 00:05.032 | [link](https://github.com/aytechnet/1brc/blob/main/src/main.go)| Golang 1.25.1| [François Pons](https://github.com/aytechnet)| Golang 1.25.1, default mode |
| 2 | 00:05.841 | [link](https://github.com/aytechnet/1brc/blob/main/src/main.go)| Golang 1.25.1| [François Pons](https://github.com/aytechnet)| Golang 1.25.1, mmap mode |
| 3 | 00:07.334 | [link](https://github.com/gunnarmorling/1brc/blob/main/src/main/java/dev/morling/onebrc/CalculateAverage_thomaswue.java)| Java 25.0.1| [Thomas Wuerthinger](https://github.com/thomaswue), [Quan Anh Mai](https://github.com/merykitty), [Alfonso² Peterssen](https://github.com/mukel) | GraalVM 25.0.1 native binary with -O3, uses Unsafe |
| 4 | 00:07.452 | [link](https://github.com/gunnarmorling/1brc/blob/main/src/main/java/dev/morling/onebrc/CalculateAverage_thomaswue.java)| Java 25.0.1| [Thomas Wuerthinger](https://github.com/thomaswue), [Quan Anh Mai](https://github.com/merykitty), [Alfonso² Peterssen](https://github.com/mukel) | OpenJDK 25.0.1, uses Unsafe |
| 5 | 00:08.247 | [link](https://github.com/aytechnet/1brc/blob/main/src/main.go)| Golang 1.25.1| [François Pons](https://github.com/aytechnet)| Golang 1.25.1, seq mode |
| 6 | 00:08.602 | [link](https://github.com/gunnarmorling/1brc/blob/main/src/main/go/AlexanderYastrebov/calc.go)| Golang 1.25.1| [Alexander Yastrebov](https://github.com/AlexanderYastrebov)| Golang 1.25.1 |
| 7 | 00:08.624 | [link](https://github.com/gunnarmorling/1brc/blob/main/src/main/go/elh/main.go)| Golang 1.25.1| [elh](https://github.com/elh)| Golang 1.25.1 |

## Results on a Ryzen 9 6900HX with 60Gb of available RAM and a NVMe SSD disk

These are the results from running the challenge winner in Java with alternate proposed Golang implementation on a eight core (16 logical CPUs) of a mini-PC running an AMD Ryzen 9 6900HX with 64Gb of RAM and a NVMe SSD disk. The `measurements.txt` file is fully cached. Tests are run 5 times.

| # | Result (m:s.ms) | Implementation     | Language | Submitter     | Notes     |
|---|-----------------|--------------------|----------|---------------|-----------|
| 1 | 00:02.007 | [link](https://github.com/gunnarmorling/1brc/blob/main/src/main/java/dev/morling/onebrc/CalculateAverage_thomaswue.java)| Java 25.0.1| [Thomas Wuerthinger](https://github.com/thomaswue), [Quan Anh Mai](https://github.com/merykitty), [Alfonso² Peterssen](https://github.com/mukel) | GraalVM 25.0.1 native binary with -O3, uses Unsafe |
| 2 | 00:02.706 | [link](https://github.com/aytechnet/1brc/blob/main/src/main.go)| Golang 1.25.1| [François Pons](https://github.com/aytechnet)| Golang 1.25.1, default mode |
| 3 | 00:02.708 | [link](https://github.com/aytechnet/1brc/blob/main/src/main.go)| Golang 1.25.1| [François Pons](https://github.com/aytechnet)| Golang 1.25.1, seq mode |
| 4 | 00:02.821 | [link](https://github.com/aytechnet/1brc/blob/main/src/main.go)| Golang 1.25.1| [François Pons](https://github.com/aytechnet)| Golang 1.25.1, mmap mode |
| 5 | 00:03.072 | [link](https://github.com/gunnarmorling/1brc/blob/main/src/main/java/dev/morling/onebrc/CalculateAverage_thomaswue.java)| Java 25.0.1| [Thomas Wuerthinger](https://github.com/thomaswue), [Quan Anh Mai](https://github.com/merykitty), [Alfonso² Peterssen](https://github.com/mukel) | OpenJDK 25.0.1, uses Unsafe |
| 6 | 00:04.845 | [link](https://github.com/gunnarmorling/1brc/blob/main/src/main/go/elh/main.go)| Golang 1.25.1| [elh](https://github.com/elh)| Golang 1.25.1 |
| 7 | 00:06.259 | [link](https://github.com/gunnarmorling/1brc/blob/main/src/main/go/AlexanderYastrebov/calc.go)| Golang 1.25.1| [Alexander Yastrebov](https://github.com/AlexanderYastrebov)| Golang 1.25.1 |

## Conclusion

This Golang implementation of One Billion Row Challenge seems now to be fastest than the best Java version using OpenJDK but not yet if the Java is compiled using GraalVM with -O3 optimization.

But Golang compilation is more than 1400 times faster than GraaVM native image generation (0.032s compared to 45.578s) on my budget laptop.

Golang is fantastic!
