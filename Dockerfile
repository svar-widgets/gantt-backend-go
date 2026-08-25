FROM debian:12-slim
WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates
ADD ./demodata /app/demodata
ADD ./gantt-backend-go /app

CMD ["/app/gantt-backend-go"]