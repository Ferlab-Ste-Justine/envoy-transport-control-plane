docker build -t server:server .
docker run -p 8081:8081 -p 8082:8082 -p 8083:8083 -p 8084:8084 -p 8085:8085 --rm server:server