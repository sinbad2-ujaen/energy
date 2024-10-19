import subprocess
import random
import time
import os
import platform
import string
import json
import os
import binascii
from concurrent.futures import ThreadPoolExecutor
import uuid


def generate_random_transaction():
    transaction_types = ["transaction-standard", "transaction-fast"]

    return {
        "from": "9a27ad25d050b690b05d38e1cf20c71e8c5314cff5a936efc024a2b5e9b07f04",
        "to": generate_master_seed(),
        "token": round(random.uniform(0, 0.1), 4),
        "data": generate_random_json(),
        "type": random.choice(transaction_types)
    }


def generate_master_seed():
    seedLength = 32
    randomBytes = os.urandom(seedLength)
    randomSeed = binascii.hexlify(randomBytes).decode('utf-8')
    return randomSeed

def generate_random_json():
    # Generate a UUID for the 'id' field
    random_id = str(uuid.uuid4())

    # Select a random IoT device type from the list
    iot_devices = ["temperature", "humidity", "motion", "light", "pressure"]
    random_type = random.choice(iot_devices)

    # Generate a random integer between 1 and 10 for the 'value' field
    random_value = random.randint(1, 100)

    # Create the dictionary with the required structure
    data = {
        "id": random_id,
        "type": random_type,
        "value": random_value
    }

    # Convert the dictionary to a JSON string
    json_string = json.dumps(data)

    return json_string


def save_json_to_file(data, filename):
    with open(filename, 'w') as file:
        json.dump(data, file)

def generate_random_port_quantity(sum_total, min_percent=0.25, max_percent=0.35, seed=None):
    if seed is not None:
        random.seed(seed)

    # Calculate the minimum and maximum values based on the total sum
    min_value = int(sum_total * min_percent)
    max_value = int(sum_total * max_percent)

    if max_value * 2 >= sum_total:
        raise ValueError("Invalid percent range: adjust min_percent and max_percent for better distribution.")

    # Generate random values for the first two ports within the defined range
    port8090 = random.randint(min_value, max_value)
    print(f"port8090: {port8090}")
    remaining_total = sum_total - port8090

    port8092 = random.randint(min_value, min(max_value, remaining_total - min_value))
    print(f"port8092: {port8092}")
    remaining_total -= port8092

    # The last port will take the remaining total (ensures the sum equals the original sum_total)
    port8094 = remaining_total
    print(f"port8094: {port8094}")

    # Ensure that the total number of transactions equals sum_total
    assert port8090 + port8092 + port8094 == sum_total, "Transaction count doesn't add up to total"

    return port8090, port8092, port8094


def prepare_body(port, quantity):
    for i in range(quantity):
        post_data_file = f'transaction_data/transaction_{port}_{i}.json'

        random_transaction = generate_random_transaction()
        save_json_to_file(random_transaction, post_data_file)
        print(f"Saved item {i}")


if __name__ == "__main__":
    ports = [8090, 8092, 8094]

    random_port_quantity = generate_random_port_quantity(1000000)

    with ThreadPoolExecutor() as executor:
        executor.map(prepare_body, ports, random_port_quantity)
