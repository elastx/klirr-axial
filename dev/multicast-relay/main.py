import socket
import time
from collections import defaultdict

PORT = 9999
registered_nodes = {}
REGISTRATION_TIMEOUT = 30  # Seconds

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM, socket.IPPROTO_UDP)
sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
sock.bind(('', PORT))  # Listen on all addresses and the relay port

recent_messages = defaultdict(float)
MESSAGE_TIMEOUT = 5.0

print("Relay running, awaiting packets...", flush=True)


def prune_stale_nodes():
    """Remove stale nodes."""
    now = time.time()
    stale_nodes = [
        ip for ip, last_seen in registered_nodes.items()
        if now - last_seen > REGISTRATION_TIMEOUT
    ]
    for ip in stale_nodes:
        print(f"Removing stale node: {ip}", flush=True)
        del registered_nodes[ip]


while True:
    prune_stale_nodes()
    data, address = sock.recvfrom(1024)
    now = time.time()

    # Register node
    if address[0] not in registered_nodes:
        print(f"New node joined: {address[0]}", flush=True)
    registered_nodes[address[0]] = now

    message_id = (data, address)
    if message_id in recent_messages and now - recent_messages[
            message_id] < MESSAGE_TIMEOUT:
        continue
    recent_messages[message_id] = now

    # Relay to all nodes
    for node_ip in registered_nodes:
        if node_ip != address[0]:
            sock.sendto(data, (node_ip, PORT))

    print(
        f"Relayed message from {address} to nodes: {list(registered_nodes.keys())}",
        flush=True)
