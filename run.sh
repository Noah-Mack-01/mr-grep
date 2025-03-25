cd "$(dirname "${BASH_SOURCE[0]}")"
go run ./coordinator/mrcoordinator.go ./inputs/*.txt &
echo "Run Started!"
echo "$1"
for i in 1 2 3 4 5
do
  go run ./worker/mrworker.go ./grep/grep.so "$1" &
done
echo "Workers Started!"
wait
rm -rf ./mr-*
echo "Run Done!"