
current_time=$(date +"%Y-%m-%d %H:%M:%S")
echo "当前时间是: [$current_time]"

docker rm $(docker ps -aqf "ancestor=my-markdown-blog")
docker ps -a  | grep my-markdown-blog

docker build --build-arg CURRENT_TIME="${current_time}"  --tag my-markdown-blog -f Dockerfile.build.go.1 .


export md_path="/Users/fullname/doc/work/gamestudio/bookmark/docs"
export custom_cmd="/bin/bash"
export custom_cmd="make run"


# --rm 退出后，自动删除容器
docker run --rm  -it -p 5006:5006 -v ./:/app  -v ${md_path}:/md  my-markdown-blog $custom_cmd 
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