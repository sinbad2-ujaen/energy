#!/bin/bash

# Directory containing post_data files
post_data_directory="./transaction_data"

# Vegeta attack parameters
rate=1
duration=0.1s

if [ $# -ne 1 ]; then
    echo "Usage: $0 <port>"
    exit 1
fi

target_file="./target_$1.list"

# Ensure the post data directory exists
if [ ! -d "$post_data_directory" ]; then
    echo "Post data directory does not exist: $post_data_directory"
    exit 1
fi

# Ensure the target file exists
if [ ! -f "$target_file" ]; then
    echo "Target file does not exist: $target_file"
    exit 1
fi

# Function to execute Vegeta attack for a single file
run_vegeta_attack() {
    post_data_file="$1"
    filename=$(basename -- "$post_data_file")
    filename_no_extension="${filename%.*}"

    echo "Running Vegeta attack on: $post_data_file"

    # Run Vegeta attack for this specific post_data_file
    vegeta attack -targets="$target_file" -rate="$rate" -duration="$duration" -body="$post_data_file" \
    | tee "results/results_$filename_no_extension.bin"

    # Check if Vegeta command succeeded
    if [ $? -ne 0 ]; then
        echo "Vegeta attack failed for file: $post_data_file"
        exit 1
    fi
}

# Export the function and required environment variables for parallel execution
export -f run_vegeta_attack
export target_file rate duration post_data_directory

# Ensure no duplicates in the files being processed
find "$post_data_directory" -name "transaction_$1_*.json" | sort | uniq | parallel -j 32 run_vegeta_attack {}