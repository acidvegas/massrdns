# MASSRDNS

## Reverse DNS Lookup Tool

This tool provides an efficient way to perform reverse DNS lookups on IP addresses, especially useful for large IP ranges. It uses concurrent workers and distributes the work among them to achieve faster results.

### Building the Project

1. Clone the repository:

```
git clone https://github.com/acidvegas/massrns
cd massrdns
```

2. Build the project:

```
go build -o massrdns
```

This will produce an executable named `massrdns`.

### Usage

The tool requires two main arguments:

- `-cidr`: The IP address CIDR range you want to perform reverse DNS lookup on.
- `-dnsfile`: The path to a file containing DNS servers, one per line.

Optional arguments:

- `-concurrency`: The number of concurrent workers for reverse DNS lookup. Default is 10.

Example:

```
./rdns_lookup -cidr "192.168.0.0/24" -dnsfile "dns_servers.txt"
```

### DNS Servers File Format

The file should contain one DNS server per line, e.g.:

```
8.8.8.8:53
1.1.1.1:53
```

___

###### Mirrors
[acid.vegas](https://git.acid.vegas/massrdns) • [GitHub](https://github.com/acidvegas/massrdns) • [GitLab](https://gitlab.com/acidvegas/massrdns) • [SuperNETs](https://git.supernets.org/acidvegas/massrdns)
