#!/bin/bash
PROCESS_PID='3360'
LOGFILE='./resources.txt'

# Define column headers
COLUMNS="%CPU %MEM COMMAND THREADS"

# Write column headers to the logfile
echo "Date Time $COLUMNS" > $LOGFILE

while true; do
    # Collect process metrics and prepend the current date and time
    THREAD_COUNT=$(ps -eLo pid | grep -w $PROCESS_PID | wc -l)
    echo "$(date +'%Y-%m-%d %H:%M:%S')" $(ps -o %cpu,%mem,comm -p $PROCESS_PID | tail -n +2) $THREAD_COUNT >> $LOGFILE
    
    sleep 1 # Adjust as necessary
done