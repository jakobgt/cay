# Observation log

As this is my first foray into optimizing memory accessed and what not in Go, I try to document my findings here (also to flush my mental cache).

## 2022-08-19 Investigating the memory
It seems that the problem is with memory and the memory caches. Using perf and the perf-test.sh tool, cay has a better TLB rate (TLB metric group)
and higher instruction per cycle count (default perf summary), but worse memory cache usage.

Thus, I need to look into the memory locality and whether the cay map is way too sparse. Maybe look into the fillrate?

### Summary data:
(No metrics or metric group given to perf. )
Check the insn per cycle (instructions per cycle), the higher the better. Except for the 8 entry case, it is higher for cay than the builtin map.
Further branch-misses is significantly lower (upwards of 10x) for Cay in all cases (except 8).
```
$ ./perf-test.sh
# Benchmark size 8
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/8/cay':

          2,783.53 msec task-clock                #    1.026 CPUs utilized
             1,584      context-switches          #    0.569 K/sec
               302      cpu-migrations            #    0.108 K/sec
             8,032      page-faults               #    0.003 M/sec
     8,830,482,141      cycles                    #    3.172 GHz
    19,640,426,332      instructions              #    2.22  insn per cycle
     3,136,398,887      branches                  # 1126.770 M/sec
        10,685,521      branch-misses             #    0.34% of all branches

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/8/built':

          2,530.04 msec task-clock                #    1.031 CPUs utilized
             1,518      context-switches          #    0.600 K/sec
               271      cpu-migrations            #    0.107 K/sec
             7,503      page-faults               #    0.003 M/sec
     8,041,054,439      cycles                    #    3.178 GHz
    23,121,725,391      instructions              #    2.88  insn per cycle
     4,859,536,122      branches                  # 1920.733 M/sec
         2,037,432      branch-misses             #    0.04% of all branches

# Benchmark size 1k
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1k/cay':

          2,688.34 msec task-clock                #    1.033 CPUs utilized
             1,513      context-switches          #    0.563 K/sec
               292      cpu-migrations            #    0.109 K/sec
             8,183      page-faults               #    0.003 M/sec
     8,549,802,679      cycles                    #    3.180 GHz
    19,139,506,643      instructions              #    2.24  insn per cycle
     3,029,410,736      branches                  # 1126.869 M/sec
         4,031,192      branch-misses             #    0.13% of all branches

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1k/built':

          2,885.53 msec task-clock                #    1.029 CPUs utilized
             1,942      context-switches          #    0.673 K/sec
               272      cpu-migrations            #    0.094 K/sec
             7,385      page-faults               #    0.003 M/sec
     9,178,403,154      cycles                    #    3.181 GHz
    16,996,576,330      instructions              #    1.85  insn per cycle
     2,965,568,826      branches                  # 1027.738 M/sec
        33,831,083      branch-misses             #    1.14% of all branches

# Benchmark size 32k
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/32k/cay':

          2,727.82 msec task-clock                #    1.033 CPUs utilized
             1,463      context-switches          #    0.536 K/sec
               177      cpu-migrations            #    0.065 K/sec
             8,134      page-faults               #    0.003 M/sec
     8,726,841,458      cycles                    #    3.199 GHz
    16,212,499,857      instructions              #    1.86  insn per cycle
     2,561,185,225      branches                  #  938.913 M/sec
         3,515,472      branch-misses             #    0.14% of all branches

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/32k/built':

          2,680.82 msec task-clock                #    1.033 CPUs utilized
             1,656      context-switches          #    0.618 K/sec
               193      cpu-migrations            #    0.072 K/sec
             8,738      page-faults               #    0.003 M/sec
     8,522,265,149      cycles                    #    3.179 GHz
    14,409,612,866      instructions              #    1.69  insn per cycle
     2,459,666,444      branches                  #  917.506 M/sec
        31,985,482      branch-misses             #    1.30% of all branches

# Benchmark size 512k
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/512k/cay':

          3,502.46 msec task-clock                #    1.024 CPUs utilized
             1,844      context-switches          #    0.526 K/sec
               342      cpu-migrations            #    0.098 K/sec
            19,913      page-faults               #    0.006 M/sec
    10,549,826,964      cycles                    #    3.012 GHz
    15,326,648,502      instructions              #    1.45  insn per cycle
     2,645,720,701      branches                  #  755.388 M/sec
         6,797,453      branch-misses             #    0.26% of all branches

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/512k/built':

          3,206.88 msec task-clock                #    1.026 CPUs utilized
             1,511      context-switches          #    0.471 K/sec
               249      cpu-migrations            #    0.078 K/sec
            20,389      page-faults               #    0.006 M/sec
    10,039,149,266      cycles                    #    3.131 GHz
    14,010,827,099      instructions              #    1.40  insn per cycle
     2,426,420,176      branches                  #  756.630 M/sec
        19,769,838      branch-misses             #    0.81% of all branches

# Benchmark size 1m
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1m/cay':

          4,092.63 msec task-clock                #    1.018 CPUs utilized
             1,839      context-switches          #    0.449 K/sec
               367      cpu-migrations            #    0.090 K/sec
            31,515      page-faults               #    0.008 M/sec
    12,407,871,126      cycles                    #    3.032 GHz
    18,415,418,260      instructions              #    1.48  insn per cycle
     3,352,476,604      branches                  #  819.151 M/sec
        10,492,666      branch-misses             #    0.31% of all branches

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1m/built':

          3,665.66 msec task-clock                #    1.022 CPUs utilized
             1,901      context-switches          #    0.519 K/sec
               380      cpu-migrations            #    0.104 K/sec
            30,913      page-faults               #    0.008 M/sec
    11,654,925,508      cycles                    #    3.179 GHz
    15,773,511,313      instructions              #    1.35  insn per cycle
     2,810,254,603      branches                  #  766.643 M/sec
        21,821,482      branch-misses             #    0.78% of all branches
```

### TLB data
Here we run with the TLB perf metric group. Looking below the `dtlb_load_misses.walk_pending` metrics is lower for cay
and especially for the big sized maps. E.g., for the 1M case cay is at `88,098,107`, whereas the builtin is at `1,009,407,450`.
Thus, it does not seem that the TLB is the culprit. The iTLB does not look very different and neither does dtlb for store_misses.

```
$ ./perf-test.sh TLB
# Benchmark size 8
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/8/cay':
        24,531,832      itlb_misses.walk_pending  #     0.00 Page_Walks_Utilization
         2,710,554      dtlb_store_misses.walk_pending
     8,672,852,677      cycles
         8,993,497      dtlb_load_misses.walk_pending
                 1      ept.walk_pending
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/8/built':
        24,878,419      itlb_misses.walk_pending  #     0.00 Page_Walks_Utilization
         2,825,332      dtlb_store_misses.walk_pending
    11,432,187,042      cycles
        10,390,090      dtlb_load_misses.walk_pending
                 1      ept.walk_pending

# Benchmark size 1k
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1k/cay':
        23,910,628      itlb_misses.walk_pending  #     0.00 Page_Walks_Utilization
         2,647,062      dtlb_store_misses.walk_pending
     8,806,311,421      cycles
         8,881,615      dtlb_load_misses.walk_pending
                 2      ept.walk_pending
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1k/built':
        23,542,237      itlb_misses.walk_pending  #     0.00 Page_Walks_Utilization
         2,460,400      dtlb_store_misses.walk_pending
     8,636,039,813      cycles
         9,985,277      dtlb_load_misses.walk_pending
                 0      ept.walk_pending

# Benchmark size 32k
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/32k/cay':
        24,763,302      itlb_misses.walk_pending  #     0.00 Page_Walks_Utilization
         2,866,027      dtlb_store_misses.walk_pending
    11,041,198,654      cycles
        10,393,644      dtlb_load_misses.walk_pending
                 2      ept.walk_pending
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/32k/built':
        24,418,387      itlb_misses.walk_pending  #     0.00 Page_Walks_Utilization
         3,238,301      dtlb_store_misses.walk_pending
     8,609,729,426      cycles
        18,096,090      dtlb_load_misses.walk_pending
                 0      ept.walk_pending

# Benchmark size 512k
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/512k/cay':
        25,283,132      itlb_misses.walk_pending  #     0.00 Page_Walks_Utilization
         5,555,387      dtlb_store_misses.walk_pending
    10,472,566,848      cycles
        45,438,830      dtlb_load_misses.walk_pending
                 0      ept.walk_pending
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/512k/built':
        26,503,918      itlb_misses.walk_pending  #     0.04 Page_Walks_Utilization
         6,113,194      dtlb_store_misses.walk_pending
    10,058,469,569      cycles
       830,822,775      dtlb_load_misses.walk_pending
                 1      ept.walk_pending

# Benchmark size 1m
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1m/cay':
        26,309,652      itlb_misses.walk_pending  #     0.00 Page_Walks_Utilization
         9,421,192      dtlb_store_misses.walk_pending
    12,705,813,840      cycles
        88,098,107      dtlb_load_misses.walk_pending
                 1      ept.walk_pending
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1m/built':
        26,283,712      itlb_misses.walk_pending  #     0.05 Page_Walks_Utilization
        11,332,328      dtlb_store_misses.walk_pending
    11,453,552,268      cycles
     1,009,407,450      dtlb_load_misses.walk_pending
                 2      ept.walk_pending
```
### Cache Misses
The cache misses for cay looks worse than builtin and that might be the reason for the latency differences. Check
out the L*MPKI values, which means (taken from perf list)
- L1MPKI: [L1 cache true misses per kilo instruction for retired demand loads]
- L2MPKI: [L2 cache true misses per kilo instruction for retired demand loads]
- L3MPKI: [L3 cache true misses per kilo instruction for retired demand loads]
- L2HPKI_All: [L2 cache hits per kilo instruction for all request types (including speculative)]
- L2MPKI_All: [L2 cache misses per kilo instruction for all request types (including speculative)]

For all five, cay is worse. For L2MPKI_All, cay is ~10% worse, and for the 1m case:
- L1MPKI is 44% higher/worse for cay
- L2MPKI is 50% higher/worse for cay
- L3MPKI is 13% higher/worse for cay.

```
$ ./perf-test.sh Cache_Misses
# Benchmark size 8
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/8/cay':
    20,985,177,035      inst_retired.any          #     0.98 L2MPKI_All
                                                  #     0.88 L2HPKI_All               (53.94%)
        20,592,361      l2_rqsts.miss                                                 (54.28%)
        39,163,824      l2_rqsts.references                                           (54.95%)
    21,007,535,636      inst_retired.any          #     0.20 L1MPKI                   (55.59%)
         4,234,160      mem_load_retired.l1_miss                                      (56.61%)
    20,820,691,960      inst_retired.any          #     0.12 L2MPKI                   (57.11%)
         2,412,020      mem_load_retired.l2_miss                                      (56.71%)
    20,812,248,529      inst_retired.any          #     0.09 L3MPKI                   (55.86%)
         1,899,423      mem_load_retired.l3_miss                                      (54.95%)
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/8/built':
    23,164,212,366      inst_retired.any          #     0.84 L2MPKI_All
                                                  #     0.78 L2HPKI_All               (54.26%)
        19,550,786      l2_rqsts.miss                                                 (54.75%)
        37,516,125      l2_rqsts.references                                           (55.41%)
    23,060,758,516      inst_retired.any          #     0.21 L1MPKI                   (55.96%)
         4,875,197      mem_load_retired.l1_miss                                      (56.63%)
    23,011,803,612      inst_retired.any          #     0.11 L2MPKI                   (56.86%)
         2,507,324      mem_load_retired.l2_miss                                      (56.19%)
    23,161,190,056      inst_retired.any          #     0.08 L3MPKI                   (55.32%)
         1,864,263      mem_load_retired.l3_miss                                      (54.61%)

# Benchmark size 1k
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1k/cay':
    19,559,433,538      inst_retired.any          #     2.13 L2MPKI_All
                                                  #    16.68 L2HPKI_All               (55.08%)
        41,616,379      l2_rqsts.miss                                                 (55.48%)
       367,807,837      l2_rqsts.references                                           (55.76%)
    19,551,157,782      inst_retired.any          #     5.32 L1MPKI                   (56.02%)
       103,936,345      mem_load_retired.l1_miss                                      (56.31%)
    19,713,489,843      inst_retired.any          #     0.22 L2MPKI                   (55.96%)
         4,267,553      mem_load_retired.l2_miss                                      (55.53%)
    19,667,423,864      inst_retired.any          #     0.09 L3MPKI                   (55.13%)
         1,836,495      mem_load_retired.l3_miss                                      (54.73%)
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1k/built':
    16,875,829,573      inst_retired.any          #     2.06 L2MPKI_All
                                                  #    19.84 L2HPKI_All               (54.59%)
        34,758,222      l2_rqsts.miss                                                 (55.46%)
       369,643,159      l2_rqsts.references                                           (55.93%)
    16,834,950,967      inst_retired.any          #     4.49 L1MPKI                   (56.15%)
        75,571,486      mem_load_retired.l1_miss                                      (56.36%)
    16,834,082,172      inst_retired.any          #     0.23 L2MPKI                   (56.40%)
         3,898,075      mem_load_retired.l2_miss                                      (55.46%)
    16,898,834,396      inst_retired.any          #     0.11 L3MPKI                   (54.92%)
         1,830,326      mem_load_retired.l3_miss                                      (54.73%)

# Benchmark size 32k
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/32k/cay':
    15,723,857,870      inst_retired.any          #    22.51 L2MPKI_All
                                                  #     5.88 L2HPKI_All               (54.17%)
       353,966,152      l2_rqsts.miss                                                 (54.50%)
       446,495,417      l2_rqsts.references                                           (55.16%)
    15,768,811,226      inst_retired.any          #     6.05 L1MPKI                   (55.80%)
        95,393,069      mem_load_retired.l1_miss                                      (56.42%)
    15,807,752,488      inst_retired.any          #     4.41 L2MPKI                   (57.01%)
        69,643,967      mem_load_retired.l2_miss                                      (56.40%)
    15,790,548,309      inst_retired.any          #     0.25 L3MPKI                   (55.68%)
         3,940,655      mem_load_retired.l3_miss                                      (54.87%)
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/32k/built':
    14,367,825,777      inst_retired.any          #    23.20 L2MPKI_All
                                                  #     5.44 L2HPKI_All               (54.97%)
       333,336,155      l2_rqsts.miss                                                 (55.40%)
       411,510,326      l2_rqsts.references                                           (55.96%)
    14,397,381,463      inst_retired.any          #     5.36 L1MPKI                   (56.18%)
        77,111,961      mem_load_retired.l1_miss                                      (56.47%)
    14,473,547,929      inst_retired.any          #     4.15 L2MPKI                   (56.04%)
        60,061,565      mem_load_retired.l2_miss                                      (55.48%)
    14,415,442,764      inst_retired.any          #     0.19 L3MPKI                   (54.86%)
         2,790,478      mem_load_retired.l3_miss                                      (54.64%)

# Benchmark size 512k
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/512k/cay':
    15,631,316,763      inst_retired.any          #    16.93 L2MPKI_All
                                                  #     5.14 L2HPKI_All               (55.23%)
       264,593,214      l2_rqsts.miss                                                 (55.83%)
       344,995,786      l2_rqsts.references                                           (56.15%)
    15,654,062,627      inst_retired.any          #     6.02 L1MPKI                   (56.29%)
        94,255,979      mem_load_retired.l1_miss                                      (56.36%)
    15,605,701,802      inst_retired.any          #     4.26 L2MPKI                   (55.65%)
        66,489,043      mem_load_retired.l2_miss                                      (55.14%)
    15,547,654,451      inst_retired.any          #     2.24 L3MPKI                   (54.74%)
        34,758,323      mem_load_retired.l3_miss                                      (54.60%)
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/512k/built':
    14,072,355,231      inst_retired.any          #    18.83 L2MPKI_All
                                                  #     5.49 L2HPKI_All               (55.16%)
       264,999,055      l2_rqsts.miss                                                 (55.27%)
       342,252,958      l2_rqsts.references                                           (55.55%)
    13,892,173,635      inst_retired.any          #     4.55 L1MPKI                   (55.72%)
        63,269,904      mem_load_retired.l1_miss                                      (56.14%)
    14,016,947,465      inst_retired.any          #     3.00 L2MPKI                   (56.00%)
        41,990,744      mem_load_retired.l2_miss                                      (55.67%)
    14,057,421,320      inst_retired.any          #     1.77 L3MPKI                   (55.51%)
        24,931,461      mem_load_retired.l3_miss                                      (54.98%)

# Benchmark size 1m
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1m/cay':
    18,198,975,943      inst_retired.any          #    19.11 L2MPKI_All
                                                  #     5.08 L2HPKI_All               (55.15%)
       347,763,047      l2_rqsts.miss                                                 (55.46%)
       440,222,073      l2_rqsts.references                                           (55.78%)
    18,713,800,265      inst_retired.any          #     7.14 L1MPKI                   (56.00%)
       133,637,509      mem_load_retired.l1_miss                                      (56.31%)
    18,731,691,340      inst_retired.any          #     5.20 L2MPKI                   (55.90%)
        97,471,680      mem_load_retired.l2_miss                                      (55.49%)
    18,294,134,138      inst_retired.any          #     1.97 L3MPKI                   (55.08%)
        36,078,463      mem_load_retired.l3_miss                                      (54.84%)
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1m/built':
    15,527,809,018      inst_retired.any          #    20.74 L2MPKI_All
                                                  #     5.44 L2HPKI_All               (54.94%)
       322,099,956      l2_rqsts.miss                                                 (55.41%)
       406,510,919      l2_rqsts.references                                           (55.75%)
    16,020,021,219      inst_retired.any          #     4.95 L1MPKI                   (56.18%)
        79,351,682      mem_load_retired.l1_miss                                      (56.51%)
    16,109,739,903      inst_retired.any          #     3.45 L2MPKI                   (56.07%)
        55,573,102      mem_load_retired.l2_miss                                      (55.47%)
    15,846,374,536      inst_retired.any          #     1.74 L3MPKI                   (55.09%)
        27,608,897      mem_load_retired.l3_miss                                      (54.59%)

```




## 2022-08-18 Comparing TLB and cache misses between builtin and cay

Honestly, it does not seem to be related to TLB misses, as cay has a much lower rate.

On the other hand mem-stores is a lot higher than builtin. For 1M, cay is at 2Bn mem-stores, whereas builtin is at 1.6M.
Maybe I should try to use [toplev](https://github.com/andikleen/pmu-tools), that can visualize what the CPU is doing.

Using perf on a modified version of cay, where we don't do a `__CompareNMask` function call, but instead have a few
bit operations "`idx := uint16(ctrl[0])&uint16(hash>>57&3) + (1)`", emulating that we fetch the control bytes and do
some byte operations on them. It does seem that changing the `(1)` factor has an impact on the latency, which could indicate that some we hit some boundary in the cache line somewhere.

Before (having `__CompareNMask`):
```
$ for TV in 8 1k 32k 512k 1m; do sudo perf stat -e dTLB-load-misses,cache-misses -g ./cay.test  -test.run=^\$ -test.cpu 1 -test.count 1 -test.bench "read_id/$TV/cay"; done
cpu: Intel(R) Core(TM) i7-8665U CPU @ 1.90GHz

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/8/cay':
           131,117      dTLB-load-misses
        14,852,562      cache-misses

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1k/cay':
           121,029      dTLB-load-misses
        14,873,464      cache-misses

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/32k/cay':
           113,144      dTLB-load-misses
        90,630,164      cache-misses

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/512k/cay':
           704,105      dTLB-load-misses
       276,259,860      cache-misses

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1m/cay':
         1,324,025      dTLB-load-misses
       316,184,011      cache-misses
```

After (removing `__CompareNMask`):

```
$ for TV in 8 1k 32k 512k 1m; do sudo perf stat -e dTLB-load-misses,cache-misses -g ./cay.test  -test.run=^\$ -test.cpu 1 -test.count 1 -test.bench "read_id/$TV/cay"; done
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/8/cay':
           116,947      dTLB-load-misses
        14,151,668      cache-misses

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1k/cay':
           124,258      dTLB-load-misses
        15,277,182      cache-misses

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/32k/cay':
           120,314      dTLB-load-misses
        54,505,594      cache-misses

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/512k/cay':
           738,956      dTLB-load-misses
       258,818,526      cache-misses

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1m/cay':
         1,470,155      dTLB-load-misses
       465,317,721      cache-misses
```


As a baseline these are the results from builtin:

```
$ for TV in 8 1k 32k 512k 1m; do sudo perf stat -e dTLB-load-misses,cache-misses -g ./cay.test  -test.run=^\$ -test.cpu 1 -test.count 1 -test.bench "read_id/$TV/built"; done
...
 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/8/built':
           118,019      dTLB-load-misses
        14,633,452      cache-misses

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1k/built':
           125,581      dTLB-load-misses
        14,527,774      cache-misses

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/32k/built':
           133,197      dTLB-load-misses
        53,978,241      cache-misses

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/512k/built':
        15,324,337      dTLB-load-misses
       311,375,476      cache-misses

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1m/built':
        16,594,415      dTLB-load-misses
       344,985,919      cache-misses
```

For the low number, cay has higher tlb misses and cache misses, and for the 1M size, cay has higher cache-miss.





## 2022-08-12: Investigating TLB-misses due to function calls

The `__CompareNMask` function call is using the old stack-based convention, where arguments are pushed and pulled
from the stack. Given these memory stores and loads, we might hit the TLB, so I want to compare the TLB-miss rate due
to :

- Stack-based function calls
- Register-based function calls
- No function call.

So far no conclusions on these functions calls, except in my dev environment the perf numbers looks reasonable:
```
$ go test -run=^\$ -cpu 1 -count 1 -bench 'function_call' -cpuprofile cpu.profile                                                             10:57 12/08
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
cpu: AMD EPYC 7B13
Benchmark_function_call_tlb_miss_rate/no_function_call         	52123150	        23.07 ns/op
Benchmark_function_call_tlb_miss_rate/register-based_function_call         	40125406	        30.02 ns/op
Benchmark_function_call_tlb_miss_rate/stack-based_ASM_function_call        	29837817	        41.30 ns/op
```

but on my Linux env (a laptop), the no function call is completely off (twice as slow as register-based one):

```
$ ./cay.test  -test.run=^\$ -test.cpu 1 -test.count 1 -test.bench 'function_call'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
cpu: Intel(R) Core(TM) i7-8665U CPU @ 1.90GHz
Benchmark_function_call_tlb_miss_rate/no_function_call         	35837146	       115.0 ns/op
Benchmark_function_call_tlb_miss_rate/register-based_function_call         	34069237	        52.23 ns/op
Benchmark_function_call_tlb_miss_rate/stack-based_ASM_function_call        	32606010	        47.48 ns/op
PASS
```

I'm unsure why that is.


#### Caymap and TLB misses
Further diving into the TLB misses for 1M case, we do see that after the `__CompareNMask` call and it access the key of an entry in the bucket 31% of cases have a TLB miss:
```
Sorted summary for file /home/cinnamon/jakob/cay.test
----------------------------------------------

   61.51 map_get.go:49 (# Is `ctrl := bucket.controls`, which makes sense)
   31.95 map_get.go:69 (# Is `ctrl := bucket.controls`, thus access the key of the )
    1.39 map_get.go:91
    1.32 map_get.go:68
    1.27 map_get.go:56
 Percent |      Source code & Disassembly of cay.test for dTLB-load-misses (1833 samples, percent: local period)
----------------------------------------------------------------------------------------------------------------
         :            Disassembly of section .text:
         :
         :            000000000059b920 <github.com/jakobgt/cay.(*Map[go.shape.[]uint8_0]).findGet>:
         :            github.com/jakobgt/cay.(*Map[go.shape.[]uint8_0]).findGet():
```

Looking into the builtin map access, we only see one place, where the TLB misses are:

```
Sorted summary for file /home/cinnamon/jakob/cay.test
----------------------------------------------

   95.42 map_faststr.go:192
    0.78 map_faststr.go:195
 Percent |      Source code & Disassembly of cay.test for dTLB-load-misses (4901 samples, percent: local period)
----------------------------------------------------------------------------------------------------------------
         :            Disassembly of section .text:
         :
         :            0000000000413ce0 <runtime.mapaccess2_faststr>:
         :            runtime.mapaccess2_faststr():

```

What is interesting is if I just run perf and record the TLB load misses, there's a factor 10 in favor of caymap? So maybe it is something else than TLB:

```
cinnamon@cinnamon-testpod:~/jakob$ sudo perf stat -e dTLB-load-misses -g ./cay.test  -test.run=^\$ -test.cpu 1 -test.count 1 -test.bench 'read_id/1m/built'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
cpu: Intel(R) Core(TM) i7-8665U CPU @ 1.90GHz
Benchmark_read_identical_string_keys/1m/builtin 	14916814	        89.78 ns/op
PASS

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1m/built':

        16,202,946      dTLB-load-misses  #Builtin has 16M tlb-misses

       3.195433605 seconds time elapsed

       2.721868000 seconds user
       0.545967000 seconds sys


cinnamon@cinnamon-testpod:~/jakob$ sudo perf stat -e dTLB-load-misses -g ./cay.test  -test.run=^\$ -test.cpu 1 -test.count 1 -test.bench 'read_id/1m/cay'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
cpu: Intel(R) Core(TM) i7-8665U CPU @ 1.90GHz
Benchmark_read_identical_string_keys/1m/caymap  	14394189	        95.27 ns/op
PASS

 Performance counter stats for './cay.test -test.run=^$ -test.cpu 1 -test.count 1 -test.bench read_id/1m/cay':

         1,236,577      dTLB-load-misses   #caymap has 1.2M tlb-misses

       3.541897468 seconds time elapsed

       3.074113000 seconds user
       0.529052000 seconds sys
```



## 2022-08-03: More deep-diving on the TLB effect of the __CompareNMask function call

TODO: Check up on
- TLB misses for the 1M case.
- the fill-ratio

Looked into
- TLB misses around the __CompareNMask call (more verification there)
- Checked whether the Go tracing could be helpful (it did) - I made the test cases more sane.


### TLB misses

By fetching the bucket.entries before and after the CompareNMask call, we can see where fetching the memory
costs something. And it does seem that fetching bucket.entries after the call does cost something.

This is from `findGet` via `go test -run=^\$ -cpu 1 -count 1 -bench 'read_id/8/caymap' -cpuprofile cpu.profile`
```
    53            .          .           		bEntries := &bucket.entries
    54         50ms      290ms           		idx := __CompareNMask(grpCtrlPointer, unsafe.Pointer(hash>>57))
    55         70ms       70ms           		bEntries = &bucket.entries
```
In the above we can see that the first `bEntries := &bucket.entries` does not cost anything, whereas the second `bEntries = &bucket.entries` does cost 70ms (out of roughly 690ms inside this function, or 10%).

TODO: Look into what perf says about the above

### Go traces
Looking at the Go traces it was obvious that the setup of the tests, where each subtest would create random keys, everytime it was run, was not beneficial. So I refactored the tests to create random key slices for each required size and use that.

These were the commands:
```
$ go test -trace trace_cay.out -run=^\$ -cpu 1 -count 1 -bench 'read_not_found_dyn/1m/caymap'
# Start trace:
$ go tool trace --http :1235 trace_cay.out
```
(Do similarly for builtin)

Running the benchmarks again we get:

```
# Not found cases:
$ go test -run=^\$ -cpu 1 -count 10 -bench 'read_not_found' | tee ~/cay_not_found_results.txt
$ benchstat ~/cay_not_found_results.txt
name                                     time/op
_read_not_found_dynamic/8/caymap         14.1ns ± 2%
_read_not_found_dynamic/8/builtin        12.1ns ± 3%
_read_not_found_dynamic/1k/caymap        17.1ns ± 1%
_read_not_found_dynamic/1k/builtin       18.1ns ± 3%
_read_not_found_dynamic/32k/caymap       24.3ns ± 9%
_read_not_found_dynamic/32k/builtin      26.8ns ± 5%
_read_not_found_dynamic/512k/caymap      65.3ns ± 9%
_read_not_found_dynamic/512k/builtin     78.5ns ± 9%
_read_not_found_dynamic/1m/caymap        80.7ns ± 3%
_read_not_found_dynamic/1m/builtin       86.8ns ±13%
_read_not_found_static_key/8/caymap      13.5ns ± 3%
_read_not_found_static_key/8/builtin     12.3ns ± 4%
_read_not_found_static_key/1k/caymap     13.2ns ± 2%
_read_not_found_static_key/1k/builtin    14.7ns ± 3%
_read_not_found_static_key/32k/caymap    13.3ns ± 2%
_read_not_found_static_key/32k/builtin   14.7ns ± 2%
_read_not_found_static_key/512k/caymap   13.6ns ± 3%
_read_not_found_static_key/512k/builtin  14.9ns ± 4%
_read_not_found_static_key/1m/caymap     13.3ns ± 2%
_read_not_found_static_key/1m/builtin    14.7ns ± 5%
```
Caymap is faster for all cases, except the 8byte cases.

For the found case, builint is faster in all cases, except 32K keys:
```
$ go test -run=^\$ -cpu 1 -count 10 -bench 'read_iden' | tee ~/cay_identical_found_results.txt
$ benchstat ~/cay_identical_found_results.txt
name                                      time/op
_read_identical_string_keys/8/caymap      18.0ns ± 4%
_read_identical_string_keys/8/builtin     10.6ns ± 1%
_read_identical_string_keys/1k/caymap     21.4ns ± 3%
_read_identical_string_keys/1k/builtin    14.9ns ± 2%
_read_identical_string_keys/32k/caymap    31.8ns ± 3%
_read_identical_string_keys/32k/builtin   39.3ns ± 1%
_read_identical_string_keys/512k/caymap    100ns ± 6%
_read_identical_string_keys/512k/builtin  94.7ns ± 4%
_read_identical_string_keys/1m/caymap      118ns ± 6%
_read_identical_string_keys/1m/builtin     100ns ± 6%
```
I wonder if we look into the fill ratio to see if the amount of memory used for caymap vs. builtin could
be interesting.

## 2021-03-30: Try memory ballast
Adding some memory ballast in the form of
```
var (
	_ballast = make([]byte, int64(1)<<int64(33)) // 8GB
)
```

does not change anything unfortunately.

Changing the number of entries in the map makes a big difference, as for all maps smaller than `1<<15`, simdmap is faster on both Mac and Linux. And it is faster always.

Maybe the type should be changes to `map<string, int>`, which is what Matt Kulundis from Google is using to benchmark his types. He has two set-ups, one with a 4 byte key/value and one with 64 bytes (he is only looking at sets, where the key and value is the same).


## 2020-01-14: Look into the cache misses

Generally, where the code is using the most cpu-cycles is also where there are cache misses, so I want to know what type of cache miss it is. E.g., TLB/L1/L2, etc. On the host, the L1 is 32K.

Running
```
./perf record -e L1-dcache-load-misses -g ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*simd.*get/same_keys.*'
```
gives that `if keyP.Len != ekeyP.Len {` (`map_get:75`) has:
- 80% of the L1 cache misses in `find_get` (2381 samples)
- 70% of the LLC cache misses. (LLC = Last Level Cache, not sure whether that is L2 or L3LLC-load-misses) (2353 samples)

`__CompareNMask` has a similar level of L1 cache misses, but not LLC misses (only 600 samples).

This is really weird, as the host has a 32KB L1 cache, 256KB L2 cache and a 20MB L3 cache and that specific line gets flushed. Could it be the GC that flushes the caches?



TODO: Try to disable GC during the test runs. That might avoid the GC. Look for deltablue benchmarks, that were used in V8/Dart, etc.


## 2020-01-13 Caching the returned pointer

It seems that we can't optimize `find` that much more, where we have to accept the second round of TLB misses, when inspecting the length of the keys. Maybe we can optimize the case where the key is found and then making sure that, once we return the value pointer we don't hit a third TLB miss there?

Inspecting the code with
```
./perf record -e cache-misses -g ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*simd.*get/same_keys.*'
```
it seems that all the cache misses happens where there are also TLB misses, including the return call in Get().

Two ideas: a) Move key and value closer, and/or b) reduce function calls and LOC between returning from find and using the byte[]

a) Idea: pair key and value together.
-------------------------------------

With
```
type entry struct {
	key   string
	value []byte
}
```
for each entry it seems that we get a lot fewer TLB misses in the return statement (~248 samples, whereas the ekeyP.Len generates ~1K TLB page misses). This could be interesting.

As a side note, SIMD get/same keys is often times, faster than the builtin:
```
$ ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 30s -test.bench '.*get/same_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	199967498	       194 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys     	202134694	       184 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1043 are full
PASS
$ ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10s -test.bench '.*get/same_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	86886939	       143 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys     	100000000	       125 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1081 are full
PASS
$ ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10s -test.bench '.*get/same_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	84544454	       149 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys     	100000000	       124 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1120 are full
PASS
$ ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10s -test.bench '.*get/same_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	94161273	       154 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys     	100000000	       172 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1109 are full
PASS
$ ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10s -test.bench '.*get/same_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	93228235	       145 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys     	98207223	       122 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1106 are full
PASS
```

## 2020-01-11 Stack-allocating the bucket

If using `grp := m.buckets[cGroup]`, I would have hoped that the bucket would have been allocated on the stack. From the generated byte code, it does not seem to be case, as `runtime.newobject` is called along with `runtime.duffcopy` and ` runtime.typedmemmove`, so maybe the Go compiler thinks that the bucket escapes and thus is allocated on the heap. (`newobject` is the same as malloc).

```
 155            .      2.17s           		grp := m.buckets[cGroup]
                   .          .  11e58b0:                     LEAQ runtime.rodata+297792(SB), AX                           map.go:155
                   .          .  11e58b7:                     MOVQ AX, 0(SP)                                               map.go:155
                   .          .  11e58bb:                     NOPL 0(AX)(AX*1)                                             map.go:155
                   .      1.13s  11e58c0:                     CALL runtime.newobject(SB)                                   map.go:155
 ...
 (AX*1)                                             map.go:155
                   .          .  11e58e0:                     CMPQ CX, BX                                                  map.go:155
                   .          .  11e58e3:                     JAE 0x11e5995                                                map.go:155
                   .          .  11e58e9:                     MOVQ DI, 0x48(SP)                                            map.go:155
                   .          .  11e58ee:                     SHLQ $0xa, BX                                                map.go:155
                   .          .  11e58f2:                     LEAQ 0(DX)(BX*1), SI                                         map.go:155
                   .          .  11e58f6:                     CMPL $0x0, runtime.writeBarrier(SB)                          map.go:155
                   .          .  11e58fd:                     JNE 0x11e596f                                                map.go:155
                   .          .  11e58ff:                     NOPL                                                         map.go:155
                   .          .  11e5900:                     MOVQ BP, -0x10(SP)                                           map.go:155
                   .          .  11e5905:                     LEAQ -0x10(SP), BP                                           map.go:155
                   .      280ms  11e590a:                     CALL runtime.duffcopy(SB)                                    map.go:155
                   .          .  11e590f:                     MOVQ 0(BP), BP                                               map.go:155
...
                   .      760ms  11e5984:                     CALL runtime.typedmemmove(SB)                                map.go:155
```

Btw., `go build -gcflags '-m' . ` is good for seeing what escapes to the heap. By ensuring that the `grp` variable does not escape, I reduce the ns/ops from 1100ns to 200ns. And thus, no mallocs and `typedmemmove`. Only a `duffcopy` call.

Making sure that `grp` does not escape, we see a different picture, where the tlb-misses are located in (percentages)
```
 47.10 map.go:138
   14.73 map.go:231
   14.66 map.go:155
   10.13 map.go:153
    5.17 map.go:236
    3.97 map.go:151
    2.34 map.go:150
```
where
 - `map.go:138` is `hash := z.Memhash(keyP.Str, 0, uintptr(keyP.Len))`
 - `map.go:231` is `uintptr(1) << (b & (PtrSize*8 - 1))`, where `b` is `m.logSize`, so I guess a lookup of the `m` variable
 - `map.go:155` is `grp := m.buckets[cGroup]`

So maybe the duffcopy is destroying the TLB cache? Looking in the duffcopy code, it has 14610 samples, whereas the `find` method has 1412 samples. Thus duffcopy seems to kill the TLB. Thus, can we load the TLB for the full page of `grp`, without a `duffcopy` call? It would be interesting to see whether the `grp` is page aligned? Or rather can we just copy the control byte array, which causes a copy of the array, but a full load of the page into the TLB?

By forcing it to copy the control byte array onto the stack, I've seen some movups, thus the compiler can't see that the data is byte-aligned (MOVUPS is Move Unaligned Packed Single-Precision). Furthermore copying a byte array is not the same as copying a string array (the latter leads to a duffcopy call).

Confirmed with the test case `Test__FirstBucketIsPageAligned`, that the buckets are page-aligned (as in the first bucket is). Thus, the TLB misses are not coming from misaligned loads. Thus need to figure out why there is a TLB miss, when accessing `unsafe.Pointer(&grp.keys[i])`.

Maybe the second TLB miss, which stands for 13% of the misses in `find()` is okay? Can I find optimizations somewhere else?

## 2021-01-08 Debugging TLB misses for get tests

Not-found tests
---------------

For the not-found tests, TLB misses are the same for both versions:
```
$ ./linux-perf stat  -e dTLB-loads,dTLB-load-misses,dTLB-prefetch-misses ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*builtin.*not_found/dynamic_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get_not_found/dynamic_keys         	10000000	       168 ns/op	       0 B/op	       0 allocs/op
PASS

 Performance counter stats for './simdmap.test -test.run=^$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench .*builtin.*not_found/dynamic_keys.*':

     2,278,803,379      dTLB-loads:u
        12,915,360      dTLB-load-misses:u        #    0.57% of all dTLB cache hits  # <- check this number and compare below, both are at 0.5% TLB misses
   <not supported>      dTLB-prefetch-misses:u

       6.397644517 seconds time elapsed

$ ./linux-perf stat  -e dTLB-loads,dTLB-load-misses,dTLB-prefetch-misses ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*simd.*not_found/dynamic_keys.*'
Buckets: 131072, of which 1060 are full
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_simdmap_get_not_found/dynamic_keys         	10000000	       117 ns/op	       0 B/op	       0 allocs/op
PASS

 Performance counter stats for './simdmap.test -test.run=^$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench .*simd.*not_found/dynamic_keys.*':

     2,474,095,181      dTLB-loads:u
        13,366,113      dTLB-load-misses:u        #    0.54% of all dTLB cache hits    # <- check this number and compare above, both are at 0.5% TLB misses
   <not supported>      dTLB-prefetch-misses:u

       5.389249197 seconds time elapsed
```

Found tests
-----------
And it seems that found tests are also better in the simdmap case (at least once):

$ ./linux-perf stat  -e dTLB-loads,dTLB-load-misses,dTLB-prefetch-misses ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*get/same_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	10000000	       173 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys     	10000000	       131 ns/op	       0 B/op	       0 allocs/op # <- Simdmap is faster!!!!!
Buckets: 131072, of which 1083 are full
PASS

 Performance counter stats for './simdmap.test -test.run=^$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench .*get/same_keys.*':

     8,018,199,273      dTLB-loads:u
        30,158,881      dTLB-load-misses:u        #    0.38% of all dTLB cache hits
   <not supported>      dTLB-prefetch-misses:u

      16.724325756 seconds time elapsed

$ ./linux-perf stat  -e dTLB-loads,dTLB-load-misses,dTLB-prefetch-misses ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*builtin.*get/same_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	10000000	       132 ns/op	       0 B/op	       0 allocs/op
PASS

 Performance counter stats for './simdmap.test -test.run=^$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench .*builtin.*get/same_keys.*':

     3,547,374,496      dTLB-loads:u
        16,341,849      dTLB-load-misses:u        #    0.46% of all dTLB cache hits
   <not supported>      dTLB-prefetch-misses:u

       7.503344017 seconds time elapsed

$ ./linux-perf stat  -e dTLB-loads,dTLB-load-misses,dTLB-prefetch-misses ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*simd.*get/same_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_simdmap_get/same_keys 	10000000	       133 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1065 are full
PASS

 Performance counter stats for './simdmap.test -test.run=^$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench .*simd.*get/same_keys.*':

     3,870,569,488      dTLB-loads:u
        14,451,216      dTLB-load-misses:u        #    0.37% of all dTLB cache hits
   <not supported>      dTLB-prefetch-misses:u

       8.990690540 seconds time elapsed

Furthermore, line 179 in map.go (`if (keyP == ekeyP || keyP.Str == ekeyP.Str || z.Memequal(ekeyP.Str, keyP.Str, uintptr(keyP.Len))`) is causing a relative high number of TLB misses. It accounts for 31% of the TLB-loads, but accounts for 68% of the dTLB-load-misses. Similarly line 155 (`idx := __CompareNMask(grpCtrlPointer, unsafe.Pointer(hash>>57))`) accounts for 1.6% of the loads, but 16% of the dTLB-load-misses. I'll try to add
```
			if keyP.Len != ekeyP.Len {
				continue
			}
```
before line 179 to see if that changes the load-misses.

With the above, the above now has 30% of the dtlb-loads and 70%(!) of the dtlb-load-misses. Thus referencing `keyP.Len` or `ekeyP.Len` results in TLB misses. Based on my reading of the generated code the culprit is the `ekeyP := (*z.StringStruct)(unsafe.Pointer(&grp.keys[i]))` that generates a bunch of extra instructions, and the question is whether these instructions mess up the TLB cache. Should I try to write these as an `unsafe` array computation?

Using unsafe array computations
-------------------------------
Interesting, replacing `ekeyP := (*z.StringStruct)(unsafe.Pointer(&grp.keys[i]))` with `ekeyP := (*z.StringStruct)(add(unsafe.Pointer(&grp.keys), uintptr(i)*stringSize))`, keeps the tlb-misses at a LEA (load-effective address) instruction (see below). Thus, it seems that `&grp.keys` causes TLB-misses...
```
 map.go:175    0.99 :     5b0d54:       lea    (%rcx,%rdx,1),%rdi
    0.00 :        5b0d58:       lea    0x10(%rdi),%rdi
    0.05 :        5b0d5c:       mov    %rsi,%rax
    0.24 :        5b0d5f:       shl    $0x4,%rsi
    0.24 :        5b0d63:       mov    0x78(%rsp),%r8
    0.00 :        5b0d68:       mov    0x8(%rdi,%rsi,1),%r9
         :      github.com/jakobgt/cay.add():
 map.go:242   68.16 :     5b0d6d:       lea    (%rdi,%rsi,1),%r10 # <- this one in the add function
         :      github.com/jakobgt/cay.(*Map).find():
    0.00 :        5b0d71:       cmp    %r8,%r9
    0.00 :        5b0d74:       jne    5b0d39 <github.com/jakobgt/cay.(*Map).find+0x79>
```

Hence next step is to figure out whether I really need `grp := &m.buckets[cGroup]` or it should be `grp := m.buckets[cGroup]`


## Run 2020-11-30


```
 $ go version                                                                                                            20:32 30/11
go version go1.15.2 darwin/amd64
```
Local Mac (all tests are green)

Run (simdmap wins):
```
$ go test -run=^\$ -cpu 1 -benchmem -bench '.*'                                                                         20:38 30/11
goos: darwin
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_string_len                    	1000000000	         0.532 ns/op	       0 B/op	       0 allocs/op
Benchmark_string_len_reverse            	772742073	         1.56 ns/op	       0 B/op	       0 allocs/op
Benchmark_string_struct_len_reverse     	761126870	         1.55 ns/op	       0 B/op	       0 allocs/op
Benchmark_builtin_map_get/same_keys     	11334487	       113 ns/op	       0 B/op	       0 allocs/op
Benchmark_builtin_map_get/same_but_fresh_keys         	 7499438	       166 ns/op	       0 B/op	       0 allocs/op
Benchmark_builtin_map_get/same,_fresh_and_random_order_keys         	 4416318	       286 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys                                     	 9499071	       154 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_but_fresh_keys                           	 7277200	       174 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same,_fresh_and_random_order_keys             	 4206589	       313 ns/op	       0 B/op	       0 allocs/op
Benchmark_builtin_map_get_not_found/static_key                      	92947183	        13.4 ns/op	       0 B/op	       0 allocs/op
Benchmark_builtin_map_get_not_found/dynamic_keys                    	 9785622	       128 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1060 are full
Benchmark_simdmap_get_not_found/static_key                          	100000000	        11.8 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get_not_found/dynamic_keys                        	14061108	        94.3 ns/op	       0 B/op	       0 allocs/op
```

For ` ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*not_found/dynamic_keys.*'` Instruments shows:

```
L2 Hit:               ~8mm for simdmap (vs. 3mm for builtin)
DTLB walk completed:  ~10.2mm for simdmap (vs. 10.5 for builtin)
DTLB STLB hit:        ~3.3mm for simdmap (vs. 9.9mm for builtin)
Mispredicted branches ~0.9mm for simdmap (vs. 12mm for builtin)
```
seems like builtin just requests more data for not found.

For `-test.run=^$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench .*get/same_but_fresh_keys.*` Intruments reports
```
L2 miss:               ~38.6mm for simdmap (vs. 35.4 for builtin)
Cycles:                ~4bn for simdmap (vs. 3.9bn for builtin)
Stalled cycles:        ~11.5bn for simdmap (vs. 9.9bn for builtin)
Mispredicted branches: ~0.4mm for simdmap (vs. 9.5mm for builtin)
```
Thus, simdmap is much more predictable for branch predictions, but the memory fetches seems to be worse.

Byte-aligned map
----------------
Making the `bucket` byte-aligned (by having a filler byte array) seems to improve performance of the simdmap

Instruments command
```
-test.run=^$ -test.cpu 1 -test.benchmem -test.benchtime 30000000x -test.cpuprofile cpu.profile -test.bench .*get/same_keys.*
```

Non-byte aligned:
```
testing: open cpu.profile: read-only file system
goos: darwin
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	30000000	       114 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys     	30000000	       143 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1112 are full
PASS
```

Byte-aligned:
```
testing: open cpu.profile: read-only file system
goos: darwin
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	30000000	       110 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys     	30000000	       134 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1115 are full
PASS
<End of Run>
```

Comparing Instrument counters:
Non-byte aligned
```
Cpu cycles stalled:    33.867.820.180 for simdmap (vs. 24.266.816.843 for builtin)
L2 Cache misses:          112.878.397 for simdmap (vs. 98.156.923 for builtin)
```

Byte-aligned
```
Cpu cycles stalled:    29.347.750.948 for simdmap (vs. 25.248.233.827 for builtin)
L2 Cache misses:          95.792.491 for simdmap (vs. 98.296.924 for builtin)
```
So byte aligned does matter.


Next up?

With a byte-aligned map, the simdmap implementation for `get/same_keys` has fewer L2D misses, fewer TLB load misses causing a walk, but higher number of instructions that causes a stall:

```
Cycles with a stall: 25.474.354.712 (simdmap) vs. 20.477.061.167 (builtin)
L2 misses:           96.441.083 (simd) vs. 96.441.083 (builtin)
DTLB load misses:    31.015.416	(simd) vs. 34.998.766	(builtin)
```
So, I need to figure out what is causing these load stalls.


Turns out that DTLB is the first-level TBL cache, and STLB is the second-level. Looking at STLB, we see that simdmap does 40% more STLB misses than builtin. The `MEM_INST_RETIRED.STLB_MISS_LOADS` counts this

Args:
```
-test.run=^$ -test.cpu 1 -test.benchmem -test.benchtime 30000000x -test.bench .*get/same_keys.*
```
See Intel CPU events: https://download.01.org/perfmon/index/skylake.html
TLB numbers
```
Load instructions:       1.463.458.848 for SIMD (vs. 1.881.617.539 for builtin)
Loads with an STLB miss: 	  28.976.282 for SIMD (2% of loads - vs. 	21.631.208 for builtin - 1.1%)
DLTB Load misses:           31.020.833 for SIMD (vs. 35.153.593 for builtint)
DLTB miss, but in STLB:      2.789.990 for SIMD (vs. 30.226.279)
```
It thus seems like SIMD has more TLB misses. With `perf` you can see the instructions where those TLBs are, so I should somehow see if I can
get a Linux box to test perf on.

My current thesis is the more bound checks in the simd code, causes more TLB cache invalidations.

## Run date not known


- Not inlining a function can add up to 4ns
- Adding `-cpuprofile cpu.profile ` adds a few ns.
- Using `groupMask(...)` to generate the group mask over (hash & m.Mask / 16) shaves off a ns/op (and the mask generate is dropped by 4.5x from 70ms to 20ms).
  Golang then uses the `MOVZX` operation over the `MOVQ`.
    Code went from
```
   148         70ms       70ms           	sGroup := hash & m.slotMask / 16
                40ms       40ms  129b7db:             MOVQ 0x90(SP), CX                                                    map.go:148
                20ms       20ms  129b7e3:             MOVQ 0x18(CX), DX                                                    map.go:148
                   .          .  129b7e7:             ANDQ AX, DX                                                          map.go:148
                10ms       10ms  129b7ea:             SHRQ $0x4, DX                                                        map.go:148

```
  To:

```
     150         20ms       20ms           	groupMask := bucketMask(m.logSize)
                 20ms       20ms  129b7fb:             MOVQ 0x90(SP), CX                                                    map.go:150
                    .          .  129b803:             MOVZX 0x18(CX), DX                                                   map.go:150

     151            .          .           	sGroup := hash & groupMask // Equal to hash & m.slotMask / 16
                    .          .  129b818:             ANDQ AX, CX
```
- The native hashmap does a lot to avoid bound checks (e.g., for bitmasks it uses uintptr that)
- The max key size for builtin map is 128 (bits/bytes?) before it is not inlined. Similarly for an elem size.
- Letting `__CompareNMask` return an int instead of returning via an argument does not change a lot (maybe 1 ns/op or so.)
- ~0.8% of the SIMD buckets are full (~1k out of 131k bucket) causing ~10ns extra time per op, if we need to search the next bucket, instead
  of stopping after searching the first bucket.
- Having different sized keys result in a drop of 12ns/op (from 100ns to 88ns) just as we avoid comparing keys.
- Wow, the caching nature of an open addressing hashmap is crazy (20ns/op for varying keys with SIMD vs. 150 for builtin)
  The differing factor is that it 7x more to fetch the buckets from memory (specifically the `MOVZX 0(BX)(CX*1), R8 ` operation on `map_faststr.go:191`.
```
~/g/s/c/i/p/l/simdmap (simdmap) $ go test -run=^\$ -benchmem -cpu 1 -bench '.*not_found.*'                                                                         17:35 17/03
goos: darwin
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_simdmap_get_not_found/static_key         	85533632	        14.3 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get_not_found/dynamic_keys       	59230143	        19.4 ns/op	       0 B/op	       0 allocs/op
Benchmark_builtin_map_get_not_found/static_key     	87805620	        13.8 ns/op	       0 B/op	       0 allocs/op
Benchmark_builtin_map_get_not_found/dynamic_key    	 8325208	       149 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/jakobgt/cay	8.869s
```
