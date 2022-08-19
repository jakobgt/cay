#!/bin/bash

# set -x
linesToInclude=30
perfDetail=""
if [ $# != 0 ]
  then
    perfDetail="-M ${1}"
fi

for TV in 8 1k 32k 512k 1m
do
  echo "# Benchmark size $TV"
  for type in cay built
    do
      sudo  perf stat $perfDetail -g ./cay.test  -test.run=^\$ -test.cpu 1 -test.count 1 -test.bench "read_id/$TV/$type" 2>&1 | grep -A $linesToInclude "Performance counter" | grep -v "seconds sys" | grep -v "seconds time elapsed" | grep -v "seconds user" | grep -v -e '^$'
  done
  echo ""
done
