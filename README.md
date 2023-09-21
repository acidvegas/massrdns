# MASSRDNS

## Reverse DNS Lookup Tool

This tool provides an efficient way to perform reverse DNS lookups on IP addresses, especially useful for large IP ranges. It uses concurrent workers and distributes the work among them to achieve faster results. Each request will randomly rotate betweeen the supplied DNS servers to split the load of a large CDIR across many DNS servers.

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
./massrdns -cidr "0.0.0.0/0" -dnsfile "dns_servers.txt"
```

### DNS Servers File Format

The file should contain one DNS server per line, e.g.:

```
8.8.8.8
1.1.1.1
4.23.54.222:9001
```

The input defaults to port 53 if no port is specified.

### Todo
- Colored console output for vanity
- Preview gifs
- Pull from public servers..
- Addition lookups beyond PTR?

___

###### Mirrors
[acid.vegas](https://git.acid.vegas/massrdns) • [GitHub](https://github.com/acidvegas/massrdns) • [GitLab](https://gitlab.com/acidvegas/massrdns) • [SuperNETs](https://git.supernets.org/acidvegas/massrdns)
