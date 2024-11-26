#!/bin/bash
cd /app && \
git clone https://github.com/wwsheng009/kun.git /app/kun && \
git clone https://github.com/wwsheng009/xun.git /app/xun && \
git clone https://github.com/wwsheng009/gou.git /app/gou && \
git clone https://github.com/wwsheng009/v8go.git /app/v8go && \
git clone https://github.com/wwsheng009/xgen.git /app/xgen-v1.0 && \
git clone https://github.com/wwsheng009/yao-init.git /app/yao-init && \
git clone https://github.com/wwsheng009/yao.git /app/yao

files=$(find /app/v8go -name "libv8*.zip")
for file in $files; do
    dir=$(dirname "$file")  # Get the directory where the ZIP file is located
    echo "Extracting $file to directory $dir"
    unzip -o -d $dir $file
    rm -rf $dir/__MACOSX
done

cd /app/yao && \
export VERSION=$(cat share/const.go  |grep 'const VERSION' | awk '{print $4}' | sed "s/\"//g") 

cd /app/yao && make tools && make artifacts-linux
mv /app/yao/dist/release/* /data/