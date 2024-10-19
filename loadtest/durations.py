import re
import numpy as np

def extract_all_durations(file_path):
    with open(file_path, 'r') as file:
        content = file.read()

    duration_matches = re.finditer(r'Duration\s+\[total, attack, wait\]\s+([\d.]+)ms', content)

    durations = [float(match.group(1)) for match in duration_matches]

    return durations

def calculate_overall_percentiles(durations):
    if durations:
        p50 = np.percentile(durations, 50)
        p90 = np.percentile(durations, 90)
        p95 = np.percentile(durations, 95)
        p99 = np.percentile(durations, 99)

        return p50, p90, p95, p99
    else:
        return None

file_path = 'durations_8090.txt'
all_durations = extract_all_durations(file_path)

if all_durations:
    print("Count: ", len(all_durations))
    print("All Durations: ", all_durations)
    p50, p90, p95, p99 = calculate_overall_percentiles(all_durations)
    print(f"Overall p50: {p50:.2f} ms")
    print(f"Overall p90: {p90:.2f} ms")
    print(f"Overall p95: {p95:.2f} ms")
    print(f"Overall p99: {p99:.2f} ms")
else:
    print("No Duration data found in the file.")
