
current_time=$(date +"%Y-%m-%d %H:%M:%S")
echo "当前时间是: [$current_time]"

build_image_name="my-markdown-blog-build"

docker rm $(docker ps -aqf "ancestor=${build_image_name}")
docker ps -a  | grep ${build_image_name}

result=$(docker build --build-arg CURRENT_TIME="${current_time}"  --tag ${build_image_name} -f Dockerfile.build.go.1 .)

# 检查退出码
if [ "$result" -ne 0 ]; then
  echo "build Dockerfile.build.go.1 failed!"
  exit 1  # 构建失败，退出脚本，退出码为 1
else
  echo "build Dockerfile.build.go.1 succeeded!"
fi

export md_path="/Users/fullname/doc/work/gamestudio/bookmark/docs"
export custom_cmd="make run"
# export custom_cmd="make package"
# export custom_cmd="make build"

# -it 是否要进行交互
export it_flag="-it"
# export it_flag=""

# --rm 退出后，自动删除容器
echo "docker run --rm ${it_flag} -p 5006:5006 -v ./:/app  -v ${md_path}:/md  ${build_image_name} $custom_cmd "
docker run --rm ${it_flag} -p 5006:5006 -v ./:/app  -v ${md_path}:/md  ${build_image_name} $custom_cmd 

# docker run --rm  -it -p 5006:5006 -v ./:/app  my-markdown-blog /bin/bash 

# 启动 web服务. 运行之后访问地址 http://localhost:5006
# make run

# 编译, 在 bin 目录生成当前系统的压缩包，如：markdown-blog-v1.1.0-darwin-amd64.tar
# make

# 打包, 在 package 目录生成当前系统的压缩包，如：markdown-blog-v1.1.0-darwin-amd64.tar
# make package


# 生成 Windows、Linux、Mac 的压缩包
# 在 package 生成压缩包，如：markdown-blog-v1.1.0-darwin-amd64.tar markdown-blog-v1.1.0-linux-amd64.tar.gz markdown-blog-v1.1.0-windows-amd64.zip
# make package-all