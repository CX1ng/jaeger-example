## Jaeger Example

需要先启动Jaeger
> docker run -d -p 5775:5775/udp -p 16686:16686 jaegertracing/all-in-one:latest

访问[http://127.0.0.1:16686/](http://127.0.0.1:16686/)可以查看结果