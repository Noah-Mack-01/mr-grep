cd "$(dirname "${BASH_SOURCE[0]}")"
mkdir ./output
go run ./coordinator/mrcoordinator.go ./inputs/*.txt &
start_time=$(date +%s%3N)  # start time in milliseconds
echo "Run Started!"
sleep 1
echo "$1"
for i in 1 2 3 4 5 6 7 8
do
  go run ./worker/mrworker.go ./grep/grep.so "$1" &
done
echo "Workers Started!"
wait
rm -rf ./mr-*

end_time=$(date +%s%3N)  # # end time in milliseconds
duration_ms=$((end_time - start_time))  # duration in milliseconds
echo "Run finished in $duration_ms ms"