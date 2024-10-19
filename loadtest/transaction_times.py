import json
from datetime import datetime
from dateutil import parser
import numpy as np

# Specify the path to your JSON file (replace 'dag.json' with your actual filename)
file_path = 'dag_8094.json'

# Lists to store time differences
time_added_differences_ms = []
time_confirmation_differences_ms = []

# Read JSON data from the file
with open(file_path, 'r') as file:
    data = json.load(file)

# Iterate through transactions
for transaction_id, transaction_data in data["Transactions"].items():
    timestamp_created_str = transaction_data["TimestampCreated"]
    timestamp_added_str = transaction_data["TimestampAdded"]
    timestamp_confirmed_str = transaction_data["TimestampConfirmed"]

    # Parse timestamp strings to datetime objects
    timestamp_created = parser.parse(timestamp_created_str)
    timestamp_added = parser.parse(timestamp_added_str)
    timestamp_confirmed = parser.parse(timestamp_confirmed_str)

    # Calculate time differences in milliseconds
    time_added_difference_ms = (timestamp_added - timestamp_created).total_seconds() * 1000
    time_confirmation_difference_ms = (timestamp_confirmed - timestamp_added).total_seconds() * 1000

    # Filter out extreme values (adjust the threshold as needed)
    if time_confirmation_difference_ms < 0 or time_confirmation_difference_ms > 1e6:
        continue

    # Append time differences to lists
    time_added_differences_ms.append(time_added_difference_ms)
    time_confirmation_differences_ms.append(time_confirmation_difference_ms)

    print(f"Transaction ID: {transaction_id}")
    print("TimestampCreated:", timestamp_created_str)
    print("TimestampAdded:", timestamp_added_str)
    print("TimestampConfirmed:", timestamp_confirmed_str)
    print("Time Difference (TimestampAdded - TimestampCreated) in ms:", time_added_difference_ms)
    print("Time Difference (TimestampConfirmed - TimestampAdded) in ms:", time_confirmation_difference_ms)
    print("\n")

# Calculate percentiles
p50_added = np.percentile(time_added_differences_ms, 50)
p90_added = np.percentile(time_added_differences_ms, 90)
p95_added = np.percentile(time_added_differences_ms, 95)
p99_added = np.percentile(time_added_differences_ms, 99)

p50_confirmation = np.percentile(time_confirmation_differences_ms, 50)
p90_confirmation = np.percentile(time_confirmation_differences_ms, 90)
p95_confirmation = np.percentile(time_confirmation_differences_ms, 95)
p99_confirmation = np.percentile(time_confirmation_differences_ms, 99)

# Print percentiles
print("\nPercentiles for Time Difference (TimestampAdded - TimestampCreated) in ms:")
print("p50:", p50_added)
print("p90:", p90_added)
print("p95:", p95_added)
print("p99:", p99_added)

print("\nPercentiles for Time Difference (TimestampConfirmed - TimestampAdded) in ms:")
print("p50:", p50_confirmation)
print("p90:", p90_confirmation)
print("p95:", p95_confirmation)
print("p99:", p99_confirmation)
