import time
from concurrent.futures import ThreadPoolExecutor
from iota.client import Client

# DevNet node - You can find a list of healthy nodes on https://status.iota.org/
node_url = "https://api.testnet.shimmer.network"
client = Client(node_url)

def prepare_transaction(index):
    # Replace with your actual transaction logic (address, amount, message, etc.)
    return client.message(index=index, data=b'Test transaction')

def send_transaction(transaction):
    submit_start_time = time.perf_counter_ns()
    message_id = client.message_submit(transaction)['message_id']
    submit_end_time = time.perf_counter_ns()

    def wait_for_confirmation():
        while True:
            message_metadata = client.get_message_metadata(message_id)
            if message_metadata['solid']:
                return time.perf_counter_ns()
            time.sleep(1)  # Adjust polling interval as needed

    confirmation_end_time = wait_for_confirmation()

    submit_time = (submit_end_time - submit_start_time) / 1e9  # Convert nanoseconds to seconds
    confirmation_time = (confirmation_end_time - submit_start_time) / 1e9

    return message_id, submit_time, confirmation_time

start_time = time.time()

# Use a thread pool with 10 threads for concurrent execution
with ThreadPoolExecutor(max_workers=10) as executor:
    # Submit transactions for execution in the thread pool
    future_results = executor.map(send_transaction, [prepare_transaction(i) for i in range(100)])

    # Collect results from the completed threads
    message_ids, submit_times, confirmation_times = zip(*future_results)

end_time = time.time()

total_submit_time = sum(submit_times)
total_confirmation_time = sum(confirmation_times)

avg_submit_time = total_submit_time / 100
avg_confirmation_time = total_confirmation_time / 100

print(f"Total time for 100 transactions: {end_time - start_time:.2f} seconds")
print(f"Average transaction submit time: {avg_submit_time:.2f} seconds")
print(f"Average transaction confirmation time: {avg_confirmation_time:.2f} seconds")