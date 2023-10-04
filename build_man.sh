#!/bin/bash

txt2man \
    -t arch-log \
    -v arch-log \
    -r arch-log-$(<VERSION) \
    -s 1 \
    MANUAL.txt > arch-log.1
