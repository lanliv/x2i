#!/bin/bash

# Clean up old log if exists
rm -f x2i.log

# Start x2i.go by building and running it
echo "Building and starting x2i.go..."
go run ./x2i.go "C:\Users\stok\gatling-performance-kotlin\results" -i gatling -b gatling -u "cO5GJoY2gv2vF6n6iDIrkDWEircxNmShqvp0q1j7sKVJVvFPe5jx71R2u3sxLA0YaYMEIWcYqjlHn88AP5WQ8g==" > x2i.log 2>&1 &

X2I_PID=$!

# Now watch the log file for "Successfully written"
echo "Waiting for 'Successfully written' in x2i.log..."

tail -n0 -F x2i.log | while read line; do
  echo "$line" | grep "Successfully written" && {
    echo "Detected success message!"
    echo "Building and starting main.go..."
    go run ./main/main.go
    kill $X2I_PID
    exit 0
  }
done
