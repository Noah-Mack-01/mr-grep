cd "$(dirname "${BASH_SOURCE[0]}")"
cd ./coordinator && go build 
cd ../worker && go build  
cd ../grep && go build -buildmode=plugin ./grep.go
echo "Build done!"