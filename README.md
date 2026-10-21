# LocalFileSender

The alpha version TCP tool for transferring files between computers.

## Features

- Send a file to another computer over TCP.
- Receive a file by starting the receiver before the sender.
- Build an executable from source with Go.

## Requirements

- [Go](https://go.dev/dl/) installed to run or build the project from source.
- Network connectivity between the computers.
- The receiving computer must allow incoming connections on the selected TCP port.

## Usage

Start the receiver first, then run the sender on the other computer.

### Receive a file

On the computer that will receive the file:

>```bash
># you can use another port if you want 
>go run . receive :8080 filename.txt 
>```

### Send a file

On the computer that will send the file:

>```bash
>go run . send ip_addr:8080 filename.txt
>```

Replace `ip_addr` with the receiving computer's IP address or hostname, and replace `filename.txt` with the path to the file you want to send.

The IP address above is an example. Use the actual address of the receiving computer.

## Build

### Linux

Build the executable and place it in `~/.local/bin`:

>```bash
>mkdir -p ~/.local/bin
>go build -o ~/.local/bin/lfs .
>```

Make sure `~/.local/bin` is included in your `PATH` if you want to run `lfs` without typing its full path.

Examples using the built executable:

>```bash
>lfs receive :8080 filename.txt     # 1ts PC
>lfs send ip_addr:8080 filename.txt # 2nd PC
>```

## Firewall

The receiving computer must allow incoming TCP connections on the selected port. For port `8080`, using UFW:

>```bash
>sudo ufw allow 8080/tcp
>```

When you no longer need this rule, remove it with:

>```bash
>sudo ufw delete allow 8080/tcp
>```

## Troubleshooting

- **The sender cannot connect:** Check that the receiver is running, the IP address and port are correct, and both computers can reach each other over the network.
- **The connection fails or times out:** Check the receiving computer's firewall and confirm that the receiver is listening on the selected port.
- **The `lfs` command is not found:** Check that `~/.local/bin` is in your `PATH`, or run the executable using its full path.

## Security

TCP alone does not encrypt transferred data or authenticate the other computer. Use LocalFileSender only on networks you trust unless additional security has been implemented.

## Development notes

As the project develops, this README can be expanded with details about supported file sizes, overwrite behavior, transfer integrity checks, encryption, and supported operating systems.
