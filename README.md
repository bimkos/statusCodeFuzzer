# statusCodeFuzzer
Simple status code fuzzer with path fuzz

Usage: main --file FILE [--suffix SUFFIX] [--proxy PROXY] [--threads THREADS]

Options:
  --file FILE, -f FILE   file with hosts
  --suffix SUFFIX, -s SUFFIX
                         file with suffix
  --proxy PROXY, -p PROXY
                         ex socks5://127.0.0.1:9050
  --threads THREADS, -t THREADS
                         threads count [default: 1]
  --help, -h             display this help and exit
