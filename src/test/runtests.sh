#!/bin/bash

RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m'

total=0
fail=0

filename="out-$1.txt"

for i in {1..1000}
do
    output="$(go test $1)" # run test
    total=$((total+1)) # increment total count
    if [[ $output =~ FAIL ]] 
    then
        fail=$((fail+1))  # increment fail count if needed
    fi


    echo "Test number: " $i >> $filename 
    echo "$output" >> $filename
    echo "" >> $filename

    echo -e "$output" | sed -e ''s/PASS/`printf "${GREEN}PASS${NC}"`/g'' -e ''s/FAIL/`printf "${RED}FAIL${NC}"`/g''
    echo ""
    echo -e "===== ${YELLOW}FAILED ${fail}/${total} TESTS${NC} ====="
    echo ""

    if [[ $output =~ "build failed" ]]
    then
        sleep 2
    fi
    # if [[ $output =~ FAIL ]] 
    # then
    #     exit 1
    # fi
done
