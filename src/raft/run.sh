#!/bin/bash
export PATH="$PATH:/usr/local/go/bin"

rm res -rf
mkdir res


test_list=(
TestBasicAgree2B
# TestRPCBytes2B
TestFailAgree2B 
TestFailNoAgree2B 
TestConcurrentStarts2B

# TestRejoin2B
# TestBackup2B

TestCount2B
)
for i in `seq 1000` # 跑2次

do
 sleep 1  
 if [ $(($i % 10)) -eq 0 ]
 then
     echo $i
 fi 
 for j in ${test_list[@]}
 do
     VERBOSE=1 go test -run $j &> ./res/$j-$i
    #  VERBOSE=1 go test -run  2A   &> ./res/$j-$i
     if [ $? -ne 0 ]
     then
         echo "run failed. 文件:"$j-$i
        #  exit  
     else
        #  echo "run success 文件:"$j-$i
         rm ./res/$j-$i
     fi
 done
done