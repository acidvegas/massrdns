# MASSRDNS

## Reverse DNS Lookup Tool
This Reverse DNS Lookup Tool is a sophisticated utility designed to perform reverse DNS lookups on large IP ranges efficiently. Built with concurrency in mind, it leverages multiple goroutines to expedite the process, making it highly scalable and performant. The tool utilizes a list of DNS servers, effectively load balancing the DNS queries across them. This not only distributes the request load but also provides redundancy; if one server fails or is slow, the tool can switch to another. Recognizing the real-world imperfections of network systems, the tool is intelligent enough to handle DNS server failures. After a certain threshold of consecutive failures, it automatically removes the faulty server from the list, ensuring that runtime is not bogged down by consistent non-performers. Furthermore, in the case of lookup failures due to network issues, the tool retries the lookup using different servers. This ensures that transient errors don't lead to missed lookups, enhancing the reliability of the results.

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
