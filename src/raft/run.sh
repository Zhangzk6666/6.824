#!/bin/bash
export PATH="$PATH:/usr/local/go/bin"

rm res -rf
mkdir res


test_list=(
2A)
for i in `seq 2` # 跑2次
do
 if [ $(($i % 10)) -eq 0 ]
 then
     echo $i
 fi
 for j in ${test_list[@]}
 do
     go test -run $j &> ./res/$j-$i
     if [ $? -ne 0 ]
     then
         echo "run failed."
         exit 1i
     fi
 done
done