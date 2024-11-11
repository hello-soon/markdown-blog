
prod_image_name="my-markdown-blog-prod"

docker ps -aqf "ancestor=${prod_image_name}" | xargs docker rm
docker ps -a  | grep ${prod_image_name}

echo "docker build  --tag ${prod_image_name} -f Dockerfile.build.go.2 ."
docker build  --tag ${prod_image_name} -f Dockerfile.build.go.2 .


export md_path="/Users/fullname/doc/work/gamestudio/bookmark/docs"

echo "docker run  -it -p 5006:5006  -v ${md_path}:/md --name my-test  ${prod_image_name}"
docker run  -it --rm -p 5006:5006  -v ${md_path}:/md --name my-test  ${prod_image_name}
docker exec -it my-test /data/app/markdown-blog

# scratch 体积更小
# gcr.io/distroless/base-debian11    72.3MB
# scratch                            40.3MB
