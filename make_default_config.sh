mkdir  $(pwd)/docker_data/mysql
mkdir  $(pwd)/docker_data/elk
mkdir  $(pwd)/docker_data/elk/elasticsearch
mkdir  $(pwd)/docker_data/elk/logstash

docker network create snaplink_net

docker run -d \
  --name snaplink-rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=admin \
  -e RABBITMQ_DEFAULT_PASS=password.123 \
  --hostname snaplink-rabbitmq \
  --network snaplink_net \
  rabbitmq:latest

docker run -d \
  --name snaplink-mysql \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD="pjJdE#j:ovE-J:Vk1~?y" \
  -e MYSQL_DATABASE=saas \
  -e MYSQL_USER=saas \
  -e MYSQL_PASSWORD="vAm>~,A*Cdo9j#6@^*k3" \
  --network snaplink_net \
  mysql:8.0.17

docker run -d \
  --name snaplink-redis \
  -p 6379:6379 \
  -p 8001:8001 \
  --network snaplink_net \
  redis/redis-stack-server

docker run -d \
  --name snaplink-maxwell \
  --hostname snaplink-maxwell \
  --network snaplink_net \
  zendesk/maxwell \
  bin/maxwell --config /etc/maxwell/config.properties

docker run -d \
  --name snaplink-elasticsearch \
  --hostname snaplink-elasticsearch \
  -p 9200:9200 \
  -e "discovery.type=single-node" \
  --network snaplink_net \
  docker.elastic.co/elasticsearch/elasticsearch:8.7.0

docker run -d \
  --name snaplink-logstash \
  -p 5000:5000 \
  --network snaplink_net \
  docker.elastic.co/logstash/logstash:8.7.0

docker run -d \
  --name snaplink-kibana \
  -p 5601:5601 \
  --network snaplink_net \
  docker.elastic.co/kibana/kibana:8.7.0


# 对于RabbitMQ

# 对于MySQL
docker cp snaplink-mysql:/etc/my.cnf $(pwd)/docker_data/mysql/my.cnf
# 对于Maxwell
docker cp snaplink-maxwell:/etc/maxwell/ $(pwd)/docker_data/maxwell

# 对于 Elasticsearch
docker cp snaplink-elasticsearch:/usr/share/elasticsearch/config/elasticsearch.yml $(pwd)/docker_data/elk/elasticsearch.yml
docker cp snaplink-elasticsearch:/usr/share/elasticsearch/data $(pwd)/docker_data/elk/elasticsearch/data
# 对于 Logstash:
docker cp snaplink-logstash:/usr/share/logstash/config/logstash.yml $(pwd)/docker_data/elk/logstash.yml
docker cp snaplink-logstash:/usr/share/logstash/pipeline $(pwd)/docker_data/elk/logstash/pipeline
# 对于 Kibana:
docker cp snaplink-kibana:/usr/share/kibana/config/kibana.yml $(pwd)/docker_data/elk/kibana.yml

# 停止并删除RabbitMQ容器
docker stop snaplink-rabbitmq
docker rm snaplink-rabbitmq

# 停止并删除MySQL容器
docker stop snaplink-mysql
docker rm snaplink-mysql

# 停止并删除Maxwell容器
docker stop snaplink-maxwell
docker rm snaplink-maxwell

# 停止并删除Redis容器
docker stop snaplink-redis
docker rm snaplink-redis

# 停止并删除elasticsearch容器
docker stop snaplink-elasticsearch snaplink-logstash snaplink-kibana
docker rm snaplink-elasticsearch snaplink-logstash snaplink-kibana

# 删除网络
docker network rm snaplink_net