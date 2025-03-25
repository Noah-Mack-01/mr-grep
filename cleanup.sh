cd "$(dirname "${BASH_SOURCE[0]}")"

rm -f ./coordinator/coordinator
rm -f ./worker/worker
rm -f ./grep/grep.so
rm -f ./output/*

echo "Cleanup done!"
