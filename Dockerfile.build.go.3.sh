
# push到腾讯云
# https://console.cloud.tencent.com/tcr/repository/
# docker login ccr.ccs.tencentyun.com --username=
prod_image_name="my-markdown-blog-prod"

# 生成新的tag
docker tag ${prod_image_name}:latest ccr.ccs.tencentyun.com/gamestudio/images:${prod_image_name}
docker images | grep ${prod_image_name}

docker push ccr.ccs.tencentyun.com/gamestudio/images:${prod_image_name}

echo "finish"