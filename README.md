# EZ-Monitor

EZ-Monitor is a Linux system monitoring tool that uses SSH connections to query for information
from a set of provided hosts. This allows you to get valuable information without the need
for a dedicated agent on every server.

![demo.gif](./docs/demo/demo.gif)

## Usage

All you need to get started is to define an inventory file of and then start up EZ-Monitor via the CLI.

See an example inventory file below

```ini
[host-1]
address=ubuntu-server-1
username=some-user
ssh_private_key_file=~/.ssh/id_rsa

[host-2]
address=ubuntu-server-2
username=some-user
password=lets-avoid-defining-passwords-in-plain-text-if-we-can-:)
port=23

[host-3]
address=rocky-server-1
username=some-user
password=`$EZ_MONITOR_ENCRYPTED;0000000000000000000000007e4a83986d07faed6729d29686b42c7c1e8bc37f`
```

Inventory files are defined in an `ini` format where every `[section]` defined in brackets signifies a host entry. The
section name servers as an alias for each host. From there, the following connection information is supported.

- address
- username
- password
- ssh_private_key_file
- port

Once you have your inventory file defined, simply run `ez-monitor` with a path to your inventory supplied as an argument.

```bash
ez-monitor inventory.ini
```

### Handling Passwords

If you have a host entry that requires you to enter a password, it is strongly encouraged that you encrypt the password
using EZ-Monitor's password encryption functionality.

Simple run

```bash
ez-monitor inventory.ini --add-encrypted-pass alias-of-host-in-inventory-file-to-encrypt
```

Then, follow the prompts to enter both the host's password, and a new encryption password which will be used
for all hosts in this file.

Finally, with the password entered, start EZ-Monitor as you normally would and follow the prompts

```bash
ez-monitor inventory.ini
Please enter your encryption password to decrypt the passwords in this file.
```

## Installation

### MacOS
If you have HomeBrew installed, use the tap shown below.

```bash
brew tap kreulenk/brew
brew install ez-monitor
```

### Linux
Navigate to the Releases section of EZ-Monitor's GitHub repository and download the latest tar for your
processor architecture. Then, untar the executable and move it to `/usr/local/bin/ez-monitor`.

E.g.
```
curl -OL https://github.com/kreulenk/ez-monitor/releases/download/v0.4.0/ez-monitor-linux-amd64.tar.gz
tar -xzvf ez-monitor-linux-amd64.tar.gz
mv ./ez-monitor /usr/local/bin/ez-monitor
```

### Build From Source

Ensure that you have at least Go 1.24 installed on your system.

Then, run
```bash
make install
```

## Development Roadmap
The high level plan for this project is as follows:

| # | Step                                            | Status |
|:-:|-------------------------------------------------|:------:|
| 1 | Support for ini inventory files                 |   ✅   |
| 2 | Display real time data in bar graphs            |   ✅   |
| 3 | Display historical data with line graphs        |   ✅   |
| 4 | Support hashing of passwords in inventory files |   ✅   |
| 5 | Improve the styling of the graphs displayed     |   ❌   |
| 5 | TBD!                                            |   💥   |

