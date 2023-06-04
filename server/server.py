import socket

host = '0.0.0.0'
port = 8888

server_socket = socket.socket(socket.AF_INET, \
    socket.SOCK_DGRAM)
server_socket.bind((host, port))

while True:
    data, addr = server_socket.recvfrom(1024)
    print(f"Received: {data}, from: {addr}")
    server_socket.sendto(b"OK...hello", addr)
