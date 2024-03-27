FROM debian:12-slim
WORKDIR /app
ADD ./demodata /app/demodata
ADD ./gantt-backend-go /app

CMD ["/app/gantt-backend-go"]